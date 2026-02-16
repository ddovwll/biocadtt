package repositories

import (
	"context"
	"fmt"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
	"github.com/ddovwll/biocadtt/src/internal/infrastructure/data"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceDataRepo struct {
	pool *pgxpool.Pool
}

func NewDeviceDataRepository(pool *pgxpool.Pool) *DeviceDataRepo {
	return &DeviceDataRepo{pool: pool}
}

func (r *DeviceDataRepo) CreateBatch(ctx context.Context, deviceData []models.DeviceData) error {
	q := data.QuerierFromContext(ctx, r.pool)

	if len(deviceData) == 0 {
		return nil
	}

	cols := []string{
		"id",
		"n",
		"mqtt",
		"invid",
		"unit_guid",
		"msg_id",
		"text",
		"context",
		"class",
		"level",
		"area",
		"addr",
		"block",
		"type",
		"bit",
		"invert_bit",
	}

	rows := make([][]any, 0, len(deviceData))
	for _, d := range deviceData {
		d.ID = uuid.New()
		rows = append(rows, []any{
			d.ID,
			d.N,
			d.MQTT,
			d.Invid,
			d.UnitGuid,
			d.MsgID,
			d.Text,
			d.Context,
			d.Class,
			d.Level,
			d.Area,
			d.Addr,
			d.Block,
			d.Type,
			d.Bit,
			d.InvertBit,
		})
	}

	_, err := q.CopyFrom(
		ctx,
		pgx.Identifier{"file_data"},
		cols,
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy file_data: %w", err)
	}
	return nil
}

func (r *DeviceDataRepo) GetByUnitUUIDPaginated(
	ctx context.Context,
	unitGuid string,
	page, limit int,
) (models.PaginatedData[models.DeviceData], error) {
	q := data.QuerierFromContext(ctx, r.pool)
	offset := (page - 1) * limit

	var total int
	if err := q.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM file_data WHERE unit_guid = $1`,
		unitGuid,
	).Scan(&total); err != nil {
		return models.PaginatedData[models.DeviceData]{}, fmt.Errorf("count file_data: %w", err)
	}

	query := `
SELECT
	id,
	n,
	mqtt,
	invid,
	unit_guid,
	msg_id,
	text,
	context,
	class,
	level,
	area,
	addr,
	block,
	type,
	bit,
	invert_bit
FROM file_data
WHERE unit_guid = $1
ORDER BY unit_guid
LIMIT $2 OFFSET $3;
`

	rows, err := r.pool.Query(ctx, query, unitGuid, limit, offset)
	if err != nil {
		return models.PaginatedData[models.DeviceData]{}, fmt.Errorf("select file_data page: %w", err)
	}
	defer rows.Close()

	items := make([]models.DeviceData, 0, limit)
	for rows.Next() {
		var d models.DeviceData
		if err := rows.Scan(
			&d.ID,
			&d.N,
			&d.MQTT,
			&d.Invid,
			&d.UnitGuid,
			&d.MsgID,
			&d.Text,
			&d.Context,
			&d.Class,
			&d.Level,
			&d.Area,
			&d.Addr,
			&d.Block,
			&d.Type,
			&d.Bit,
			&d.InvertBit,
		); err != nil {
			return models.PaginatedData[models.DeviceData]{}, fmt.Errorf("scan file_data: %w", err)
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return models.PaginatedData[models.DeviceData]{}, fmt.Errorf("iterate file_data: %w", err)
	}

	return models.PaginatedData[models.DeviceData]{
		Data:  items,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}
