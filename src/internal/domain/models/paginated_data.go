package models

type PaginatedData[T any] struct {
	Data   []T
	Take   int
	Offset int
	Total  int
}

const MaxTake = 1000
