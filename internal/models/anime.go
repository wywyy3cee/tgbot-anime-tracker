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
	AiredOn     string     `json:"aired_on"`
	ReleasedOn  string     `json:"released_on"`
	Description string     `json:"description"`
	Genres      []Genre    `json:"genres"`
}

type AnimeImage struct {
	Original string `json:"original"`
	Preview  string `json:"preview"`
}

type Genre struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Russian   string `json:"russian"`
	Kind      string `json:"kind"`
	EntryType string `json:"entry_type"`
}
