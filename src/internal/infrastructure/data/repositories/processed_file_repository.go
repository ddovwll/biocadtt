package repositories

import (
	"context"
	"fmt"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
	"github.com/ddovwll/biocadtt/src/internal/infrastructure/data"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessedFileRepo struct {
	pool *pgxpool.Pool
}

func NewProcessedFileRepo(pool *pgxpool.Pool) *ProcessedFileRepo {
	return &ProcessedFileRepo{pool: pool}
}

func (r *ProcessedFileRepo) Create(ctx context.Context, processedFile models.ProcessedFile) error {
	q := data.QuerierFromContext(ctx, r.pool)

	_, err := q.Exec(ctx, `
INSERT INTO processed_files (id, file_name, processed_at, processed_successfully, error_message)
VALUES ($1, $2, $3, $4, $5)
`,
		processedFile.ID,
		processedFile.FileName,
		processedFile.ProcessedAt,
		processedFile.ProcessedSuccessfully,
		processedFile.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("insert processed_files: %w", err)
	}
	return nil
}

func (r *ProcessedFileRepo) FileExists(ctx context.Context, filename string) (bool, error) {
	q := data.QuerierFromContext(ctx, r.pool)

	var exists bool
	if err := q.QueryRow(ctx, `
SELECT EXISTS(SELECT 1 FROM processed_files WHERE file_name = $1)
`, filename).Scan(&exists); err != nil {
		return false, fmt.Errorf("check processed_files exists: %w", err)
	}
	return exists, nil
}
