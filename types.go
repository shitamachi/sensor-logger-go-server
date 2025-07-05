package main

import (
	"time"
)

// SensorMessage 表示完整的传感器消息结构
type SensorMessage struct {
	MessageID int64           `json:"messageId"`
	SessionID string          `json:"sessionId"`
	DeviceID  string          `json:"deviceId"`
	Payload   []SensorReading `json:"payload"`
}

// SensorReading 表示单个传感器读数
type SensorReading struct {
	Name     string                 `json:"name"`
	Time     int64                  `json:"time"`
	Values   map[string]interface{} `json:"values"`
	Accuracy int                    `json:"accuracy,omitempty"`
}

// ParsedSensorData 表示解析后的传感器数据
type ParsedSensorData struct {
	MessageID      int64
	SessionID      string
	DeviceID       string
	TotalReadings  int
	SensorTypes    []string
	SensorCounts   map[string]int
	TimeRange      TimeRange
	ParsedReadings []HumanReadableSensorData
	ReceivedAt     time.Time
}

// TimeRange 表示时间范围
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// HumanReadableSensorData 表示人类可读的传感器数据
type HumanReadableSensorData struct {
	SensorType   string
	Timestamp    time.Time
	ReadableTime string
	Values       []SensorValue
	Accuracy     string
}

// SensorValue 表示传感器值
type SensorValue struct {
	Name        string
	Value       string
	Unit        string
	Description string
}

// DashboardData 表示仪表板页面的数据
type DashboardData struct {
	HasData         bool
	TotalMessages   int
	TotalReadings   int
	SensorTypeCount int
	DeviceCount     int
	LatestData      []HumanReadableSensorData
}


