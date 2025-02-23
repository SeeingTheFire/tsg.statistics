package dtos

type ReplayRow struct {
	Rows   []Row  `json:"rows"`
	Total  int    `json:"total"`
	Source string `json:"source"`
	Error  string `json:"error"`
}

type Row struct {
	Name     string   `json:"name"`
	Archive  int      `json:"archive"`
	FileSize int      `json:"fileSize"`
	Array    []string `json:"array"`
}

type Answer struct {
	Json  string `json:"json"`
	Error string `json:"error"`
}
