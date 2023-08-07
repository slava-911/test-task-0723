package dmodel

type Product struct {
	Id          string   `json:"id"`
	Price       int      `json:"price"`
	Quantity    int      `json:"quantity"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}
