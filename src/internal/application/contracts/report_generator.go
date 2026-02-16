package contracts

import (
	"context"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
)

type ReportGenerator interface {
	Generate(ctx context.Context, unitGUID string, data []models.DeviceData) error
}
