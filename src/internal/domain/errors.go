package domain

import (
	"errors"
	"fmt"

	"github.com/ddovwll/biocadtt/src/internal/domain/models"
)

var (
	ErrFileProcessed    = errors.New("file processed already")
	ErrMaxLimitExceeded = fmt.Errorf("max limit exceeded, max limit is %d", models.MaxLimit)
)
