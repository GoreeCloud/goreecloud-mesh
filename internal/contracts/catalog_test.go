package contracts

import "testing"

func TestCatalogCoversMandatoryPlatforms(t *testing.T) {
	entries := Catalog()
	if len(entries) != len(Mandatory()) {
		t.Fatalf("catalog entries = %d, mandatory platforms = %d", len(entries), len(Mandatory()))
	}
	if len(entries) != 7 {
		t.Fatalf("catalog entries = %d, expected seven Integral Platform Systems", len(entries))
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

	for _, platform := range []Platform{Manager, PrivacyShield, Wardveil, Everkeep, GlazeUI, Mesh, Identity} {
		if !seen[platform] {
			t.Fatalf("Integral Platform System %q missing from catalog", platform)
		}
	}
}

func TestCatalogUsesCurrentCanonicalGlazeRepository(t *testing.T) {
	entry, ok := CatalogFor(GlazeUI)
	if !ok {
		t.Fatal("Glaze UI missing from platform catalog")
	}
	if entry.Repository != "GoreeCloud/goreecloud-glaze-ui" {
		t.Fatalf("Glaze UI repository = %q, expected canonical GoreeCloud/goreecloud-glaze-ui", entry.Repository)
	}
}

func TestCatalogForUnknownPlatformFailsClosed(t *testing.T) {
	if _, ok := CatalogFor(Platform("unknown")); ok {
		t.Fatal("unknown platform unexpectedly resolved")
	}
}
