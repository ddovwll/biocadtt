package models

import "github.com/google/uuid"

type DeviceData struct {
	ID        uuid.UUID
	N         int
	MQTT      string
	Invid     string
	UnitGuid  string
	MsgID     string
	Text      string
	Context   string
	Class     string
	Level     int
	Area      string
	Addr      string
	Block     string
	Type      string
	Bit       string
	InvertBit string
}
