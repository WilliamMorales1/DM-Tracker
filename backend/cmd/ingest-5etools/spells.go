// spells.go ingests the 2014 D&D spell list from the 5etools data mirror
// into the Spells table.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"tracker/internal/db"
)

const spellDataBaseURL = "https://raw.githubusercontent.com/5etools-mirror-3/5etools-src/main/data/spells/"

// spellSourceExclude lists sourcebooks that are not part of the 2014 ruleset
// (the 2024 "One D&D" revision reprints spells under the XPHB source).
var spellSourceExclude = map[string]bool{
	"XPHB": true,
}

type spellRangeJSON struct {
	Type     string `json:"type"`
	Distance *struct {
		Type   string `json:"type"`
		Amount int    `json:"amount"`
	} `json:"distance"`
}

type spellComponentsJSON struct {
	V bool            `json:"v"`
	S bool            `json:"s"`
	M json.RawMessage `json:"m"`
}

type spellTimeJSON struct {
	Number    int    `json:"number"`
	Unit      string `json:"unit"`
	Condition string `json:"condition"`
}

type spellDurationJSON struct {
	Type     string `json:"type"`
	Duration *struct {
		Type   string `json:"type"`
		Amount int    `json:"amount"`
	} `json:"duration"`
	Concentration bool `json:"concentration"`
}

type spellClassesJSON struct {
	FromClassList []struct {
		Name string `json:"name"`
	} `json:"fromClassList"`
}

type spellJSON struct {
	Name               string              `json:"name"`
	Source             string              `json:"source"`
	Level              int                 `json:"level"`
	School             string              `json:"school"`
	Time               []spellTimeJSON     `json:"time"`
	Range              spellRangeJSON      `json:"range"`
	Components         spellComponentsJSON `json:"components"`
	Duration           []spellDurationJSON `json:"duration"`
	Entries            []any               `json:"entries"`
	EntriesHigherLevel []entryBlock        `json:"entriesHigherLevel"`
	Classes            spellClassesJSON    `json:"classes"`
	Meta               struct {
		Ritual bool `json:"ritual"`
	} `json:"meta"`
}

type spellFile struct {
	Spell []spellJSON `json:"spell"`
}

var schoolNames = map[string]string{
	"A": "Abjuration",
	"C": "Conjuration",
	"D": "Divination",
	"E": "Enchantment",
	"V": "Evocation",
	"I": "Illusion",
	"N": "Necromancy",
	"T": "Transmutation",
}

func schoolName(code string) string {
	if n, ok := schoolNames[code]; ok {
		return n
	}
	return code
}

func spellTimeToString(t []spellTimeJSON) string {
	var parts []string
	for _, e := range t {
		s := fmt.Sprintf("%d %s", e.Number, e.Unit)
		if e.Number != 1 {
			s += "s"
		}
		if e.Condition != "" {
			s += ", " + cleanText(e.Condition)
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " or ")
}

// capitalize uppercases the first rune of s. Used instead of the deprecated
// strings.Title for the single-word range/distance-type tokens below.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

func spellRangeToString(r spellRangeJSON) string {
	switch r.Type {
	case "point":
		if r.Distance == nil {
			return ""
		}
		if r.Distance.Type == "self" || r.Distance.Type == "touch" || r.Distance.Type == "unlimited" || r.Distance.Type == "sight" {
			return capitalize(r.Distance.Type)
		}
		return fmt.Sprintf("%d %s", r.Distance.Amount, r.Distance.Type)
	case "radius", "sphere", "cone", "line", "cube", "hemisphere", "cylinder":
		if r.Distance == nil {
			return capitalize(r.Type)
		}
		return fmt.Sprintf("Self (%d-%s %s)", r.Distance.Amount, r.Distance.Type, r.Type)
	case "special":
		return "Special"
	default:
		return capitalize(r.Type)
	}
}

func spellComponentsToString(c spellComponentsJSON) string {
	var parts []string
	if c.V {
		parts = append(parts, "V")
	}
	if c.S {
		parts = append(parts, "S")
	}
	if len(c.M) > 0 {
		var s string
		if err := json.Unmarshal(c.M, &s); err == nil {
			parts = append(parts, fmt.Sprintf("M (%s)", cleanText(s)))
		} else {
			var obj struct {
				Text string `json:"text"`
			}
			if json.Unmarshal(c.M, &obj) == nil && obj.Text != "" {
				parts = append(parts, fmt.Sprintf("M (%s)", cleanText(obj.Text)))
			} else {
				parts = append(parts, "M")
			}
		}
	}
	return strings.Join(parts, ", ")
}

func spellDurationToString(d []spellDurationJSON) (string, bool) {
	var parts []string
	conc := false
	for _, e := range d {
		if e.Concentration {
			conc = true
		}
		switch e.Type {
		case "instant":
			parts = append(parts, "Instantaneous")
		case "permanent":
			parts = append(parts, "Until dispelled")
		case "special":
			parts = append(parts, "Special")
		case "timed":
			if e.Duration == nil {
				continue
			}
			s := fmt.Sprintf("%d %s", e.Duration.Amount, e.Duration.Type)
			if e.Duration.Amount != 1 {
				s += "s"
			}
			if e.Concentration {
				s = "Concentration, up to " + s
			}
			parts = append(parts, s)
		default:
			parts = append(parts, capitalize(e.Type))
		}
	}
	return strings.Join(parts, " or "), conc
}

func spellClassesToString(c spellClassesJSON) string {
	var names []string
	for _, cl := range c.FromClassList {
		names = append(names, cl.Name)
	}
	return strings.Join(names, ", ")
}

func spellHigherLevelToString(blocks []entryBlock) string {
	var parts []string
	for _, b := range blocks {
		if t := flattenEntries(b.Entries); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, "\n")
}

func ingestSpells(ctx context.Context, q *db.Queries) error {
	log.Printf("fetching spell index.json")
	var index map[string]string
	if err := fetchJSON(spellDataBaseURL+"index.json", &index); err != nil {
		return err
	}

	seen := map[string]bool{}
	total := 0
	for source, file := range index {
		if spellSourceExclude[source] {
			continue
		}
		log.Printf("fetching %s", file)
		var sf spellFile
		if err := fetchJSON(spellDataBaseURL+file, &sf); err != nil {
			log.Printf("warning: failed to fetch spells for %s: %v", source, err)
			continue
		}
		for _, sp := range sf.Spell {
			if seen[sp.Name] {
				continue
			}
			seen[sp.Name] = true

			duration, conc := spellDurationToString(sp.Duration)
			params := db.UpsertSpellParams{
				Name:          sp.Name,
				Level:         int64(sp.Level),
				School:        sql.NullString{String: schoolName(sp.School), Valid: true},
				Ritual:        sp.Meta.Ritual,
				CastingTime:   sql.NullString{String: spellTimeToString(sp.Time), Valid: true},
				Range:         sql.NullString{String: spellRangeToString(sp.Range), Valid: true},
				Components:    sql.NullString{String: spellComponentsToString(sp.Components), Valid: true},
				Duration:      sql.NullString{String: duration, Valid: true},
				Concentration: conc,
				Classes:       sql.NullString{String: spellClassesToString(sp.Classes), Valid: true},
				Description:   sql.NullString{String: flattenEntries(sp.Entries), Valid: true},
				HigherLevel:   sql.NullString{String: spellHigherLevelToString(sp.EntriesHigherLevel), Valid: true},
				Source:        sql.NullString{String: sp.Source, Valid: true},
			}
			if err := q.UpsertSpell(ctx, params); err != nil {
				return fmt.Errorf("upsert spell %q: %w", sp.Name, err)
			}
			total++
		}
	}
	log.Printf("done: %d spells ingested", total)
	return nil
}
