package contracts

import "testing"

func TestCatalogCoversMandatoryPlatforms(t *testing.T) {
	entries := Catalog()
	if len(entries) != len(Mandatory()) {
		t.Fatalf("catalog entries = %d, mandatory platforms = %d", len(entries), len(Mandatory()))
	}

	seen := map[Platform]bool{}
	for _, entry := range entries {
		if !IsMandatory(entry.Platform) {
			t.Fatalf("catalog contains non-mandatory platform %q", entry.Platform)
		}
		if seen[entry.Platform] {
			t.Fatalf("catalog contains duplicate platform %q", entry.Platform)
		}
		if entry.DisplayName == "" || entry.Repository == "" || entry.Authority == "" || entry.ContractSource == "" || entry.IntegrationManifest == "" {
			t.Fatalf("catalog entry for %q is incomplete: %+v", entry.Platform, entry)
		}
		if !entry.Required {
			t.Fatalf("mandatory platform %q is not marked required", entry.Platform)
		}
		seen[entry.Platform] = true
	}

	for _, platform := range Mandatory() {
		if !seen[platform] {
			t.Fatalf("mandatory platform %q missing from catalog", platform)
		}
	}
}

func TestCatalogForUnknownPlatformFailsClosed(t *testing.T) {
	if _, ok := CatalogFor(Platform("unknown")); ok {
		t.Fatal("unknown platform unexpectedly resolved")
	}
}
