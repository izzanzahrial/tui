package entity

type Anime struct {
	ID               int              `json:"id"`
	Title            string           `json:"title"`
	Image            Image            `json:"main_picture"`
	AlternativeTitle AlternativeTitle `json:"alternative_titles"`
}

type Image struct {
	Picture string `json:"medium"`
}

type Ranking struct {
	Rank int `json:"rank"`
}

type AnimeRank struct {
	Anime Anime   `json:"node"`
	Rank  Ranking `json:"ranking"`
}

type Data struct {
	AnimeRank []AnimeRank `json:"data"`
}
