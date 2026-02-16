package contracts

import (
	"context"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
)

type TsvReader interface {
	ReadDeviceData(ctx context.Context, filename string) ([]models.DeviceData, error)
}
