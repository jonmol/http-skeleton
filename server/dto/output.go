package dto

type OutputHello struct {
	Response string `json:"response"`
}

type Meta struct {
	Total    uint64 `json:"total"`
	ThisWord uint64 `json:"thisWord"`
}
