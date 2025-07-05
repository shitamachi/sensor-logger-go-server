package main

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

// TestParseSensorMessage 测试传感器消息解析功能
func TestParseSensorMessage(t *testing.T) {
	// 使用真实的传感器数据进行测试
	testData := `{
		"messageId": 27,
		"sessionId": "f08abd14-4a86-4913-8c33-f75269e862b9",
		"deviceId": "0e35011f-e2fe-482e-b9a1-1625bb039f37",
		"payload": [
			{
				"name": "accelerometer",
				"values": {
					"z": 0.0890951156616211,
					"y": -0.004899024963378906,
					"x": -0.03284934163093567
				},
				"accuracy": 3,
				"time": 1751729987437545000
			},
			{
				"name": "gyroscope",
				"values": {
					"z": 0.016910573467612267,
					"y": -0.0319569893181324,
					"x": -0.05113118514418602
				},
				"accuracy": 3,
				"time": 1751729987477628400
			},
			{
				"name": "magnetometer",
				"values": {
					"magneticBearing": 137.27661523593756
				},
				"accuracy": 3,
				"time": 1751729987486290400
			}
		]
	}`

	parsed, err := parseSensorMessage([]byte(testData))
	if err != nil {
		t.Fatalf("解析传感器数据失败: %v", err)
	}

	// 验证基本信息
	if parsed.MessageID != 27 {
		t.Errorf("期望消息ID为27，实际为%d", parsed.MessageID)
	}
	if parsed.DeviceID != "0e35011f-e2fe-482e-b9a1-1625bb039f37" {
		t.Errorf("设备ID不匹配")
	}
	if parsed.TotalReadings != 3 {
		t.Errorf("期望总读数为3，实际为%d", parsed.TotalReadings)
	}

	// 验证传感器类型
	expectedSensors := []string{"accelerometer", "gyroscope", "magnetometer"}
	if len(parsed.SensorTypes) != len(expectedSensors) {
		t.Errorf("期望传感器类型数量为%d，实际为%d", len(expectedSensors), len(parsed.SensorTypes))
	}

	// 验证解析后的数据
	if len(parsed.ParsedReadings) != 3 {
		t.Errorf("期望解析后读数为3，实际为%d", len(parsed.ParsedReadings))
	}

	// 验证加速度计数据解析
	accelReading := parsed.ParsedReadings[0]
	if accelReading.SensorType != "accelerometer" {
		t.Errorf("期望传感器类型为accelerometer，实际为%s", accelReading.SensorType)
	}
	if accelReading.Accuracy != "高精度" {
		t.Errorf("期望精度为高精度，实际为%s", accelReading.Accuracy)
	}
	if len(accelReading.Values) != 3 {
		t.Errorf("期望加速度计有3个值，实际为%d", len(accelReading.Values))
	}

	// 验证数据单位
	for _, value := range accelReading.Values {
		if value.Unit != "m/s²" {
			t.Errorf("期望加速度计单位为m/s²，实际为%s", value.Unit)
		}
	}
}

// TestParseAccelerometer 测试加速度计数据解析
func TestParseAccelerometer(t *testing.T) {
	values := map[string]interface{}{
		"x": -0.032849,
		"y": -0.004899,
		"z": 0.089095,
	}

	result := parseAccelerometer(values)

	if len(result) != 3 {
		t.Errorf("期望3个值，实际为%d", len(result))
	}

	expectedNames := []string{"X轴加速度", "Y轴加速度", "Z轴加速度"}
	for i, expected := range expectedNames {
		if result[i].Name != expected {
			t.Errorf("期望名称为%s，实际为%s", expected, result[i].Name)
		}
		if result[i].Unit != "m/s²" {
			t.Errorf("期望单位为m/s²，实际为%s", result[i].Unit)
		}
	}
}

// TestParseGyroscope 测试陀螺仪数据解析
func TestParseGyroscope(t *testing.T) {
	values := map[string]interface{}{
		"x": -0.051131,
		"y": -0.031957,
		"z": 0.016911,
	}

	result := parseGyroscope(values)

	if len(result) != 3 {
		t.Errorf("期望3个值，实际为%d", len(result))
	}

	for _, value := range result {
		if value.Unit != "rad/s" {
			t.Errorf("期望单位为rad/s，实际为%s", value.Unit)
		}
		if !strings.Contains(value.Name, "角速度") {
			t.Errorf("期望名称包含'角速度'，实际为%s", value.Name)
		}
	}
}

// TestParseMagnetometer 测试磁力计数据解析
func TestParseMagnetometer(t *testing.T) {
	values := map[string]interface{}{
		"magneticBearing": 137.27661523593756,
	}

	result := parseMagnetometer(values)

	if len(result) != 1 {
		t.Errorf("期望1个值，实际为%d", len(result))
	}

	if result[0].Name != "磁方位角" {
		t.Errorf("期望名称为磁方位角，实际为%s", result[0].Name)
	}
	if result[0].Unit != "度" {
		t.Errorf("期望单位为度，实际为%s", result[0].Unit)
	}
}

// TestGetAccuracyDescription 测试精度描述功能
func TestGetAccuracyDescription(t *testing.T) {
	tests := []struct {
		accuracy int
		expected string
	}{
		{0, "不可靠"},
		{1, "低精度"},
		{2, "中等精度"},
		{3, "高精度"},
		{99, "未知"},
	}

	for _, test := range tests {
		result := getAccuracyDescription(test.accuracy)
		if result != test.expected {
			t.Errorf("精度%d期望描述为%s，实际为%s", test.accuracy, test.expected, result)
		}
	}
}

// TestGetFloat64 测试类型转换功能
func TestGetFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
	}{
		{float64(1.23), 1.23},
		{float32(1.23), float64(float32(1.23))}, // float32精度问题
		{int(123), 123.0},
		{int64(123), 123.0},
		{"invalid", 0.0},
		{nil, 0.0},
	}

	for _, test := range tests {
		result := getFloat64(test.input)
		// 使用适当的精度比较浮点数
		if math.Abs(result-test.expected) > 1e-9 {
			t.Errorf("输入%v期望结果为%f，实际为%f", test.input, test.expected, result)
		}
	}
}

// TestParseToHumanReadable 测试人类可读格式转换
func TestParseToHumanReadable(t *testing.T) {
	reading := SensorReading{
		Name:     "accelerometer",
		Time:     1751729987437545000,
		Accuracy: 3,
		Values: map[string]interface{}{
			"x": -0.032849,
			"y": -0.004899,
			"z": 0.089095,
		},
	}

	result := parseToHumanReadable(reading)

	if result.SensorType != "accelerometer" {
		t.Errorf("期望传感器类型为accelerometer，实际为%s", result.SensorType)
	}
	if result.Accuracy != "高精度" {
		t.Errorf("期望精度为高精度，实际为%s", result.Accuracy)
	}
	if len(result.Values) != 3 {
		t.Errorf("期望3个值，实际为%d", len(result.Values))
	}

	// 验证时间格式
	expectedTime := time.Unix(0, 1751729987437545000)
	if !result.Timestamp.Equal(expectedTime) {
		t.Errorf("时间戳不匹配")
	}
}

// TestParseRealSensorData 测试真实传感器数据文件解析
func TestParseRealSensorData(t *testing.T) {
	// 测试小文件
	testFile := "testdata/sensor_data_small.json"
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Skipf("跳过测试，文件不存在: %s", testFile)
		return
	}

	parsed, err := parseSensorMessage(data)
	if err != nil {
		t.Fatalf("解析真实数据失败: %v", err)
	}

	// 验证基本结构
	if parsed.MessageID == 0 {
		t.Error("消息ID不应该为0")
	}
	if parsed.DeviceID == "" {
		t.Error("设备ID不应该为空")
	}
	if parsed.TotalReadings == 0 {
		t.Error("总读数不应该为0")
	}
	if len(parsed.ParsedReadings) == 0 {
		t.Error("解析后的读数不应该为空")
	}

	t.Logf("成功解析真实数据: 消息ID=%d, 设备ID=%s, 总读数=%d, 传感器类型=%v",
		parsed.MessageID, parsed.DeviceID, parsed.TotalReadings, parsed.SensorTypes)
}

// TestParseMediumSensorData 测试中型传感器数据文件解析
func TestParseMediumSensorData(t *testing.T) {
	testFile := "testdata/sensor_data_medium.json"
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Skipf("跳过测试，文件不存在: %s", testFile)
		return
	}

	parsed, err := parseSensorMessage(data)
	if err != nil {
		t.Fatalf("解析中型数据失败: %v", err)
	}

	// 验证基本结构
	if parsed.MessageID != 2 {
		t.Errorf("期望消息ID为2，实际为%d", parsed.MessageID)
	}
	if parsed.DeviceID != "mock-device-12345" {
		t.Errorf("期望设备ID为mock-device-12345，实际为%s", parsed.DeviceID)
	}
	if parsed.TotalReadings != 8 {
		t.Errorf("期望总读数为8，实际为%d", parsed.TotalReadings)
	}

	// 验证传感器类型
	expectedSensors := []string{"accelerometer", "gravity", "gyroscope", "magnetometer", "compass", "pedometer", "magnetometeruncalibrated", "orientation"}
	if len(parsed.SensorTypes) != len(expectedSensors) {
		t.Errorf("期望传感器类型数量为%d，实际为%d", len(expectedSensors), len(parsed.SensorTypes))
	}

	t.Logf("成功解析中型数据: 消息ID=%d, 设备ID=%s, 总读数=%d, 传感器类型=%v",
		parsed.MessageID, parsed.DeviceID, parsed.TotalReadings, parsed.SensorTypes)
}

// BenchmarkParseSensorMessage 性能基准测试
func BenchmarkParseSensorMessage(b *testing.B) {
	testData := `{
		"messageId": 1,
		"sessionId": "test-session",
		"deviceId": "test-device",
		"payload": [
			{
				"name": "accelerometer",
				"values": {"x": 1.0, "y": 2.0, "z": 3.0},
				"accuracy": 3,
				"time": 1751729987437545000
			}
		]
	}`

	data := []byte(testData)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parseSensorMessage(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestJSONMarshaling 测试JSON序列化
func TestJSONMarshaling(t *testing.T) {
	// 创建测试数据
	parsed := &ParsedSensorData{
		MessageID:     1,
		SessionID:     "test-session",
		DeviceID:      "test-device",
		TotalReadings: 1,
		SensorTypes:   []string{"accelerometer"},
		ReceivedAt:    time.Now(),
	}

	// 序列化
	data, err := json.Marshal(parsed)
	if err != nil {
		t.Fatalf("JSON序列化失败: %v", err)
	}

	// 验证结果不为空
	if len(data) == 0 {
		t.Error("序列化结果不应该为空")
	}

	t.Logf("序列化结果长度: %d bytes", len(data))
}

// TestLargeSensorDataFile 测试大型传感器数据文件
func TestLargeSensorDataFile(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大文件测试（使用 -short 标志）")
	}

	testFile := "testdata/sensor_data_large.json"
	file, err := os.Open(testFile)
	if err != nil {
		t.Skipf("跳过测试，文件不存在: %s", testFile)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}

	start := time.Now()
	parsed, err := parseSensorMessage(data)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("解析大型数据文件失败: %v", err)
	}

	t.Logf("大型文件解析性能: 文件大小=%d bytes, 耗时=%v, 读数=%d",
		len(data), duration, parsed.TotalReadings)

	// 验证解析结果
	if parsed.TotalReadings == 0 {
		t.Error("大型文件的总读数不应该为0")
	}
	if len(parsed.SensorTypes) == 0 {
		t.Error("大型文件应该包含传感器类型")
	}
}
