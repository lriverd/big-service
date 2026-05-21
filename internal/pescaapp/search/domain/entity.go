package domain

type SearchResult struct {
	Spots   []SpotResult    `json:"spots"`
	Species []SpeciesResult `json:"species"`
	Users   []UserResult    `json:"users"`
}

type SpotResult struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Region string  `json:"region"`
	Rating float64 `json:"rating"`
}

type SpeciesResult struct {
	ID             string `json:"id"`
	CommonName     string `json:"commonName"`
	ScientificName string `json:"scientificName"`
}

type UserResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

