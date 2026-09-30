package snapshot

type Snapshot struct {
	Version  int       `json:"version"`
	Mappings []Mapping `json:"mappings"`
}

type Mapping struct {
	Registry    string   `json:"registry"`
	Scope       string   `json:"scope,omitempty"`
	Route       string   `json:"route"`
	Origin      Endpoint `json:"origin"`
	Destination Endpoint `json:"destination"`
	Symbol      string   `json:"symbol,omitempty"`
	Active      *bool    `json:"active,omitempty"`
	Supersedes  string   `json:"supersedes,omitempty"`
}

type Endpoint struct {
	Chain    string `json:"chain"`
	Kind     string `json:"kind"`
	Address  string `json:"address"`
	Decimals uint8  `json:"decimals"`
}

func (m Mapping) IsActive() bool {
	return m.Active == nil || *m.Active
}

func (m Mapping) Identity() string {
	return m.Registry + "|" + m.Scope + "|" + m.Route + "|" + m.Origin.Chain + "|" + m.Origin.Address + "|" + m.Destination.Chain
}
