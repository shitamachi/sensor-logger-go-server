package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// parseSensorMessage 解析传感器消息
func parseSensorMessage(data []byte) (*ParsedSensorData, error) {
	var message SensorMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	parsed := &ParsedSensorData{
		MessageID:      message.MessageID,
		SessionID:      message.SessionID,
		DeviceID:       message.DeviceID,
		TotalReadings:  len(message.Payload),
		SensorCounts:   make(map[string]int),
		ParsedReadings: make([]HumanReadableSensorData, 0),
		ReceivedAt:     time.Now(),
	}

	// 统计传感器类型
	sensorTypeSet := make(map[string]bool)
	var minTime, maxTime int64

	for i, reading := range message.Payload {
		sensorTypeSet[reading.Name] = true
		parsed.SensorCounts[reading.Name]++

		// 计算时间范围
		if i == 0 {
			minTime = reading.Time
			maxTime = reading.Time
		} else {
			if reading.Time < minTime {
				minTime = reading.Time
			}
			if reading.Time > maxTime {
				maxTime = reading.Time
			}
		}

		// 解析为人类可读格式
		humanReadable := parseToHumanReadable(reading)
		parsed.ParsedReadings = append(parsed.ParsedReadings, humanReadable)
	}

	// 设置传感器类型列表
	for sensorType := range sensorTypeSet {
		parsed.SensorTypes = append(parsed.SensorTypes, sensorType)
	}
	sort.Strings(parsed.SensorTypes)

	// 设置时间范围
	parsed.TimeRange = TimeRange{
		Start: time.Unix(0, minTime),
		End:   time.Unix(0, maxTime),
	}

	return parsed, nil
}

// parseToHumanReadable 将传感器读数转换为人类可读格式
func parseToHumanReadable(reading SensorReading) HumanReadableSensorData {
	timestamp := time.Unix(0, reading.Time)

	result := HumanReadableSensorData{
		SensorType:   reading.Name,
		Timestamp:    timestamp,
		ReadableTime: timestamp.Format("2006-01-02 15:04:05.000"),
		Values:       make([]SensorValue, 0),
		Accuracy:     getAccuracyDescription(reading.Accuracy),
	}

	// 根据传感器类型解析值
	switch strings.ToLower(reading.Name) {
	case "accelerometer":
		result.Values = parseAccelerometer(reading.Values)
	case "gyroscope":
		result.Values = parseGyroscope(reading.Values)
	case "magnetometer":
		result.Values = parseMagnetometer(reading.Values)
	case "gravity":
		result.Values = parseGravity(reading.Values)
	case "orientation":
		result.Values = parseOrientation(reading.Values)
	case "compass":
		result.Values = parseCompass(reading.Values)
	case "pedometer":
		result.Values = parsePedometer(reading.Values)
	case "magnetometeruncalibrated":
		result.Values = parseMagnetometerUncalibrated(reading.Values)
	case "location":
		result.Values = parseLocation(reading.Values)
	case "barometer":
		result.Values = parseBarometer(reading.Values)
	default:
		result.Values = parseGeneric(reading.Values)
	}

	return result
}

// parseAccelerometer 解析加速度计数据
func parseAccelerometer(values map[string]interface{}) []SensorValue {
	return []SensorValue{
		{Name: "X轴加速度", Value: fmt.Sprintf("%.6f", getFloat64(values["x"])), Unit: "m/s²", Description: "X轴方向的加速度"},
		{Name: "Y轴加速度", Value: fmt.Sprintf("%.6f", getFloat64(values["y"])), Unit: "m/s²", Description: "Y轴方向的加速度"},
		{Name: "Z轴加速度", Value: fmt.Sprintf("%.6f", getFloat64(values["z"])), Unit: "m/s²", Description: "Z轴方向的加速度"},
	}
}

// parseGyroscope 解析陀螺仪数据
func parseGyroscope(values map[string]interface{}) []SensorValue {
	return []SensorValue{
		{Name: "X轴角速度", Value: fmt.Sprintf("%.6f", getFloat64(values["x"])), Unit: "rad/s", Description: "绕X轴的角速度"},
		{Name: "Y轴角速度", Value: fmt.Sprintf("%.6f", getFloat64(values["y"])), Unit: "rad/s", Description: "绕Y轴的角速度"},
		{Name: "Z轴角速度", Value: fmt.Sprintf("%.6f", getFloat64(values["z"])), Unit: "rad/s", Description: "绕Z轴的角速度"},
	}
}

// parseMagnetometer 解析磁力计数据
func parseMagnetometer(values map[string]interface{}) []SensorValue {
	if bearing, ok := values["magneticBearing"]; ok {
		return []SensorValue{
			{Name: "磁方位角", Value: fmt.Sprintf("%.2f", getFloat64(bearing)), Unit: "度", Description: "相对于磁北的方位角"},
		}
	}
	return []SensorValue{
		{Name: "X轴磁场", Value: fmt.Sprintf("%.6f", getFloat64(values["x"])), Unit: "μT", Description: "X轴方向的磁场强度"},
		{Name: "Y轴磁场", Value: fmt.Sprintf("%.6f", getFloat64(values["y"])), Unit: "μT", Description: "Y轴方向的磁场强度"},
		{Name: "Z轴磁场", Value: fmt.Sprintf("%.6f", getFloat64(values["z"])), Unit: "μT", Description: "Z轴方向的磁场强度"},
	}
}

// parseGravity 解析重力传感器数据
func parseGravity(values map[string]interface{}) []SensorValue {
	return []SensorValue{
		{Name: "X轴重力", Value: fmt.Sprintf("%.6f", getFloat64(values["x"])), Unit: "m/s²", Description: "X轴方向的重力分量"},
		{Name: "Y轴重力", Value: fmt.Sprintf("%.6f", getFloat64(values["y"])), Unit: "m/s²", Description: "Y轴方向的重力分量"},
		{Name: "Z轴重力", Value: fmt.Sprintf("%.6f", getFloat64(values["z"])), Unit: "m/s²", Description: "Z轴方向的重力分量"},
	}
}

// parseOrientation 解析方向传感器数据
func parseOrientation(values map[string]interface{}) []SensorValue {
	result := make([]SensorValue, 0)
	if qw, ok := values["qw"]; ok {
		result = append(result, SensorValue{Name: "四元数W", Value: fmt.Sprintf("%.6f", getFloat64(qw)), Unit: "", Description: "四元数W分量"})
	}
	if qx, ok := values["qx"]; ok {
		result = append(result, SensorValue{Name: "四元数X", Value: fmt.Sprintf("%.6f", getFloat64(qx)), Unit: "", Description: "四元数X分量"})
	}
	if qy, ok := values["qy"]; ok {
		result = append(result, SensorValue{Name: "四元数Y", Value: fmt.Sprintf("%.6f", getFloat64(qy)), Unit: "", Description: "四元数Y分量"})
	}
	if qz, ok := values["qz"]; ok {
		result = append(result, SensorValue{Name: "四元数Z", Value: fmt.Sprintf("%.6f", getFloat64(qz)), Unit: "", Description: "四元数Z分量"})
	}
	return result
}

// parseCompass 解析指南针数据
func parseCompass(values map[string]interface{}) []SensorValue {
	return []SensorValue{
		{Name: "指南针方位", Value: fmt.Sprintf("%.2f", getFloat64(values["magneticBearing"])), Unit: "度", Description: "指南针方位角"},
	}
}

// parsePedometer 解析计步器数据
func parsePedometer(values map[string]interface{}) []SensorValue {
	return []SensorValue{
		{Name: "步数", Value: fmt.Sprintf("%.0f", getFloat64(values["steps"])), Unit: "步", Description: "累计步数"},
	}
}

// parseMagnetometerUncalibrated 解析未校准磁力计数据
func parseMagnetometerUncalibrated(values map[string]interface{}) []SensorValue {
	return []SensorValue{
		{Name: "X轴磁场(未校准)", Value: fmt.Sprintf("%.6f", getFloat64(values["x"])), Unit: "μT", Description: "X轴方向的未校准磁场强度"},
		{Name: "Y轴磁场(未校准)", Value: fmt.Sprintf("%.6f", getFloat64(values["y"])), Unit: "μT", Description: "Y轴方向的未校准磁场强度"},
		{Name: "Z轴磁场(未校准)", Value: fmt.Sprintf("%.6f", getFloat64(values["z"])), Unit: "μT", Description: "Z轴方向的未校准磁场强度"},
	}
}

// parseLocation 解析位置数据
func parseLocation(values map[string]interface{}) []SensorValue {
	result := make([]SensorValue, 0)
	if lat, ok := values["latitude"]; ok {
		result = append(result, SensorValue{Name: "纬度", Value: fmt.Sprintf("%.8f", getFloat64(lat)), Unit: "度", Description: "地理纬度"})
	}
	if lng, ok := values["longitude"]; ok {
		result = append(result, SensorValue{Name: "经度", Value: fmt.Sprintf("%.8f", getFloat64(lng)), Unit: "度", Description: "地理经度"})
	}
	if alt, ok := values["altitude"]; ok {
		result = append(result, SensorValue{Name: "海拔", Value: fmt.Sprintf("%.2f", getFloat64(alt)), Unit: "米", Description: "海拔高度"})
	}
	if speed, ok := values["speed"]; ok {
		result = append(result, SensorValue{Name: "速度", Value: fmt.Sprintf("%.2f", getFloat64(speed)), Unit: "m/s", Description: "移动速度"})
	}
	if bearing, ok := values["bearing"]; ok {
		result = append(result, SensorValue{Name: "方位角", Value: fmt.Sprintf("%.2f", getFloat64(bearing)), Unit: "度", Description: "移动方位角"})
	}
	return result
}

// parseBarometer 解析气压计数据
func parseBarometer(values map[string]interface{}) []SensorValue {
	result := make([]SensorValue, 0)
	if pressure, ok := values["pressure"]; ok {
		result = append(result, SensorValue{Name: "气压", Value: fmt.Sprintf("%.2f", getFloat64(pressure)), Unit: "hPa", Description: "大气压力"})
	}
	if altitude, ok := values["altitude"]; ok {
		result = append(result, SensorValue{Name: "气压高度", Value: fmt.Sprintf("%.2f", getFloat64(altitude)), Unit: "米", Description: "基于气压计算的高度"})
	}
	return result
}

// parseGeneric 解析通用传感器数据
func parseGeneric(values map[string]interface{}) []SensorValue {
	result := make([]SensorValue, 0)
	for key, value := range values {
		result = append(result, SensorValue{
			Name:        key,
			Value:       fmt.Sprintf("%v", value),
			Unit:        "",
			Description: fmt.Sprintf("%s数值", key),
		})
	}
	return result
}
