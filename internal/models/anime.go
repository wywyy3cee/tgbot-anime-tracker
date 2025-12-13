package models

type Anime struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Russian     string     `json:"russian"`
	Image       AnimeImage `json:"image"`
	Kind        string     `json:"kind"`
	Score       string     `json:"score"`
	Status      string     `json:"status"`
	Episodes    int        `json:"episodes"`
	ReleasedOn  string     `json:"released_on"`
	Description string     `json:"description"`
}

type AnimeImage struct {
	Original string `json:"original"`
	Preview  string `json:"preview"`
}
