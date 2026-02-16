package services

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/ddovwll/biocadtt/src/internal/application/mocks"
	"github.com/ddovwll/biocadtt/src/internal/domain"
	"github.com/ddovwll/biocadtt/src/internal/domain/models"
	"go.uber.org/mock/gomock"
)

func TestDeviceDataService_ProcessFile_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deviceRepo := mocks.NewMockDeviceDataRepository(ctrl)
	processedRepo := mocks.NewMockProcessedFileRepository(ctrl)
	tsvReader := mocks.NewMockTsvReader(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)
	reportGen := mocks.NewMockReportGenerator(ctrl)

	svc := NewDeviceDataService(deviceRepo, processedRepo, tsvReader, txManager, reportGen)

	ctx := context.Background()
	filename := "file.tsv"
	data := []models.DeviceData{
		{N: 1, UnitGuid: "u-1"},
		{N: 2, UnitGuid: "u-1"},
		{N: 3, UnitGuid: "u-2"},
	}

	txManager.EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(txCtx context.Context, fn func(context.Context) error) error {
			return fn(txCtx)
		})

	processedRepo.EXPECT().FileExists(ctx, filename).Return(false, nil)
	tsvReader.EXPECT().ReadDeviceData(ctx, filename).Return(data, nil)
	deviceRepo.EXPECT().CreateBatch(ctx, data).Return(nil)
	processedRepo.EXPECT().
		Create(ctx, gomock.AssignableToTypeOf(models.ProcessedFile{})).
		DoAndReturn(func(_ context.Context, pf models.ProcessedFile) error {
			if !pf.ProcessedSuccessfully {
				t.Fatalf("expected successful processed file record")
			}
			if pf.FileName != filename {
				t.Fatalf("unexpected filename: %s", pf.FileName)
			}
			return nil
		})

	calls := map[string]int{}
	mu := &sync.Mutex{}
	reportGen.EXPECT().
		Generate(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, guid string, rows []models.DeviceData) error {
			mu.Lock()
			defer mu.Unlock()
			calls[guid]++
			if len(rows) == 0 {
				t.Fatalf("report data is empty for %s", guid)
			}
			return nil
		}).
		Times(2)

	err := svc.ProcessFile(ctx, filename)
	if err != nil {
		t.Fatalf("ProcessFile() error = %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected reports for 2 unit guids, got %d", len(calls))
	}
}

func TestDeviceDataService_ProcessFile_AlreadyProcessed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deviceRepo := mocks.NewMockDeviceDataRepository(ctrl)
	processedRepo := mocks.NewMockProcessedFileRepository(ctrl)
	tsvReader := mocks.NewMockTsvReader(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)
	reportGen := mocks.NewMockReportGenerator(ctrl)

	svc := NewDeviceDataService(deviceRepo, processedRepo, tsvReader, txManager, reportGen)

	ctx := context.Background()
	filename := "file.tsv"

	txManager.EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(txCtx context.Context, fn func(context.Context) error) error {
			return fn(txCtx)
		})
	processedRepo.EXPECT().FileExists(ctx, filename).Return(true, nil)

	err := svc.ProcessFile(ctx, filename)
	if !errors.Is(err, domain.ErrFileProcessed) {
		t.Fatalf("expected ErrFileProcessed, got %v", err)
	}
}

func TestDeviceDataService_ProcessFile_ReadErrorWritesFailureRecord(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deviceRepo := mocks.NewMockDeviceDataRepository(ctrl)
	processedRepo := mocks.NewMockProcessedFileRepository(ctrl)
	tsvReader := mocks.NewMockTsvReader(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)
	reportGen := mocks.NewMockReportGenerator(ctrl)

	svc := NewDeviceDataService(deviceRepo, processedRepo, tsvReader, txManager, reportGen)

	ctx := context.Background()
	filename := "file.tsv"
	readErr := errors.New("read failed")

	txManager.EXPECT().
		WithinTransaction(ctx, gomock.Any()).
		DoAndReturn(func(txCtx context.Context, fn func(context.Context) error) error {
			return fn(txCtx)
		})
	processedRepo.EXPECT().FileExists(ctx, filename).Return(false, nil)
	tsvReader.EXPECT().ReadDeviceData(ctx, filename).Return(nil, readErr)
	processedRepo.EXPECT().
		Create(ctx, gomock.AssignableToTypeOf(models.ProcessedFile{})).
		DoAndReturn(func(_ context.Context, pf models.ProcessedFile) error {
			if pf.ProcessedSuccessfully {
				t.Fatalf("expected failed processed file record")
			}
			if !strings.Contains(pf.ErrorMessage, readErr.Error()) {
				t.Fatalf("unexpected error message: %s", pf.ErrorMessage)
			}
			return nil
		})

	err := svc.ProcessFile(ctx, filename)
	if !errors.Is(err, readErr) {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestDeviceDataService_GetUnitData_MaxTakeExceeded(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deviceRepo := mocks.NewMockDeviceDataRepository(ctrl)
	processedRepo := mocks.NewMockProcessedFileRepository(ctrl)
	tsvReader := mocks.NewMockTsvReader(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)
	reportGen := mocks.NewMockReportGenerator(ctrl)

	svc := NewDeviceDataService(deviceRepo, processedRepo, tsvReader, txManager, reportGen)

	_, err := svc.GetUnitData(context.Background(), "u-1", models.MaxTake+1, 0)
	if !errors.Is(err, domain.ErrMaxTakeExceeded) {
		t.Fatalf("expected ErrMaxTakeExceeded, got %v", err)
	}
}

func TestDeviceDataService_GetUnitData_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deviceRepo := mocks.NewMockDeviceDataRepository(ctrl)
	processedRepo := mocks.NewMockProcessedFileRepository(ctrl)
	tsvReader := mocks.NewMockTsvReader(ctrl)
	txManager := mocks.NewMockTxManager(ctrl)
	reportGen := mocks.NewMockReportGenerator(ctrl)

	svc := NewDeviceDataService(deviceRepo, processedRepo, tsvReader, txManager, reportGen)

	ctx := context.Background()
	unitGUID := "u-1"
	take := 10
	offset := 0
	expected := models.PaginatedData[models.DeviceData]{
		Data:   []models.DeviceData{{N: 1, UnitGuid: unitGUID}},
		Take:   take,
		Offset: offset,
		Total:  1,
	}

	deviceRepo.EXPECT().
		GetByUnitUUIDPaginated(ctx, unitGUID, take, offset).
		Return(expected, nil)

	got, err := svc.GetUnitData(ctx, unitGUID, take, offset)
	if err != nil {
		t.Fatalf("GetUnitData() error = %v", err)
	}
	if got.Total != expected.Total || got.Take != expected.Take || got.Offset != expected.Offset {
		t.Fatalf("unexpected pagination: %+v", got)
	}
	if len(got.Data) != 1 || got.Data[0].UnitGuid != unitGUID {
		t.Fatalf("unexpected data: %+v", got.Data)
	}
}
