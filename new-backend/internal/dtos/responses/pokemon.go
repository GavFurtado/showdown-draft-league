package responses

type PokemonSpeciesListResponse struct {
	ID           int64    `json:"ID"`
	Name         string   `json:"Name"`
	Types        []string `json:"Types"`
	FrontDefault string   `json:"FrontDefault"`
}
