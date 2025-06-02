package proofs

type TraitRegion struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type TraitVariant struct {
	Trait      string      `json:"trait"`
	Gene       string      `json:"gene"`
	Chromosome int         `json:"chromosome"`
	Position   int         `json:"position"`
	Region     TraitRegion `json:"region"`
	Ref        string      `json:"ref"`
	Alt        string      `json:"alt"`
}

type TraitPanel struct{}
