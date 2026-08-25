package contracts

import "testing"

func validManifestFor(entry CatalogEntry) IntegrationManifest {
	return IntegrationManifest{
		SchemaVersion:               IntegrationManifestSchemaVersion,
		Platform:                    entry.Platform,
		ProducerRepository:          entry.Repository,
		ConsumerRepository:          MeshRepository,
		AuthoritativeContract:       entry.ContractSource,
		IntegrationMode:             ReadOnlyCoordination,
		AuthorityTransfer:           false,
		RuntimeAcceptanceImplied:    false,
		StableAcceptanceImplied:     false,
		SensitiveInformationAllowed: false,
		EvidenceValidity: EvidenceValidity{
			Authority:                      ProducerOrApplicablePolicy,
			ValidUntilRequiredForValidated: true,
			ConsumerOverrideAllowed:        false,
		},
		Purpose: "bounded Mesh coordination",
	}
}

func TestIntegrationManifestValidForAllCatalogEntries(t *testing.T) {
	for _, entry := range Catalog() {
		if err := ValidateIntegrationManifest(entry, validManifestFor(entry)); err != nil {
			t.Fatalf("valid manifest for %q rejected: %v", entry.Platform, err)
		}
	}
}

func TestIntegrationManifestFailsClosedOnAuthorityTransfer(t *testing.T) {
	entry, _ := CatalogFor(Wardveil)
	manifest := validManifestFor(entry)
	manifest.AuthorityTransfer = true
	if err := ValidateIntegrationManifest(entry, manifest); err == nil {
		t.Fatal("authority transfer unexpectedly accepted")
	}
}

func TestIntegrationManifestFailsClosedOnAcceptanceClaims(t *testing.T) {
	entry, _ := CatalogFor(PrivacyShield)
	manifest := validManifestFor(entry)
	manifest.RuntimeAcceptanceImplied = true
	if err := ValidateIntegrationManifest(entry, manifest); err == nil {
		t.Fatal("runtime acceptance implication unexpectedly accepted")
	}
	manifest = validManifestFor(entry)
	manifest.StableAcceptanceImplied = true
	if err := ValidateIntegrationManifest(entry, manifest); err == nil {
		t.Fatal("Stable acceptance implication unexpectedly accepted")
	}
}

func TestIntegrationManifestFailsClosedOnSensitiveInformation(t *testing.T) {
	entry, _ := CatalogFor(Everkeep)
	manifest := validManifestFor(entry)
	manifest.SensitiveInformationAllowed = true
	if err := ValidateIntegrationManifest(entry, manifest); err == nil {
		t.Fatal("sensitive-information permission unexpectedly accepted")
	}
}

func TestIntegrationManifestFailsClosedOnCatalogMismatch(t *testing.T) {
	entry, _ := CatalogFor(GlazeUI)
	manifest := validManifestFor(entry)
	manifest.AuthoritativeContract = "wrong-contract"
	if err := ValidateIntegrationManifest(entry, manifest); err == nil {
		t.Fatal("authoritative-contract mismatch unexpectedly accepted")
	}
}

func TestIntegrationManifestFailsClosedOnEvidenceValidityWeakening(t *testing.T) {
	entry, _ := CatalogFor(Wardveil)

	manifest := validManifestFor(entry)
	manifest.EvidenceValidity.Authority = "mesh"
	if err := ValidateIntegrationManifest(entry, manifest); err == nil {
		t.Fatal("Mesh evidence-validity authority unexpectedly accepted")
	}

	manifest = validManifestFor(entry)
	manifest.EvidenceValidity.ValidUntilRequiredForValidated = false
	if err := ValidateIntegrationManifest(entry, manifest); err == nil {
		t.Fatal("optional valid_until unexpectedly accepted")
	}

	manifest = validManifestFor(entry)
	manifest.EvidenceValidity.ConsumerOverrideAllowed = true
	if err := ValidateIntegrationManifest(entry, manifest); err == nil {
		t.Fatal("consumer validity override unexpectedly accepted")
	}
}

func TestDecodeIntegrationManifestRejectsUnknownFields(t *testing.T) {
	data := []byte(`{"schema_version":2,"platform":"glaze-ui","producer_repository":"GoreeCloud/glaze-ui","consumer_repository":"GoreeCloud/goreecloud-mesh","authoritative_contract":"CONFORMANCE.md","integration_mode":"read-only-coordination","authority_transfer":false,"runtime_acceptance_implied":false,"stable_acceptance_implied":false,"sensitive_information_allowed":false,"evidence_validity":{"authority":"producer-or-applicable-policy","valid_until_required_for_validated":true,"consumer_override_allowed":false},"purpose":"bounded coordination","unexpected":true}`)
	if _, err := DecodeIntegrationManifest(data); err == nil {
		t.Fatal("unknown manifest field unexpectedly accepted")
	}
}
