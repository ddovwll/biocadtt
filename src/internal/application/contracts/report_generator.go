package contracts

import (
	"context"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
)

type ReportGenerator interface {
	GenerateReport(ctx context.Context, unitGUID string, data []models.DeviceData) error
	GenerateError(ctx context.Context, file models.ProcessedFile) error
}
