package contracts

import (
	"context"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
)

type DeviceDataRepository interface {
	CreateBatch(ctx context.Context, data []models.DeviceData) error
	GetByUnitUUIDPaginated(
		ctx context.Context,
		uuid string,
		page, limit int,
	) (models.PaginatedData[models.DeviceData], error)
}
