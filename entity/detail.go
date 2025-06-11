package entity

type Detail struct {
	Title            string           `json:"Title"`
	AlternativeTitle AlternativeTitle `json:"alternative_titles"`
	StartDate        string           `json:"start_date"`
	Synopsis         string           `json:"synopsis"`
	Rank             int              `json:"rank"`
	Popularity       int              `json:"popularity"`
	Status           string           `json:"status"`
	Genres           []Genre          `json:"genres"`
	Rating           string           `json:"rating`
	Background       string           `json:"background,omitzero"`
	RelatedAnimes    []RelatedAnime   `json:"related_anime"`
	Recomendations   []Recomendation  `json:"recommendations"`
	Studios          []Studio         `json:"studios"`
}

// we only care about the english and japan alternative title
type AlternativeTitle struct {
	EngTitle string `json:"en"`
	JaTitle  string `json:"ja"`
}

type Genre struct {
	Name string `json:"name"`
}

type RelatedAnime struct {
	Node         Node   `json:"node"`
	RelationType string `json:"relation_type_formatted"`
}

type Recomendation struct {
	Node Node `json:"node"`
}

type Node struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type Studio struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
