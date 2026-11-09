package main

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"tracker/internal/db"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

// httpError logs the error (with method, path, and status) via the log
// package, then sends publicMsg to the client. err may be nil for handler
// paths that reject the request without an underlying error (e.g. bad method).
func httpError(w http.ResponseWriter, r *http.Request, status int, publicMsg string, err error) {
	if err != nil {
		log.Printf("%s %s -> %d %s: %v", r.Method, r.URL.Path, status, publicMsg, err)
	} else {
		log.Printf("%s %s -> %d %s", r.Method, r.URL.Path, status, publicMsg)
	}
	http.Error(w, publicMsg, status)
}

// listHandler builds a GET handler that fetches rows and maps each to a DTO.
func listHandler[T, R any](list func(ctx context.Context) ([]T, error), toDTO func(T) R) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := list(r.Context())
		if err != nil {
			httpError(w, r, 500, err.Error(), err)
			return
		}
		out := make([]R, len(rows))
		for i, row := range rows {
			out[i] = toDTO(row)
		}
		writeJSON(w, out)
	}
}

type actionDTO struct {
	Name        string `json:"name"`
	Tag         string `json:"tag"`
	Description string `json:"description"`
}

func actionToDTO(a db.Action) actionDTO {
	return actionDTO{Name: a.Name, Tag: a.Tag, Description: a.Description}
}

func handleAPIActions(w http.ResponseWriter, r *http.Request) {
	listHandler(q.ListActions, actionToDTO)(w, r)
}

type skillAreaDTO struct {
	Skill string `json:"skill"`
	Areas string `json:"areas"`
}

func skillAreaToDTO(s db.SkillArea) skillAreaDTO {
	return skillAreaDTO{Skill: s.Skill, Areas: s.Areas}
}

func handleAPISkillAreas(w http.ResponseWriter, r *http.Request) {
	listHandler(q.ListSkillAreas, skillAreaToDTO)(w, r)
}

type conditionDTO struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	Effects          string `json:"effects"`
	DescriptionAfter string `json:"descriptionAfter"`
}

func conditionToDTO(c db.Condition) conditionDTO {
	return conditionDTO{
		Name:             c.Name,
		Description:      c.Description.String,
		Effects:          c.Effects.String,
		DescriptionAfter: c.DescriptionAfter.String,
	}
}

func handleAPIConditions(w http.ResponseWriter, r *http.Request) {
	listHandler(q.ListConditions, conditionToDTO)(w, r)
}

type exhaustionLevelDTO struct {
	Level  string `json:"level"`
	Effect string `json:"effect"`
}

func exhaustionLevelToDTO(e db.ExhaustionLevel) exhaustionLevelDTO {
	return exhaustionLevelDTO{Level: e.Level, Effect: e.Effect}
}

func handleAPIExhaustionLevels(w http.ResponseWriter, r *http.Request) {
	listHandler(q.ListExhaustionLevels, exhaustionLevelToDTO)(w, r)
}

type monsterDTO struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Cr       string `json:"cr"`
	ImageUrl string `json:"imageUrl"`
	Source   string `json:"source"`
}

func monsterToDTO(m db.ListMonstersRow) monsterDTO {
	return monsterDTO{
		ID:       m.ID,
		Name:     m.Name,
		Type:     m.Type.String,
		Cr:       m.Cr.String,
		ImageUrl: m.ImageUrl,
		Source:   m.Source,
	}
}

func handleAPIMonsters(w http.ResponseWriter, r *http.Request) {
	listHandler(q.ListMonsters, monsterToDTO)(w, r)
}

// monsterStatDTO is the lightweight ac/hp/dex lookup used to autofill the
// initiative tracker, kept separate from the full bestiary list/detail DTOs.
type monsterStatDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Ac   int64  `json:"ac"`
	Hp   int64  `json:"hp"`
	Dex  int64  `json:"dex"`
}

func monsterStatToDTO(m db.ListMonsterStatsRow) monsterStatDTO {
	return monsterStatDTO{
		ID:   m.ID,
		Name: m.Name,
		Ac:   m.Ac.Int64,
		Hp:   m.Hp.Int64,
		Dex:  m.Dex.Int64,
	}
}

func handleAPIMonsterStats(w http.ResponseWriter, r *http.Request) {
	listHandler(q.ListMonsterStats, monsterStatToDTO)(w, r)
}

// statBlockEntry is one named trait/action/reaction/legendary-action block.
type statBlockEntry struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func parseStatBlockEntries(jsonText string) []statBlockEntry {
	if jsonText == "" {
		return nil
	}
	var entries []statBlockEntry
	if err := json.Unmarshal([]byte(jsonText), &entries); err != nil {
		return nil
	}
	return entries
}

type monsterDetailDTO struct {
	ID                  int64            `json:"id"`
	Name                string           `json:"name"`
	Type                string           `json:"type"`
	Size                string           `json:"size"`
	Alignment           string           `json:"alignment"`
	Cr                  string           `json:"cr"`
	Source              string           `json:"source"`
	Hp                  int64            `json:"hp"`
	HpFormula           string           `json:"hpFormula"`
	Ac                  int64            `json:"ac"`
	AcDesc              string           `json:"acDesc"`
	Speed               string           `json:"speed"`
	Str                 int64            `json:"str"`
	Dex                 int64            `json:"dex"`
	Con                 int64            `json:"con"`
	Int                 int64            `json:"int"`
	Wis                 int64            `json:"wis"`
	Cha                 int64            `json:"cha"`
	SavingThrows        string           `json:"savingThrows"`
	Skills              string           `json:"skills"`
	DamageResistances   string           `json:"damageResistances"`
	DamageImmunities    string           `json:"damageImmunities"`
	Vulnerabilities     string           `json:"vulnerabilities"`
	ConditionImmunities string           `json:"conditionImmunities"`
	Senses              string           `json:"senses"`
	PassivePerception   int64            `json:"passivePerception"`
	Languages           string           `json:"languages"`
	Environment         string           `json:"environment"`
	ImageUrl            string           `json:"imageUrl"`
	TokenUrl            string           `json:"tokenUrl"`
	Traits              []statBlockEntry `json:"traits"`
	Actions             []statBlockEntry `json:"actions"`
	Reactions           []statBlockEntry `json:"reactions"`
	LegendaryActions    []statBlockEntry `json:"legendaryActions"`
	BonusActions        []statBlockEntry `json:"bonusActions"`
	Spellcasting        []statBlockEntry `json:"spellcasting"`
	LairActions         []statBlockEntry `json:"lairActions"`
	RegionalEffects     []statBlockEntry `json:"regionalEffects"`
	Notes               string           `json:"notes"`
}

func monsterDetailToDTO(m db.GetMonsterRow) monsterDetailDTO {
	return monsterDetailDTO{
		ID:                  m.ID,
		Name:                m.Name,
		Type:                m.Type.String,
		Size:                m.Size,
		Alignment:           m.Alignment,
		Cr:                  m.Cr.String,
		Source:              m.Source,
		Hp:                  m.Hp.Int64,
		HpFormula:           m.HpFormula.String,
		Ac:                  m.Ac.Int64,
		AcDesc:              m.AcDesc.String,
		Speed:               m.Speed.String,
		Str:                 m.Str.Int64,
		Dex:                 m.Dex.Int64,
		Con:                 m.Con.Int64,
		Int:                 m.IntScore.Int64,
		Wis:                 m.Wis.Int64,
		Cha:                 m.Cha.Int64,
		SavingThrows:        m.SavingThrows,
		Skills:              m.Skills,
		DamageResistances:   m.DamageResistances,
		DamageImmunities:    m.DamageImmunities,
		Vulnerabilities:     m.Vulnerabilities,
		ConditionImmunities: m.ConditionImmunities,
		Senses:              m.Senses,
		PassivePerception:   m.PassivePerception.Int64,
		Languages:           m.Languages,
		Environment:         m.Environment,
		ImageUrl:            m.ImageUrl,
		TokenUrl:            m.TokenUrl,
		Traits:              parseStatBlockEntries(m.Traits),
		Actions:             parseStatBlockEntries(m.Actions),
		Reactions:           parseStatBlockEntries(m.Reactions),
		LegendaryActions:    parseStatBlockEntries(m.LegendaryActions),
		BonusActions:        parseStatBlockEntries(m.BonusActions),
		Spellcasting:        parseStatBlockEntries(m.Spellcasting),
		LairActions:         parseStatBlockEntries(m.LairActions),
		RegionalEffects:     parseStatBlockEntries(m.RegionalEffects),
		Notes:               m.Notes,
	}
}

func handleAPIMonster(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpError(w, r, 400, "invalid monster id", err)
		return
	}
	m, err := q.GetMonster(r.Context(), id)
	if err != nil {
		httpError(w, r, 404, err.Error(), err)
		return
	}
	writeJSON(w, monsterDetailToDTO(m))
}

type spellDTO struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Level   int64  `json:"level"`
	School  string `json:"school"`
	Classes string `json:"classes"`
	Source  string `json:"source"`
}

func spellToDTO(s db.ListSpellsRow) spellDTO {
	return spellDTO{
		ID:      s.ID,
		Name:    s.Name,
		Level:   s.Level,
		School:  s.School.String,
		Classes: s.Classes,
		Source:  s.Source,
	}
}

func handleAPISpells(w http.ResponseWriter, r *http.Request) {
	listHandler(q.ListSpells, spellToDTO)(w, r)
}

type spellDetailDTO struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Level         int64  `json:"level"`
	School        string `json:"school"`
	Ritual        bool   `json:"ritual"`
	CastingTime   string `json:"castingTime"`
	Range         string `json:"range"`
	Components    string `json:"components"`
	Duration      string `json:"duration"`
	Concentration bool   `json:"concentration"`
	Classes       string `json:"classes"`
	Description   string `json:"description"`
	HigherLevel   string `json:"higherLevel"`
	Source        string `json:"source"`
}

func spellDetailToDTO(s db.GetSpellRow) spellDetailDTO {
	return spellDetailDTO{
		ID:            s.ID,
		Name:          s.Name,
		Level:         s.Level,
		School:        s.School,
		Ritual:        s.Ritual,
		CastingTime:   s.CastingTime,
		Range:         s.Range,
		Components:    s.Components,
		Duration:      s.Duration,
		Concentration: s.Concentration,
		Classes:       s.Classes,
		Description:   s.Description,
		HigherLevel:   s.HigherLevel,
		Source:        s.Source,
	}
}

func handleAPISpell(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpError(w, r, 400, "invalid spell id", err)
		return
	}
	s, err := q.GetSpell(r.Context(), id)
	if err != nil {
		httpError(w, r, 404, err.Error(), err)
		return
	}
	writeJSON(w, spellDetailToDTO(s))
}

func handleAPIChat(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.FormValue("q")
	}
	if q == "" {
		httpError(w, r, 400, "missing q", nil)
		return
	}
	model := r.URL.Query().Get("model")
	if model == "" {
		model = r.FormValue("model")
	}
	if model == "" {
		httpError(w, r, 400, "missing model", nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 240*time.Second)
	defer cancel()
	answer, err := runAgent(ctx, model, q)
	if err != nil {
		httpError(w, r, 500, err.Error(), err)
		return
	}
	writeJSON(w, struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}{q, answer})
}

func handleAPIModels(w http.ResponseWriter, r *http.Request) {
	models, err := listChatModels(r.Context())
	if err != nil {
		httpError(w, r, 502, err.Error(), err)
		return
	}
	writeJSON(w, models)
}

var reFTSTerm = regexp.MustCompile(`[A-Za-z0-9']+`)

// ftsMatchQuery turns free text into an FTS5 MATCH expression. Each term is
// prefix-matched and OR'd together so partial/misspelled words still surface
// results, the same tolerant-matching behavior search engines use.
func ftsMatchQuery(query string) string {
	terms := reFTSTerm.FindAllString(query, -1)
	if len(terms) == 0 {
		return ""
	}
	quoted := make([]string, len(terms))
	for i, t := range terms {
		quoted[i] = `"` + strings.ReplaceAll(t, `"`, `""`) + `"*`
	}
	return strings.Join(quoted, " OR ")
}

// rrfConst is the standard reciprocal-rank-fusion constant used by hybrid
// search systems (e.g. Elasticsearch, Azure AI Search) to blend independently
// ranked result lists - here, BM25 keyword relevance and embedding similarity.
const rrfConst = 60

func searchLore(ctx context.Context, query string) ([]searchResult, error) {
	type chunk struct {
		chapterTitle string
		content      string
	}

	// No query: return every block in document (id) order, unscored, so the
	// UI can page through the full lore. Score -1 signals "not a match".
	if query == "" {
		rows, err := conn.QueryContext(ctx,
			`SELECT chapter_title, content FROM chapter_chunks ORDER BY id`)
		if err != nil {
			return nil, fmt.Errorf("search error: %w", err)
		}
		defer rows.Close()
		results := []searchResult{}
		for rows.Next() {
			var c chunk
			if err := rows.Scan(&c.chapterTitle, &c.content); err != nil {
				return nil, err
			}
			results = append(results, searchResult{ChapterTitle: c.chapterTitle, Content: c.content, Score: -1})
		}
		return results, nil
	}

	fused := map[int]float64{}
	chunks := map[int]chunk{}

	// Keyword relevance: SQLite FTS5 BM25 ranking.
	if matchQ := ftsMatchQuery(query); matchQ != "" {
		rows, err := conn.QueryContext(ctx,
			`SELECT rowid, chapter_title, content FROM chapter_chunks_fts
			 WHERE chapter_chunks_fts MATCH ? ORDER BY bm25(chapter_chunks_fts) LIMIT 25`, matchQ)
		if err != nil {
			return nil, fmt.Errorf("search error: %w", err)
		}
		rank := 0
		for rows.Next() {
			var id int
			var c chunk
			if err := rows.Scan(&id, &c.chapterTitle, &c.content); err != nil {
				rows.Close()
				return nil, err
			}
			rank++
			fused[id] += 1.0 / float64(rrfConst+rank)
			chunks[id] = c
		}
		rows.Close()
	}

	// Semantic relevance: embedding cosine similarity.
	if emb, err := queryEmbedding(ctx, query); err == nil {
		rows, err := conn.QueryContext(ctx, `SELECT id, chapter_title, content, embedding FROM chapter_chunks`)
		if err != nil {
			return nil, fmt.Errorf("search error: %w", err)
		}
		type scored struct {
			id    int
			chunk chunk
			score float64
		}
		var semantic []scored
		for rows.Next() {
			var id int
			var c chunk
			var embeddingJSON string
			if err := rows.Scan(&id, &c.chapterTitle, &c.content, &embeddingJSON); err != nil {
				rows.Close()
				return nil, err
			}
			score, err := cosineSimilarity(emb, embeddingJSON)
			if err != nil {
				continue
			}
			semantic = append(semantic, scored{id, c, score})
		}
		rows.Close()
		slices.SortFunc(semantic, func(a, b scored) int { return cmp.Compare(b.score, a.score) })
		semantic = semantic[:min(len(semantic), 25)]
		for rank, s := range semantic {
			fused[s.id] += 1.0 / float64(rrfConst+rank+1)
			chunks[s.id] = s.chunk
		}
	}

	type idScore struct {
		id    int
		score float64
	}
	ranked := make([]idScore, 0, len(fused))
	for id, score := range fused {
		ranked = append(ranked, idScore{id, score})
	}
	slices.SortFunc(ranked, func(a, b idScore) int { return cmp.Compare(b.score, a.score) })
	ranked = ranked[:min(len(ranked), 5)]

	results := []searchResult{}
	maxScore := 0.0
	if len(ranked) > 0 {
		maxScore = ranked[0].score
	}
	for _, rs := range ranked {
		c := chunks[rs.id]
		norm := 0.0
		if maxScore > 0 {
			norm = rs.score / maxScore
		}
		results = append(results, searchResult{ChapterTitle: c.chapterTitle, Content: c.content, Score: norm})
	}

	return results, nil
}

func handleAPISearch(w http.ResponseWriter, r *http.Request) {
	results, err := searchLore(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpError(w, r, 500, err.Error(), err)
		return
	}
	writeJSON(w, results)
}

type initiativePresetDTO struct {
	Name       string          `json:"name"`
	Combatants json.RawMessage `json:"combatants"`
}

func initiativePresetToDTO(p db.InitiativePreset) initiativePresetDTO {
	return initiativePresetDTO{Name: p.Name, Combatants: json.RawMessage(p.Combatants)}
}

func handleAPIInitiativePresets(w http.ResponseWriter, r *http.Request) {
	listHandler(q.ListInitiativePresets, initiativePresetToDTO)(w, r)
}

var validPresetName = regexp.MustCompile(`^.{1,100}$`)

// presetName returns the validated {name} path wildcard, or "" after writing
// a 400 response. ServeMux already percent-decodes the segment.
func presetName(w http.ResponseWriter, r *http.Request) string {
	name := r.PathValue("name")
	if !validPresetName.MatchString(name) {
		httpError(w, r, 400, "invalid preset name", nil)
		return ""
	}
	return name
}

func handleAPIPutInitiativePreset(w http.ResponseWriter, r *http.Request) {
	name := presetName(w, r)
	if name == "" {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, r, 400, err.Error(), err)
		return
	}
	if !json.Valid(body) {
		httpError(w, r, 400, "invalid json", nil)
		return
	}
	if err := q.UpsertInitiativePreset(r.Context(), db.UpsertInitiativePresetParams{Name: name, Combatants: string(body)}); err != nil {
		httpError(w, r, 500, err.Error(), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleAPIDeleteInitiativePreset(w http.ResponseWriter, r *http.Request) {
	name := presetName(w, r)
	if name == "" {
		return
	}
	if err := q.DeleteInitiativePreset(r.Context(), name); err != nil {
		httpError(w, r, 500, err.Error(), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var validNoteName = regexp.MustCompile(`^[A-Za-z0-9_-]+\.md$`)

const notesDir = "notes"

func syncNotesFromDisk(ctx context.Context) error {
	files, err := filepath.Glob(filepath.Join(notesDir, "*.md"))
	if err != nil {
		return err
	}
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		name := filepath.Base(f)
		if err := q.UpsertNote(ctx, db.UpsertNoteParams{Name: name, Content: string(content)}); err != nil {
			return err
		}
	}
	return nil
}

func handleAPINotesList(w http.ResponseWriter, r *http.Request) {
	files, err := filepath.Glob(filepath.Join(notesDir, "*.md"))
	if err != nil {
		httpError(w, r, 500, err.Error(), err)
		return
	}
	names := make([]string, 0, len(files))
	for _, file := range files {
		names = append(names, filepath.Base(file))
	}
	writeJSON(w, names)
}

// noteName returns the validated {name} path wildcard, or "" after writing a
// 400 response. The regexp also blocks path traversal in the file write.
func noteName(w http.ResponseWriter, r *http.Request) string {
	name := r.PathValue("name")
	if !validNoteName.MatchString(name) {
		httpError(w, r, 400, "invalid note name", nil)
		return ""
	}
	return name
}

func handleAPIGetNote(w http.ResponseWriter, r *http.Request) {
	name := noteName(w, r)
	if name == "" {
		return
	}
	note, err := q.GetNote(r.Context(), name)
	if err != nil {
		httpError(w, r, 404, err.Error(), err)
		return
	}
	writeJSON(w, struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}{note.Name, note.Content})
}

func handleAPIPutNote(w http.ResponseWriter, r *http.Request) {
	name := noteName(w, r)
	if name == "" {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, r, 400, err.Error(), err)
		return
	}
	if err := q.UpsertNote(r.Context(), db.UpsertNoteParams{Name: name, Content: string(body)}); err != nil {
		httpError(w, r, 500, err.Error(), err)
		return
	}
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		httpError(w, r, 500, err.Error(), err)
		return
	}
	if err := os.WriteFile(filepath.Join(notesDir, name), body, 0644); err != nil {
		httpError(w, r, 500, err.Error(), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
