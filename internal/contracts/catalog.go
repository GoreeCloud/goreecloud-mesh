package contracts

// CatalogEntry describes one mandatory integral GoreeCloud platform system
// without transferring that system's authority into Mesh.
type CatalogEntry struct {
	Platform       Platform `json:"platform"`
	DisplayName    string   `json:"display_name"`
	Repository     string   `json:"repository"`
	Authority      string   `json:"authority"`
	ContractSource string   `json:"contract_source"`
	Required       bool     `json:"required"`
}

var catalog = []CatalogEntry{
	{
		Platform:       GlazeUI,
		DisplayName:    "Glaze UI",
		Repository:     "GoreeCloud/glaze-ui",
		Authority:      "design, interaction, accessibility, adaptive behavior, and visual presentation",
		ContractSource: "CONFORMANCE.md",
		Required:       true,
	},
	{
		Platform:       Wardveil,
		DisplayName:    "Wardveil Security",
		Repository:     "GoreeCloud/goreecloud-wardveil-security",
		Authority:      "security and protection status, evidence semantics, and protection presentation",
		ContractSource: "contracts/wardveil.status.schema.json",
		Required:       true,
	},
	{
		Platform:       PrivacyShield,
		DisplayName:    "GoreeCloud Privacy Shield",
		Repository:     "GoreeCloud/goreecloud-privacy-shield",
		Authority:      "privacy controls, privacy status, data minimization, retention expectations, and privacy capability governance",
		ContractSource: "contracts/privacy-shield.platform.json",
		Required:       true,
	},
	{
		Platform:       Everkeep,
		DisplayName:    "Everkeep",
		Repository:     "GoreeCloud/goreecloud-everkeep",
		Authority:      "resilience, recovery, preservation, portability, succession, and digital-legacy evidence",
		ContractSource: "contracts/continuity.status.schema.json",
		Required:       true,
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
