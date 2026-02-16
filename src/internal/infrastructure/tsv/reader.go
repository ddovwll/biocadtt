package tsv

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
)

type TsvReader struct {
	logger *slog.Logger
}

func NewTsvReader(logger *slog.Logger) *TsvReader {
	return &TsvReader{
		logger: logger,
	}
}

const deviceTSVCols = 15

func (r *TsvReader) ReadDeviceData(ctx context.Context, filename string) ([]models.DeviceData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("failed to close file %s: %v", filename, err)
		}
	}()

	tsvReader := csv.NewReader(file)
	tsvReader.Comma = '\t'

	i := 1

	var data []models.DeviceData
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		rec, err := tsvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if i < 3 {
			i++
			continue
		}

		parsed, err := tsvRecordToDeviceData(rec)
		if err != nil {
			return nil, err
		}

		data = append(data, parsed)
	}

	if len(data) == 0 {
		return nil, errors.New("no data")
	}

	return data, nil
}

func tsvRecordToDeviceData(rec []string) (models.DeviceData, error) {
	if len(rec) != deviceTSVCols {
		return models.DeviceData{}, fmt.Errorf("invalid record: want %d cols, got %d", deviceTSVCols, len(rec))
	}

	trim := func(s string) string { return strings.TrimSpace(s) }

	nStr := trim(rec[0])
	if nStr == "" {
		return models.DeviceData{}, fmt.Errorf("invalid record: empty n")
	}
	n64, err := strconv.ParseInt(nStr, 10, 64)
	if err != nil {
		return models.DeviceData{}, fmt.Errorf("invalid n %q: %w", nStr, err)
	}

	levelStr := trim(rec[8])
	if levelStr == "" {
		return models.DeviceData{}, fmt.Errorf("invalid record: empty level")
	}
	level64, err := strconv.ParseInt(levelStr, 10, 32)
	if err != nil {
		return models.DeviceData{}, fmt.Errorf("invalid level %q: %w", levelStr, err)
	}

	d := models.DeviceData{
		N:         int(n64),
		MQTT:      trim(rec[1]),
		Invid:     trim(rec[2]),
		UnitGuid:  trim(rec[3]),
		MsgID:     trim(rec[4]),
		Text:      trim(rec[5]),
		Context:   trim(rec[6]),
		Class:     trim(rec[7]),
		Level:     int(level64),
		Area:      trim(rec[9]),
		Addr:      trim(rec[10]),
		Block:     trim(rec[11]),
		Type:      trim(rec[12]),
		Bit:       trim(rec[13]),
		InvertBit: trim(rec[14]),
	}

	return d, nil
}
