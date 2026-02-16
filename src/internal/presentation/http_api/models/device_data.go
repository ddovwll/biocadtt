package models

import "github.com/ddovwll/biocadtt/src/internal/domain/models"

type DeviceDataResponse struct {
	N         int    `json:"n"`
	MQTT      string `json:"mqtt"`
	Invid     string `json:"invid"`
	UnitGuid  string `json:"unit_guid"`
	MsgID     string `json:"msg_id"`
	Text      string `json:"text"`
	Context   string `json:"context"`
	Class     string `json:"class"`
	Level     int    `json:"level"`
	Area      string `json:"area"`
	Addr      string `json:"addr"`
	Block     string `json:"block"`
	Type      string `json:"type"`
	Bit       string `json:"bit"`
	InvertBit string `json:"invert_bit"`
}

func MapDeviceDataToResponse(data []models.DeviceData) []DeviceDataResponse {
	result := make([]DeviceDataResponse, 0, len(data))
	for _, d := range data {
		result = append(result, DeviceDataResponse{
			N:         d.N,
			MQTT:      d.MQTT,
			Invid:     d.Invid,
			UnitGuid:  d.UnitGuid,
			MsgID:     d.MsgID,
			Text:      d.Text,
			Context:   d.Context,
			Class:     d.Class,
			Level:     d.Level,
			Area:      d.Area,
			Addr:      d.Addr,
			Block:     d.Block,
			Type:      d.Type,
			Bit:       d.Bit,
			InvertBit: d.InvertBit,
		})
	}
	return result
}
