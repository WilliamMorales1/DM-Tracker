package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSpellcastingToJSONIncludesUseLists(t *testing.T) {
	var blocks []spellcastingBlock
	err := json.Unmarshal([]byte(`[{"name":"Spellcasting","headerEntries":["Tirion casts one of the following spells, using Charisma as the spellcasting ability (spell save DC 17):"],"daily":{"1":["{@spell word of recall}","{@spell raise dead}"],"2e":["{@spell dispel evil and good}"]}}]`), &blocks)
	if err != nil {
		t.Fatal(err)
	}

	got := spellcastingToJSON(blocks)
	for _, want := range []string{
		"1/day: word of recall, raise dead",
		"2/day each: dispel evil and good",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("spellcasting output missing %q: %s", want, got)
		}
	}
}
