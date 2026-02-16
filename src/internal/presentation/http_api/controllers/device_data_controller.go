package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ddovwll/biocadtt/src/internal/application/services"
	"github.com/ddovwll/biocadtt/src/internal/domain"
	"github.com/ddovwll/biocadtt/src/internal/presentation/http_api/models"
)

type DeviceDataController struct {
	deviceDataService *services.DeviceDataService
	logger            *slog.Logger
}

func NewDeviceDataController(deviceDataService *services.DeviceDataService, logger *slog.Logger) *DeviceDataController {
	return &DeviceDataController{
		deviceDataService: deviceDataService,
		logger:            logger,
	}
}

func (c *DeviceDataController) UseController(mux *http.ServeMux) {
	mux.HandleFunc("GET /device-data/{unit_guid}", c.GetDeviceData)
}

func (c *DeviceDataController) GetDeviceData(w http.ResponseWriter, r *http.Request) {
	take := r.URL.Query().Get("take")
	offset := r.URL.Query().Get("offset")
	unitGUID := r.PathValue("unit_guid")

	takeInt, err := strconv.Atoi(take)
	if err != nil || takeInt < 0 {
		takeInt = 100
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 0 {
		offsetInt = 0
	}

	data, err := c.deviceDataService.GetUnitData(r.Context(), unitGUID, takeInt, offsetInt)
	if err != nil {
		if errors.Is(err, domain.ErrMaxTakeExceeded) {
			c.writeErrorResponse(w, http.StatusBadRequest, "max take exceeded")
			return
		}

		c.writeErrorResponse(w, http.StatusInternalServerError, "internal server error")
		c.logger.Error("failed to get device data",
			"unit_guid", unitGUID,
			"err", err,
		)
		return
	}

	resp := models.PaginatedResponse[models.DeviceDataResponse]{
		Data:   models.MapDeviceDataToResponse(data.Data),
		Take:   data.Take,
		Offset: data.Offset,
		Total:  data.Total,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		c.logger.Error("failed to encode response",
			"err", err,
		)
	}
}

func (c *DeviceDataController) writeErrorResponse(w http.ResponseWriter, statusCode int, msg string) {
	errResponse := models.ErrorResponse{
		Error: msg,
	}

	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		c.logger.Error("failed to encode error response",
			"err", err,
		)
	}
}
