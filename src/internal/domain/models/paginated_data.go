package models

type PaginatedData[T any] struct {
	Data  []T
	Page  int
	Limit int
	Total int
}

const MaxLimit = 1000
