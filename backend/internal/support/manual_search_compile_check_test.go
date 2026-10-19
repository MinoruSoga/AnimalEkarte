package support

import "testing"

func TestSearchDictLoaded(t *testing.T) {
	if len(searchDictStopwords) == 0 || len(searchDictSuffixes) == 0 || len(searchDictSynonymGroups) == 0 {
		t.Fatal("search dictionary must be populated")
	}
	groups := extractSearchTermGroups("レジ締めの手順を教えてください")
	if len(groups) == 0 {
		t.Fatal("expected term groups")
	}
	found := false
	for _, g := range groups {
		for _, v := range g {
			if v == "レジ締め" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("expected レジ締め variant, got %v", groups)
	}
}
