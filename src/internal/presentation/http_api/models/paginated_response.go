package models

type PaginatedResponse[T any] struct {
	Data   []T `json:"data"`
	Take   int `json:"take"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}
