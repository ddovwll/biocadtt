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
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")
	unitGUID := r.PathValue("unit_guid")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 100
	}

	data, err := c.deviceDataService.GetUnitData(r.Context(), unitGUID, pageInt, limitInt)
	if err != nil {
		if errors.Is(err, domain.ErrMaxLimitExceeded) {
			c.writeErrorResponse(w, http.StatusBadRequest, "max limit exceeded")
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
		Data:  models.MapDeviceDataToResponse(data.Data),
		Page:  data.Page,
		Limit: data.Limit,
		Total: data.Total,
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
