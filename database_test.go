package main

import (
	"testing"
	"time"
)

func TestGetAccuracyInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"不可靠", 0},
		{"低精度", 1},
		{"中等精度", 2},
		{"高精度", 3},
		{"未知", -1},
		{"", -1},
	}

	for _, test := range tests {
		result := getAccuracyInt(test.input)
		if result != test.expected {
			t.Errorf("getAccuracyInt(%s) = %d, 期望 %d", test.input, result, test.expected)
		}
	}
}

func TestSensorDataDocument(t *testing.T) {
	// 测试SensorDataDocument结构
	doc := SensorDataDocument{
		MessageID:  123,
		SessionID:  "test-session",
		DeviceID:   "test-device",
		SensorType: "accelerometer",
		Timestamp:  time.Now(),
		ReceivedAt: time.Now(),
		Accuracy:   3,
	}

	if doc.MessageID != 123 {
		t.Errorf("期望MessageID为123，实际为%d", doc.MessageID)
	}

	if doc.SensorType != "accelerometer" {
		t.Errorf("期望SensorType为accelerometer，实际为%s", doc.SensorType)
	}
}

func TestDeviceInfoDocument(t *testing.T) {
	// 测试DeviceInfoDocument结构
	now := time.Now()
	doc := DeviceInfoDocument{
		DeviceID:     "test-device",
		FirstSeen:    now,
		LastSeen:     now,
		TotalRecords: 100,
		SensorTypes:  []string{"accelerometer", "gyroscope"},
		Sessions:     []string{"session1", "session2"},
	}

	if doc.DeviceID != "test-device" {
		t.Errorf("期望DeviceID为test-device，实际为%s", doc.DeviceID)
	}

	if len(doc.SensorTypes) != 2 {
		t.Errorf("期望SensorTypes长度为2，实际为%d", len(doc.SensorTypes))
	}

	if doc.TotalRecords != 100 {
		t.Errorf("期望TotalRecords为100，实际为%d", doc.TotalRecords)
	}
}

// 注意：这些测试不需要实际的MongoDB连接
// 实际的数据库操作测试需要在集成测试中进行
func TestMongoDBFunctionsWithoutConnection(t *testing.T) {
	// 测试在没有MongoDB连接时的错误处理

	// 重置MongoDB客户端为nil
	originalClient := mongoClient
	mongoClient = nil
	defer func() {
		mongoClient = originalClient
	}()

	// 测试SaveSensorData
	testData := &ParsedSensorData{
		MessageID: 1,
		DeviceID:  "test",
		SessionID: "test",
	}

	err := SaveSensorData(testData)
	if err == nil {
		t.Error("期望SaveSensorData在没有MongoDB连接时返回错误")
	}

	// 测试GetSensorDataFromDB
	_, err = GetSensorDataFromDB(10, "", "")
	if err == nil {
		t.Error("期望GetSensorDataFromDB在没有MongoDB连接时返回错误")
	}

	// 测试GetDeviceInfo
	_, err = GetDeviceInfo()
	if err == nil {
		t.Error("期望GetDeviceInfo在没有MongoDB连接时返回错误")
	}

	// 测试GetDashboardStats
	_, err = GetDashboardStats()
	if err == nil {
		t.Error("期望GetDashboardStats在没有MongoDB连接时返回错误")
	}
}
