package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"tracker/internal/embeddings"
	"tracker/internal/llm"
)

// Local CPU inference needs ~1 min to cold-load and prefill system prompt;
// tool calls (web/API) should fail fast instead of hanging forever.
var llmClient = &http.Client{Timeout: 180 * time.Second}
var toolClient = &http.Client{Timeout: 15 * time.Second}

const systemPrompt = `You are a D&D Dungeon Master assistant.
Tools: sql (campaign database, try first), lore_search (adventure story text), dnd_lookup (SRD rules/items/classes), web_search (last resort).

SQLite schema (read-only):
  Monsters(name, type, cr, hp, ac, speed, size, alignment, str, dex, con, int_score, wis, cha,
    saving_throws, skills, passive_perception, damage_resistances, damage_immunities, vulnerabilities,
    condition_immunities, senses, languages, environment, source, traits, actions, bonus_actions,
    reactions, legendary_actions, lair_actions, spellcasting)
  Spells(name, level, school, ritual, casting_time, range, components, duration, concentration, description, higher_level, source)
  Notes(name, content)  -- DM markdown notes
  MapMarkers(map_id, title, body)
  InitiativePresets(name, combatants)  -- JSON [{name, hp, ac}]
  Conditions(name, effects, description)
  ExhaustionLevels(level, effect)
  Actions(name, tag, description)
  SkillAreas(skill, areas)

Formats:
- cr TEXT '0','1/8','1/4','1/2','1'..'30'. Numeric: (CASE cr WHEN '1/8' THEN 0.125 WHEN '1/4' THEN 0.25 WHEN '1/2' THEN 0.5 ELSE CAST(cr AS REAL) END)
- type lowercase with subtype, e.g. 'humanoid (goblinoid)': use LIKE 'humanoid%'.
- size T/S/M/L/H/G. alignment codes 'L E', 'C G', 'N', 'U'.
- saving_throws/skills 'Dex +6, Wis +8'. immunities 'fire; poison'. speed 'walk 40 ft., fly 80 ft.'.
- traits/actions/etc: JSON [{name, text}], search with LIKE.
- Spells.level 0 = cantrip. ritual/concentration 0/1. school 'Evocation'.
- Fuzzy names: name LIKE '%goblin%'.
Select needed columns only, add LIMIT. On error or no rows, fix/loosen query and retry.

Examples:
SELECT name, cr, ac, hp, speed, damage_immunities FROM Monsters WHERE name LIKE '%adult red dragon%' LIMIT 5;
SELECT name, cr, hp FROM Monsters WHERE type LIKE 'undead%' AND (CASE cr WHEN '1/8' THEN 0.125 WHEN '1/4' THEN 0.25 WHEN '1/2' THEN 0.5 ELSE CAST(cr AS REAL) END) BETWEEN 3 AND 5 LIMIT 50;
SELECT name, range, duration FROM Spells WHERE level = 3 AND concentration = 1 AND school = 'Evocation' LIMIT 50;

Strict rules:
    No articles (the, a, an).
    No pleasantries (certainly, sure, hello).
    No pronouns (I, you, me).
    Short sentences (3-6 words).
    Prefer code/data over talk.
    Absolute bluntness.
    No Emojis.
    Always use a tool before answering. Never invent stats.

Response style:
"Adult red dragon. CR 17. Legendary. See stats below." NOT "An adult red dragon is a very powerful creature with a challenge rating of 17 and several legendary actions you should be aware of."`

type chatMsg struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type toolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type toolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type chatTool struct {
	Type     string       `json:"type"`
	Function toolFunction `json:"function"`
}

type chatReq struct {
	Model    string     `json:"model"`
	Messages []chatMsg  `json:"messages"`
	Tools    []chatTool `json:"tools,omitempty"`
	Stream   bool       `json:"stream"`
}

type chatResp struct {
	Choices []struct {
		Message      chatMsg `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

var tools = []chatTool{
	{Type: "function", Function: toolFunction{
		Name:        "sql",
		Description: "Run read-only SQLite SELECT query against campaign database. Schema and column formats in system prompt.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "SQL SELECT statement"},
			},
			"required": []string{"query"},
		},
	}},
	{Type: "function", Function: toolFunction{
		Name:        "web_search",
		Description: "Search the web for general information not in the campaign database.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Search query"},
			},
			"required": []string{"query"},
		},
	}},
	{Type: "function", Function: toolFunction{
		Name:        "lore_search",
		Description: "Search campaign lore (chapter text, hybrid keyword + semantic search) for narrative/story content.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Search query"},
			},
			"required": []string{"query"},
		},
	}},
	{Type: "function", Function: toolFunction{
		Name:        "dnd_lookup",
		Description: "Look up D&D 5e rules data (monsters, spells, equipment, classes, etc.) from the official 5e API.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"category": map[string]any{
					"type":        "string",
					"description": "Resource category, e.g. monsters, spells, equipment, magic-items, classes, races, conditions, damage-types",
				},
				"index": map[string]any{
					"type":        "string",
					"description": "Specific resource slug (e.g. 'aboleth', 'fireball'). Omit to list all in category.",
				},
			},
			"required": []string{"category"},
		},
	}},
}

// listChatModels returns chat-capable models on the LLM server, so callers can
// offer a picker instead of a hardcoded model.
func listChatModels(ctx context.Context) ([]string, error) {
	return llm.ChatModels(ctx, toolClient)
}

func chat(ctx context.Context, model string, messages []chatMsg) (*chatResp, error) {
	var cr chatResp
	err := llm.PostJSON(ctx, llmClient, "/chat/completions", chatReq{
		Model:    model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}, &cr)
	if err != nil {
		return nil, fmt.Errorf("chat request: %w", err)
	}
	return &cr, nil
}

func runAgent(ctx context.Context, model, question string) (string, error) {
	messages := []chatMsg{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: question},
	}

	for range 8 {
		resp, err := chat(ctx, model, messages)
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "No response from model.", nil
		}

		msg := resp.Choices[0].Message
		messages = append(messages, msg)

		if resp.Choices[0].FinishReason != "tool_calls" || len(msg.ToolCalls) == 0 {
			return msg.Content, nil
		}

		for _, tc := range msg.ToolCalls {
			result := executeTool(ctx, tc.Function.Name, tc.Function.Arguments)
			messages = append(messages, chatMsg{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result,
			})
		}
	}
	return "Max tool iterations reached.", nil
}

func executeTool(ctx context.Context, name, argsJSON string) string {
	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "Error parsing args: " + err.Error()
	}
	switch name {
	case "sql":
		query, _ := args["query"].(string)
		return execSQL(ctx, query)
	case "web_search":
		query, _ := args["query"].(string)
		return webSearch(ctx, query)
	case "lore_search":
		query, _ := args["query"].(string)
		return loreSearchTool(ctx, query)
	case "dnd_lookup":
		category, _ := args["category"].(string)
		index, _ := args["index"].(string)
		return dndLookup(ctx, category, index)
	}
	return "Unknown tool: " + name
}

func loreSearchTool(ctx context.Context, query string) string {
	results, err := searchLore(ctx, query)
	if err != nil {
		return "Search error: " + err.Error()
	}
	if len(results) == 0 {
		return "No results found."
	}
	var sb strings.Builder
	for _, r := range results {
		fmt.Fprintf(&sb, "- %s (score %.2f)\n  %s\n\n", r.ChapterTitle, r.Score, r.Content)
	}
	return sb.String()
}

var reResult = regexp.MustCompile(`(?s)class="result__title"[^>]*>.*?<a[^>]*href="([^"]*)"[^>]*>(.*?)</a>.*?class="result__snippet"[^>]*>(.*?)</span>`)
var reTag = regexp.MustCompile(`<[^>]+>`)

func webSearch(ctx context.Context, query string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://html.duckduckgo.com/html/?q="+url.QueryEscape(query), nil)
	if err != nil {
		return "Error: " + err.Error()
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := toolClient.Do(req)
	if err != nil {
		return "Error: " + err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Sprintf("Error: web search returned %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Error reading web search response: " + err.Error()
	}

	matches := reResult.FindAllSubmatch(body, 5)
	if len(matches) == 0 {
		return "No results found."
	}
	var sb strings.Builder
	for _, m := range matches {
		title := reTag.ReplaceAllString(string(m[2]), "")
		snippet := reTag.ReplaceAllString(string(m[3]), "")
		link := string(m[1])
		fmt.Fprintf(&sb, "- %s\n  %s\n  %s\n\n", strings.TrimSpace(title), strings.TrimSpace(snippet), link)
	}
	return sb.String()
}

const dndAPIBase = "https://www.dnd5eapi.co/api/2014"

func dndLookup(ctx context.Context, category, index string) string {
	reqURL := dndAPIBase + "/" + url.PathEscape(category)
	if index != "" {
		reqURL += "/" + url.PathEscape(index)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "Error building request: " + err.Error()
	}
	resp, err := toolClient.Do(req)
	if err != nil {
		return "Error calling D&D API: " + err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "Error: D&D API returned " + resp.Status
	}
	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "Error decoding response: " + err.Error()
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "Error encoding response: " + err.Error()
	}
	s := string(out)
	if len(s) > 4000 {
		s = s[:4000] + "\n...(truncated)"
	}
	return s
}

func execSQL(ctx context.Context, query string) string {
	q := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(q, "SELECT") && !strings.HasPrefix(q, "WITH") {
		return "Error: only SELECT or WITH queries allowed."
	}
	rows, err := roConn.QueryContext(ctx, query)
	if err != nil {
		return "Query error: " + err.Error() + "\nFix query against schema in system prompt and retry."
	}
	defer rows.Close()

	colNames, err := rows.Columns()
	if err != nil {
		return "Query error: " + err.Error()
	}

	var sb strings.Builder
	sb.WriteString(strings.Join(colNames, " | "))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", 60))
	sb.WriteString("\n")

	count := 0
	for rows.Next() && count < 50 {
		vals := make([]any, len(colNames))
		ptrs := make([]any, len(colNames))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return "Query error: " + err.Error()
		}
		var strs []string
		for _, v := range vals {
			strs = append(strs, fmt.Sprintf("%v", v))
		}
		sb.WriteString(strings.Join(strs, " | "))
		sb.WriteString("\n")
		count++
	}
	if count == 0 {
		return "No rows returned."
	}
	if count == 50 {
		sb.WriteString("(truncated at 50 rows)\n")
	}
	return sb.String()
}

func cosineSimilarity(aJSON, bJSON string) (float64, error) {
	a, err := parseEmbedding(aJSON)
	if err != nil {
		return 0, err
	}
	b, err := parseEmbedding(bJSON)
	if err != nil {
		return 0, err
	}
	if len(a) != len(b) || len(a) == 0 {
		return 0, fmt.Errorf("embedding length mismatch")
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0, nil
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}

func parseEmbedding(s string) ([]float64, error) {
	var v []float64
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil, err
	}
	return v, nil
}

func queryEmbedding(ctx context.Context, text string) (string, error) {
	return embeddings.Embed(ctx, text)
}

type searchResult struct {
	ChapterTitle string  `json:"chapterTitle"`
	Content      string  `json:"content"`
	Score        float64 `json:"score"`
}
