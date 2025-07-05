package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// 初始化测试环境
func init() {
	// 初始化应用配置
	AppConfig = Config{
		ServerPort:    "18000",
		DataDir:       "./test_data",
		EnableFileLog: true,
		EnableLogging: true,
		LogLevel:      "info",
		Environment:   "dev",
	}

	// 初始化Logger
	logConfig := LogConfig{
		Level:       slog.LevelInfo,
		Environment: "test",
		EnableFile:  false,
		AddSource:   false,
	}

	if err := InitLogger(logConfig); err != nil {
		panic("初始化Logger失败: " + err.Error())
	}

	// 确保测试数据目录存在
	os.MkdirAll(AppConfig.DataDir, 0755)
}

// TestHandleSensorData 测试传感器数据接收处理程序
func TestHandleSensorData(t *testing.T) {
	// 准备测试数据
	testData := map[string]interface{}{
		"messageId": 1,
		"sessionId": "test-session-123",
		"deviceId":  "test-device-456",
		"payload": []map[string]interface{}{
			{
				"name":     "accelerometer",
				"time":     1751729987437545000,
				"accuracy": 3,
				"values": map[string]interface{}{
					"x": -0.032849,
					"y": -0.004899,
					"z": 0.089095,
				},
			},
		},
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("JSON序列化失败: %v", err)
	}

	// 创建HTTP请求
	req := httptest.NewRequest("POST", "/data", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理程序
	handleSensorData(rr, req)

	// 验证响应状态码
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("期望状态码200，实际为%d", status)
	}

	// 验证响应内容
	expected := "数据接收成功"
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("响应内容不包含期望的文本: %s，实际响应: %s", expected, rr.Body.String())
	}
}

// TestHandleSensorDataInvalidJSON 测试无效JSON数据处理
func TestHandleSensorDataInvalidJSON(t *testing.T) {
	// 无效的JSON数据
	invalidJSON := `{"messageId": 1, "invalid": }`

	req := httptest.NewRequest("POST", "/data", strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handleSensorData(rr, req)

	// 验证返回状态码（解析失败应该返回400）
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("期望状态码400，实际为%d", status)
	}
}

// TestHandleSensorDataWrongMethod 测试错误的HTTP方法
func TestHandleSensorDataWrongMethod(t *testing.T) {
	req := httptest.NewRequest("GET", "/data", nil)
	rr := httptest.NewRecorder()

	handleSensorData(rr, req)

	// 验证返回方法不允许状态码
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("期望状态码405，实际为%d", status)
	}
}

// TestHandleDashboard 测试仪表板处理程序
func TestHandleDashboard(t *testing.T) {
	req := httptest.NewRequest("GET", "/dashboard", nil)
	rr := httptest.NewRecorder()

	handleDashboard(rr, req)

	// 验证响应状态码
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("期望状态码200，实际为%d", status)
	}

	// 验证响应内容类型
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("期望Content-Type为text/html; charset=utf-8，实际为%s", ct)
	}

	// 验证HTML内容包含关键元素
	body := rr.Body.String()
	expectedElements := []string{
		"传感器数据仪表板",
		"总消息数",
		"设备数量",
		"传感器类型",
	}

	for _, element := range expectedElements {
		if !strings.Contains(body, element) {
			t.Errorf("HTML内容不包含期望的元素: %s", element)
		}
	}
}

// TestHandleAPIData 测试API数据处理程序
func TestHandleAPIData(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/data", nil)
	rr := httptest.NewRecorder()

	handleAPIData(rr, req)

	// 验证响应状态码
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("期望状态码200，实际为%d", status)
	}

	// 验证响应内容类型
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("期望Content-Type为application/json，实际为%s", ct)
	}

	// 验证JSON响应格式
	var response []ParsedSensorData
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("JSON解析失败: %v", err)
	}

	// 至少应该返回一个空数组
	if response == nil {
		t.Error("响应不应该为nil")
	}
}

// TestHandleRoot 测试根路径处理程序
func TestHandleRoot(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handleRoot(rr, req)

	// 验证响应状态码
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("期望状态码200，实际为%d", status)
	}

	// 验证响应内容包含欢迎信息
	body := rr.Body.String()
	if !strings.Contains(body, "传感器日志服务器") {
		t.Error("根路径响应不包含欢迎信息")
	}
}

// TestConcurrentSensorDataHandling 测试并发数据处理
func TestConcurrentSensorDataHandling(t *testing.T) {
	// 并发发送多个请求
	concurrency := 5
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			testData := map[string]interface{}{
				"messageId": id,
				"sessionId": "concurrent-session",
				"deviceId":  "concurrent-device",
				"payload": []map[string]interface{}{
					{
						"name":     "accelerometer",
						"time":     1751729987437545000,
						"accuracy": 3,
						"values": map[string]interface{}{
							"x": float64(id),
							"y": float64(id * 2),
							"z": float64(id * 3),
						},
					},
				},
			}

			jsonData, _ := json.Marshal(testData)
			req := httptest.NewRequest("POST", "/data", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handleSensorData(rr, req)

			// 验证响应
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("并发请求%d失败，状态码: %d", id, status)
			}

			done <- true
		}(i)
	}

	// 等待所有请求完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	t.Logf("并发测试完成，处理了%d个请求", concurrency)
}
