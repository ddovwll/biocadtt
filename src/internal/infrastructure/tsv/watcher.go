package tsv

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ddovwll/biocadtt/src/internal/application/services"
	"github.com/ddovwll/biocadtt/src/internal/domain"
)

type TsvWatcher struct {
	deviceDataService *services.DeviceDataService
	logger            *slog.Logger
	directory         string
	fileNames         chan string
	processingFiles   map[string]struct{}
	mutex             sync.Mutex
}

const (
	fileProcessTimeout = time.Second * 10
	processQueueLength = 100
)

func NewTsvWatcher(deviceDataService *services.DeviceDataService, logger *slog.Logger, directory string) *TsvWatcher {
	return &TsvWatcher{
		deviceDataService: deviceDataService,
		logger:            logger,
		directory:         directory,
		fileNames:         make(chan string, processQueueLength),
		processingFiles:   make(map[string]struct{}),
	}
}

func (s *TsvWatcher) Start(ctx context.Context, heartbeat time.Duration) {
	wg := &sync.WaitGroup{}
	s.startWorkers(ctx, wg)
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()
	s.checkFolder()

	for {
		select {
		case <-ctx.Done():
			close(s.fileNames)
			wg.Wait()
			return
		case <-ticker.C:
			s.checkFolder()
		}
	}
}

func (s *TsvWatcher) checkFolder() {
	entry, err := os.ReadDir(s.directory)
	if err != nil {
		s.logger.Error("error reading directory",
			"folder", s.directory,
			"error", err,
		)

		return
	}

	for _, file := range entry {
		info, err := file.Info()
		if err != nil {
			s.logger.Error("error reading file",
				"file", file.Name(),
				"error", err,
			)

			continue
		}

		if info.IsDir() || !strings.HasSuffix(file.Name(), ".tsv") {
			continue
		}

		s.mutex.Lock()
		if _, ok := s.processingFiles[file.Name()]; ok {
			s.mutex.Unlock()
			continue
		}

		s.processingFiles[file.Name()] = struct{}{}
		s.mutex.Unlock()

		select {
		case s.fileNames <- file.Name():
		default:
			s.mutex.Lock()
			delete(s.processingFiles, file.Name())
			s.mutex.Unlock()

			s.logger.Warn("queue is full, skip file for now", "file", file.Name())
		}
	}
}

func (s *TsvWatcher) startWorkers(ctx context.Context, wg *sync.WaitGroup) {
	for range runtime.GOMAXPROCS(0) {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case filename, ok := <-s.fileNames:
					if !ok {
						return
					}

					s.processFile(ctx, filename)
				}
			}
		}()
	}
}

func (s *TsvWatcher) processFile(ctx context.Context, filename string) {
	defer func() {
		s.mutex.Lock()
		delete(s.processingFiles, filename)
		s.mutex.Unlock()
	}()

	processCtx, cancel := context.WithTimeout(ctx, fileProcessTimeout)
	defer cancel()

	err := s.deviceDataService.ProcessFile(processCtx, filepath.Join(s.directory, filename))
	if err != nil {
		if errors.Is(err, domain.ErrFileProcessed) {
			s.logger.Debug("error processing file",
				"filename", filename,
				"error", err,
			)

			return
		}

		s.logger.Error("error processing file",
			"filename", filename,
			"error", err,
		)

		return
	}

	s.logger.Info("file processed successfully", "filename", filename)
}
