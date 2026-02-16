package models

import (
	"time"

	"github.com/google/uuid"
)

type ProcessedFile struct {
	ID                    uuid.UUID
	FileName              string
	ProcessedAt           time.Time
	ProcessedSuccessfully bool
	ErrorMessage          string
}
