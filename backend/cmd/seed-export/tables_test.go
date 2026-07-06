package main

import "testing"

func TestBundleTablesCountMatchesConstant(t *testing.T) {
	total := 0
	for _, tables := range bundleTables {
		total += len(tables)
	}
	if total != totalSeedTableCount {
		t.Fatalf("bundleTables has %d tables total, want %d (totalSeedTableCount)", total, totalSeedTableCount)
	}
}

func TestBundleTablesHasNoDuplicateTableAcrossBundles(t *testing.T) {
	seen := map[string]string{}
	for bundle, tables := range bundleTables {
		for _, table := range tables {
			if owner, ok := seen[table]; ok {
				t.Fatalf("table %q listed in both bundle %q and %q — a table must be owned by exactly one bundle", table, owner, bundle)
			}
			seen[table] = bundle
		}
	}
}

func TestBundleOrderCoversEveryBundleTablesKey(t *testing.T) {
	if len(bundleOrder) != len(bundleTables) {
		t.Fatalf("bundleOrder has %d entries, bundleTables has %d keys — they must match 1:1", len(bundleOrder), len(bundleTables))
	}
	for _, b := range bundleOrder {
		if _, ok := bundleTables[b]; !ok {
			t.Fatalf("bundleOrder references bundle %q which has no entry in bundleTables", b)
		}
	}
}
