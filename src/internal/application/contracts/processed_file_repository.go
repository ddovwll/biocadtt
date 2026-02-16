package contracts

import (
	"context"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
)

type ProcessedFileRepository interface {
	Create(ctx context.Context, data models.ProcessedFile) error
	FileExists(ctx context.Context, filename string) (bool, error)
}
