package contracts

// CatalogEntry describes one mandatory Integral Platform System without
// transferring that system's authority into GoreeCloud Mesh.
type CatalogEntry struct {
	Platform            Platform `json:"platform"`
	DisplayName         string   `json:"display_name"`
	Repository          string   `json:"repository"`
	Authority           string   `json:"authority"`
	ContractSource      string   `json:"contract_source"`
	IntegrationManifest string   `json:"integration_manifest"`
	Required            bool     `json:"required"`
}

var catalog = []CatalogEntry{
	{
		Platform:            Manager,
		DisplayName:         "GoreeCloud Manager",
		Repository:          "GoreeCloud/goreecloud-manager",
		Authority:           "bounded platform management, inventory, lifecycle, operational visibility, and management-plane state",
		ContractSource:      "goreecloud.platform.yaml",
		IntegrationManifest: "goreecloud.platform.yaml",
		Required:            true,
	},
	{
		Platform:            PrivacyShield,
		DisplayName:         "Privacy Shield",
		Repository:          "GoreeCloud/goreecloud-privacy-shield",
		Authority:           "privacy controls, privacy status, data minimization, retention expectations, and privacy capability governance",
		ContractSource:      "contracts/privacy-shield.platform.json",
		IntegrationManifest: "contracts/mesh.integration.json",
		Required:            true,
	},
	{
		Platform:            Wardveil,
		DisplayName:         "Wardveil Security",
		Repository:          "GoreeCloud/goreecloud-wardveil-security",
		Authority:           "security and protection status, evidence semantics, and protection presentation",
		ContractSource:      "contracts/wardveil.status.schema.json",
		IntegrationManifest: "contracts/mesh.integration.json",
		Required:            true,
	},
	{
		Platform:            Everkeep,
		DisplayName:         "Everkeep",
		Repository:          "GoreeCloud/goreecloud-everkeep",
		Authority:           "resilience, recovery, preservation, portability, succession, and digital-legacy evidence",
		ContractSource:      "contracts/continuity.status.schema.json",
		IntegrationManifest: "contracts/mesh.integration.json",
		Required:            true,
	},
	{
		Platform:            GlazeUI,
		DisplayName:         "Glaze UI",
		Repository:          "GoreeCloud/goreecloud-glaze-ui",
		Authority:           "design, interaction, accessibility, adaptive behavior, and visual presentation",
		ContractSource:      "CONFORMANCE.md",
		IntegrationManifest: "contracts/mesh.integration.json",
		Required:            true,
	},
	{
		Platform:            Mesh,
		DisplayName:         "GoreeCloud Mesh",
		Repository:          "GoreeCloud/goreecloud-mesh",
		Authority:           "bounded first-party discovery, coordination, relationship, evidence-transport, and event/capability exchange state",
		ContractSource:      "contracts/mesh.platform-evidence-plane.v1.json",
		IntegrationManifest: "goreecloud.platform.yaml",
		Required:            true,
	},
	{
		Platform:            Identity,
		DisplayName:         "GoreeCloud Identity",
		Repository:          "GoreeCloud/goreecloud-identity",
		Authority:           "identity, authentication, authorization, accounts, devices, credentials, sessions, service identity, and delegated authority",
		ContractSource:      "contracts/identity.mesh-evidence-profile.json",
		IntegrationManifest: "contracts/identity.mesh-evidence-profile.json",
		Required:            true,
	},
}

// Catalog returns a copy so callers cannot mutate the canonical process-local catalog.
func Catalog() []CatalogEntry {
	return append([]CatalogEntry(nil), catalog...)
}

// CatalogFor returns the authority metadata for a mandatory platform system.
func CatalogFor(platform Platform) (CatalogEntry, bool) {
	for _, entry := range catalog {
		if entry.Platform == platform {
			return entry, true
		}
	}
	return CatalogEntry{}, false
}
