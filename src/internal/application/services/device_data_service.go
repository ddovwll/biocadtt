package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ddovwll/biocadtt/src/internal/application/contracts"
	"github.com/ddovwll/biocadtt/src/internal/domain"
	"github.com/ddovwll/biocadtt/src/internal/domain/models"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type DeviceDataService struct {
	deviceDataRepo    contracts.DeviceDataRepository
	processedFileRepo contracts.ProcessedFileRepository
	tsvReader         contracts.TsvReader
	txManager         contracts.TxManager
	reportGenerator   contracts.ReportGenerator
}

func NewDeviceDataService(
	deviceDataRepo contracts.DeviceDataRepository,
	processedFileRepo contracts.ProcessedFileRepository,
	tsvReader contracts.TsvReader,
	txManager contracts.TxManager,
	reportGenerator contracts.ReportGenerator,
) *DeviceDataService {
	return &DeviceDataService{
		deviceDataRepo:    deviceDataRepo,
		processedFileRepo: processedFileRepo,
		tsvReader:         tsvReader,
		txManager:         txManager,
		reportGenerator:   reportGenerator,
	}
}

func (s *DeviceDataService) ProcessFile(ctx context.Context, filename string) error {
	err := s.txManager.WithinTransaction(ctx, func(ctx context.Context) error {
		exists, err := s.processedFileRepo.FileExists(ctx, filename)
		if err != nil {
			return err
		}

		if exists {
			return domain.ErrFileProcessed
		}

		data, err := s.tsvReader.ReadDeviceData(ctx, filename)
		if err != nil {
			return err
		}

		err = s.deviceDataRepo.CreateBatch(ctx, data)
		if err != nil {
			return err
		}

		processedFile := models.ProcessedFile{
			ID:                    uuid.New(),
			FileName:              filename,
			ProcessedAt:           time.Now().UTC(),
			ProcessedSuccessfully: true,
			ErrorMessage:          "",
		}

		err = s.processedFileRepo.Create(ctx, processedFile)
		if err != nil {
			return err
		}

		err = s.generateReports(ctx, data)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, domain.ErrFileProcessed) {
			return domain.ErrFileProcessed
		}
		processErr := err

		processedFile := models.ProcessedFile{
			ID:                    uuid.New(),
			FileName:              filename,
			ProcessedAt:           time.Now().UTC(),
			ProcessedSuccessfully: false,
			ErrorMessage:          fmt.Sprintf("fail to process file %s", processErr),
		}

		saveErr := s.processedFileRepo.Create(ctx, processedFile)
		if saveErr != nil {
			return errors.Join(processErr, saveErr)
		}

		reportErr := s.reportGenerator.GenerateError(ctx, processedFile)
		if reportErr != nil {
			return errors.Join(processErr, reportErr)
		}

		return processErr
	}

	return err
}

func (s *DeviceDataService) generateReports(ctx context.Context, data []models.DeviceData) error {
	units := map[string][]models.DeviceData{}
	for _, d := range data {
		if _, ok := units[d.UnitGuid]; !ok {
			units[d.UnitGuid] = []models.DeviceData{d}
			continue
		}

		units[d.UnitGuid] = append(units[d.UnitGuid], d)
	}

	eg, egCtx := errgroup.WithContext(ctx)
	for GUID, unit := range units {
		eg.Go(func() error {
			return s.reportGenerator.GenerateReport(egCtx, GUID, unit)
		})
	}

	return eg.Wait()
}

func (s *DeviceDataService) GetUnitData(
	ctx context.Context,
	unitGUID string,
	page, limit int,
) (models.PaginatedData[models.DeviceData], error) {
	if limit > models.MaxLimit {
		return models.PaginatedData[models.DeviceData]{}, domain.ErrMaxLimitExceeded
	}
	return s.deviceDataRepo.GetByUnitUUIDPaginated(ctx, unitGUID, page, limit)
}
