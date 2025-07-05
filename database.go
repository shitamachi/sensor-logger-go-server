package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB客户端和集合
var (
	mongoClient    *mongo.Client
	sensorDataColl *mongo.Collection
	deviceInfoColl *mongo.Collection
)

// SensorMessageDocument MongoDB中的传感器消息文档结构（整个消息作为一个文档）
type SensorMessageDocument struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	MessageID   int64              `bson:"messageId"`
	SessionID   string             `bson:"sessionId"`
	DeviceID    string             `bson:"deviceId"`
	Payload     []SensorReading    `bson:"payload"`
	ReceivedAt  time.Time          `bson:"receivedAt"`
	ProcessedAt time.Time          `bson:"processedAt"`

	// 解析后的统计信息
	TotalReadings int            `bson:"totalReadings"`
	SensorTypes   []string       `bson:"sensorTypes"`
	SensorCounts  map[string]int `bson:"sensorCounts"`
	TimeRange     TimeRange      `bson:"timeRange"`

	// 解析后的可读数据
	ParsedReadings []HumanReadableSensorData `bson:"parsedReadings"`
}

// DeviceInfoDocument 设备信息文档结构
type DeviceInfoDocument struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	DeviceID      string             `bson:"deviceId"`
	FirstSeen     time.Time          `bson:"firstSeen"`
	LastSeen      time.Time          `bson:"lastSeen"`
	TotalMessages int64              `bson:"totalMessages"`
	TotalRecords  int64              `bson:"totalRecords"`
	SensorTypes   []string           `bson:"sensorTypes"`
	Sessions      []string           `bson:"sessions"`
}

// InitMongoDB 初始化MongoDB连接
func InitMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(AppConfig.MongoTimeout)*time.Second)
	defer cancel()

	// 创建MongoDB客户端
	clientOptions := options.Client().ApplyURI(AppConfig.MongoURI)

	var err error
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %v", err)
	}

	// 测试连接
	if err = mongoClient.Ping(ctx, nil); err != nil {
		return fmt.Errorf("MongoDB连接测试失败: %v", err)
	}

	// 获取数据库和集合
	db := mongoClient.Database(AppConfig.MongoDatabase)
	sensorDataColl = db.Collection("sensor_messages") // 改名为sensor_messages更合适
	deviceInfoColl = db.Collection("device_info")

	// 创建索引
	if err = createIndexes(); err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}

	Logger.Info("MongoDB连接成功",
		slog.String("uri", AppConfig.MongoURI),
		slog.String("database", AppConfig.MongoDatabase))
	return nil
}

// createIndexes 创建数据库索引
func createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 传感器消息索引
	messageIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "deviceId", Value: 1},
				{Key: "receivedAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "sessionId", Value: 1},
				{Key: "messageId", Value: 1},
			},
			Options: options.Index().SetUnique(true), // 确保同一会话中的消息ID唯一
		},
		{
			Keys: bson.D{
				{Key: "receivedAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "sensorTypes", Value: 1},
				{Key: "receivedAt", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "deviceId", Value: 1},
				{Key: "sessionId", Value: 1},
			},
		},
	}

	if _, err := sensorDataColl.Indexes().CreateMany(ctx, messageIndexes); err != nil {
		return fmt.Errorf("创建传感器消息索引失败: %v", err)
	}

	// 设备信息索引
	deviceIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "deviceId", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "lastSeen", Value: -1},
			},
		},
	}

	if _, err := deviceInfoColl.Indexes().CreateMany(ctx, deviceIndexes); err != nil {
		return fmt.Errorf("创建设备信息索引失败: %v", err)
	}

	Logger.Debug("数据库索引创建成功")
	return nil
}

// SaveSensorData 保存传感器数据到MongoDB（整个消息作为一个文档）
func SaveSensorData(parsedData *ParsedSensorData) error {
	if mongoClient == nil {
		return fmt.Errorf("MongoDB未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建传感器消息文档
	messageDoc := SensorMessageDocument{
		MessageID:      parsedData.MessageID,
		SessionID:      parsedData.SessionID,
		DeviceID:       parsedData.DeviceID,
		Payload:        extractOriginalPayload(parsedData),
		ReceivedAt:     parsedData.ReceivedAt,
		ProcessedAt:    time.Now(),
		TotalReadings:  parsedData.TotalReadings,
		SensorTypes:    parsedData.SensorTypes,
		SensorCounts:   parsedData.SensorCounts,
		TimeRange:      parsedData.TimeRange,
		ParsedReadings: parsedData.ParsedReadings,
	}

	// 插入传感器消息文档
	result, err := sensorDataColl.InsertOne(ctx, messageDoc)
	if err != nil {
		return fmt.Errorf("保存传感器消息失败: %v", err)
	}

	Logger.Debug("传感器消息保存成功",
		slog.String("document_id", result.InsertedID.(primitive.ObjectID).Hex()),
		slog.String("device_id", parsedData.DeviceID),
		slog.Int64("message_id", parsedData.MessageID),
		slog.Int("readings_count", parsedData.TotalReadings))

	// 更新设备信息
	if err := updateDeviceInfo(parsedData); err != nil {
		Logger.Error("更新设备信息失败",
			slog.String("error", err.Error()),
			slog.String("device_id", parsedData.DeviceID))
	}

	return nil
}

// extractOriginalPayload 从解析后的数据中提取原始payload结构
func extractOriginalPayload(parsedData *ParsedSensorData) []SensorReading {
	// 这里我们需要重新构造原始的SensorReading结构
	// 由于我们已经有了ParsedReadings，我们可以从中提取信息
	payload := make([]SensorReading, 0, len(parsedData.ParsedReadings))

	for _, reading := range parsedData.ParsedReadings {
		// 重新构造values map
		values := make(map[string]interface{})
		for _, value := range reading.Values {
			// 尝试解析回原始格式
			switch value.Name {
			case "X轴加速度", "Y轴加速度", "Z轴加速度":
				axisName := string(value.Name[0]) // 取第一个字符
				values[strings.ToLower(axisName)] = parseFloat(value.Value)
			case "X轴角速度", "Y轴角速度", "Z轴角速度":
				axisName := string(value.Name[0])
				values[strings.ToLower(axisName)] = parseFloat(value.Value)
			case "X轴磁场", "Y轴磁场", "Z轴磁场":
				axisName := string(value.Name[0])
				values[strings.ToLower(axisName)] = parseFloat(value.Value)
			case "磁方位角":
				values["magneticBearing"] = parseFloat(value.Value)
			case "指南针方位":
				values["magneticBearing"] = parseFloat(value.Value)
			case "步数":
				values["steps"] = parseFloat(value.Value)
			default:
				// 对于其他类型，使用通用方法
				values[value.Name] = parseFloat(value.Value)
			}
		}

		sensorReading := SensorReading{
			Name:     reading.SensorType,
			Time:     reading.Timestamp.UnixNano(),
			Values:   values,
			Accuracy: getAccuracyInt(reading.Accuracy),
		}

		payload = append(payload, sensorReading)
	}

	return payload
}

// parseFloat 安全地解析字符串为float64
func parseFloat(s string) float64 {
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return 0.0
}

// updateDeviceInfo 更新设备信息
func updateDeviceInfo(parsedData *ParsedSensorData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"deviceId": parsedData.DeviceID}

	// 检查设备是否已存在
	var existingDevice DeviceInfoDocument
	err := deviceInfoColl.FindOne(ctx, filter).Decode(&existingDevice)

	if err == mongo.ErrNoDocuments {
		// 创建新设备记录
		newDevice := DeviceInfoDocument{
			DeviceID:      parsedData.DeviceID,
			FirstSeen:     parsedData.ReceivedAt,
			LastSeen:      parsedData.ReceivedAt,
			TotalMessages: 1,
			TotalRecords:  int64(parsedData.TotalReadings),
			SensorTypes:   parsedData.SensorTypes,
			Sessions:      []string{parsedData.SessionID},
		}

		_, err = deviceInfoColl.InsertOne(ctx, newDevice)
		if err != nil {
			return fmt.Errorf("创建设备信息失败: %v", err)
		}
		Logger.Info("创建新设备记录", slog.String("device_id", parsedData.DeviceID))
	} else if err == nil {
		// 更新现有设备记录
		update := bson.M{
			"$set": bson.M{
				"lastSeen": parsedData.ReceivedAt,
			},
			"$inc": bson.M{
				"totalMessages": 1,
				"totalRecords":  int64(parsedData.TotalReadings),
			},
			"$addToSet": bson.M{
				"sensorTypes": bson.M{"$each": parsedData.SensorTypes},
				"sessions":    parsedData.SessionID,
			},
		}

		_, err = deviceInfoColl.UpdateOne(ctx, filter, update)
		if err != nil {
			return fmt.Errorf("更新设备信息失败: %v", err)
		}
		Logger.Debug("设备信息更新成功", slog.String("device_id", parsedData.DeviceID))
	} else {
		return fmt.Errorf("查询设备信息失败: %v", err)
	}

	return nil
}

// getAccuracyInt 将精度描述转换为数字
func getAccuracyInt(accuracy string) int {
	switch accuracy {
	case "不可靠":
		return 0
	case "低精度":
		return 1
	case "中等精度":
		return 2
	case "高精度":
		return 3
	default:
		return -1
	}
}

// GetSensorDataFromDB 从数据库获取传感器消息
func GetSensorDataFromDB(limit int, deviceID string, sensorType string) ([]SensorMessageDocument, error) {
	if mongoClient == nil {
		return nil, fmt.Errorf("MongoDB未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建查询条件
	filter := bson.M{}
	if deviceID != "" {
		filter["deviceId"] = deviceID
	}
	if sensorType != "" {
		filter["sensorTypes"] = sensorType
	}

	// 设置查询选项
	opts := options.Find().
		SetSort(bson.D{{Key: "receivedAt", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := sensorDataColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询传感器消息失败: %v", err)
	}
	defer cursor.Close(ctx)

	var results []SensorMessageDocument
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析查询结果失败: %v", err)
	}

	Logger.Debug("数据库查询完成",
		slog.Int("count", len(results)),
		slog.String("device", deviceID),
		slog.String("sensor", sensorType))

	return results, nil
}

// GetDeviceInfo 获取设备信息
func GetDeviceInfo() ([]DeviceInfoDocument, error) {
	if mongoClient == nil {
		return nil, fmt.Errorf("MongoDB未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "lastSeen", Value: -1}})
	cursor, err := deviceInfoColl.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("查询设备信息失败: %v", err)
	}
	defer cursor.Close(ctx)

	var results []DeviceInfoDocument
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("解析设备信息失败: %v", err)
	}

	Logger.Debug("设备信息查询完成", slog.Int("count", len(results)))
	return results, nil
}

// GetDashboardStats 获取仪表板统计信息
func GetDashboardStats() (map[string]interface{}, error) {
	if mongoClient == nil {
		return nil, fmt.Errorf("MongoDB未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	stats := make(map[string]interface{})

	// 总消息数
	totalMessages, err := sensorDataColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("查询总消息数失败: %v", err)
	}
	stats["totalMessages"] = totalMessages

	// 总记录数（所有消息中的传感器读数总和）
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":          nil,
			"totalRecords": bson.M{"$sum": "$totalReadings"},
		}},
	}
	cursor, err := sensorDataColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("统计总记录数失败: %v", err)
	}
	defer cursor.Close(ctx)

	var totalRecordsResult []bson.M
	if err = cursor.All(ctx, &totalRecordsResult); err == nil && len(totalRecordsResult) > 0 {
		stats["totalRecords"] = totalRecordsResult[0]["totalRecords"]
	} else {
		stats["totalRecords"] = 0
	}

	// 设备数量
	deviceCount, err := deviceInfoColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("查询设备数量失败: %v", err)
	}
	stats["deviceCount"] = deviceCount

	// 传感器类型数量
	sensorTypes, err := sensorDataColl.Distinct(ctx, "sensorTypes", bson.M{})
	if err != nil {
		return nil, fmt.Errorf("查询传感器类型失败: %v", err)
	}
	// 去重传感器类型
	uniqueSensorTypes := make(map[string]bool)
	for _, sensorType := range sensorTypes {
		if typeArray, ok := sensorType.(bson.A); ok {
			for _, t := range typeArray {
				if typeStr, ok := t.(string); ok {
					uniqueSensorTypes[typeStr] = true
				}
			}
		}
	}
	sensorTypesList := make([]string, 0, len(uniqueSensorTypes))
	for sensorType := range uniqueSensorTypes {
		sensorTypesList = append(sensorTypesList, sensorType)
	}
	stats["sensorTypeCount"] = len(sensorTypesList)
	stats["sensorTypes"] = sensorTypesList

	// 最新数据时间
	var latestMessage SensorMessageDocument
	opts := options.FindOne().SetSort(bson.D{{Key: "receivedAt", Value: -1}})
	err = sensorDataColl.FindOne(ctx, bson.M{}, opts).Decode(&latestMessage)
	if err == nil {
		stats["latestDataTime"] = latestMessage.ReceivedAt
	}

	Logger.Debug("统计信息查询完成",
		slog.Int64("total_messages", totalMessages),
		slog.Int64("device_count", deviceCount),
		slog.Int("sensor_types", len(sensorTypesList)))

	return stats, nil
}

// CloseMongoDB 关闭MongoDB连接
func CloseMongoDB() error {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := mongoClient.Disconnect(ctx); err != nil {
			return fmt.Errorf("关闭MongoDB连接失败: %v", err)
		}

		Logger.Info("MongoDB连接已关闭")
	}
	return nil
}
