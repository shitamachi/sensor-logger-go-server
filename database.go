package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB客户端和集合
var (
	mongoClient     *mongo.Client
	sensorDataColl  *mongo.Collection
	deviceInfoColl  *mongo.Collection
)

// SensorDataDocument MongoDB中的传感器数据文档结构
type SensorDataDocument struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	MessageID      int64              `bson:"messageId"`
	SessionID      string             `bson:"sessionId"`
	DeviceID       string             `bson:"deviceId"`
	SensorType     string             `bson:"sensorType"`
	Timestamp      time.Time          `bson:"timestamp"`
	Values         bson.M             `bson:"values"`
	Accuracy       int                `bson:"accuracy"`
	ReceivedAt     time.Time          `bson:"receivedAt"`
	ProcessedData  bson.M             `bson:"processedData,omitempty"`
}

// DeviceInfoDocument 设备信息文档结构
type DeviceInfoDocument struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	DeviceID     string             `bson:"deviceId"`
	FirstSeen    time.Time          `bson:"firstSeen"`
	LastSeen     time.Time          `bson:"lastSeen"`
	TotalRecords int64              `bson:"totalRecords"`
	SensorTypes  []string           `bson:"sensorTypes"`
	Sessions     []string           `bson:"sessions"`
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
	sensorDataColl = db.Collection("sensor_data")
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

	// 传感器数据索引
	sensorIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "deviceId", Value: 1},
				{Key: "timestamp", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "sensorType", Value: 1},
				{Key: "timestamp", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "sessionId", Value: 1},
				{Key: "messageId", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "receivedAt", Value: -1},
			},
		},
	}

	if _, err := sensorDataColl.Indexes().CreateMany(ctx, sensorIndexes); err != nil {
		return fmt.Errorf("创建传感器数据索引失败: %v", err)
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

// SaveSensorData 保存传感器数据到MongoDB
func SaveSensorData(parsedData *ParsedSensorData) error {
	if mongoClient == nil {
		return fmt.Errorf("MongoDB未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 准备批量插入的文档
	var documents []interface{}
	
	for _, reading := range parsedData.ParsedReadings {
		// 转换Values为bson.M格式
		valuesMap := make(bson.M)
		for _, value := range reading.Values {
			valuesMap[value.Name] = bson.M{
				"value":       value.Value,
				"unit":        value.Unit,
				"description": value.Description,
			}
		}

		// 创建处理后的数据
		processedData := bson.M{
			"readableTime": reading.ReadableTime,
			"accuracy":     reading.Accuracy,
			"valueCount":   len(reading.Values),
		}

		doc := SensorDataDocument{
			MessageID:     parsedData.MessageID,
			SessionID:     parsedData.SessionID,
			DeviceID:      parsedData.DeviceID,
			SensorType:    reading.SensorType,
			Timestamp:     reading.Timestamp,
			Values:        valuesMap,
			Accuracy:      getAccuracyInt(reading.Accuracy),
			ReceivedAt:    parsedData.ReceivedAt,
			ProcessedData: processedData,
		}

		documents = append(documents, doc)
	}

	// 批量插入传感器数据
	if len(documents) > 0 {
		result, err := sensorDataColl.InsertMany(ctx, documents)
		if err != nil {
			return fmt.Errorf("保存传感器数据失败: %v", err)
		}
		Logger.Debug("传感器数据保存成功", 
			slog.Int("count", len(result.InsertedIDs)),
			slog.String("device_id", parsedData.DeviceID))
	}

	// 更新设备信息
	if err := updateDeviceInfo(parsedData); err != nil {
		Logger.Error("更新设备信息失败", 
			slog.String("error", err.Error()),
			slog.String("device_id", parsedData.DeviceID))
	}

	return nil
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
			DeviceID:     parsedData.DeviceID,
			FirstSeen:    parsedData.ReceivedAt,
			LastSeen:     parsedData.ReceivedAt,
			TotalRecords: int64(parsedData.TotalReadings),
			SensorTypes:  parsedData.SensorTypes,
			Sessions:     []string{parsedData.SessionID},
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
				"totalRecords": int64(parsedData.TotalReadings),
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

// GetSensorDataFromDB 从数据库获取传感器数据
func GetSensorDataFromDB(limit int, deviceID string, sensorType string) ([]SensorDataDocument, error) {
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
		filter["sensorType"] = sensorType
	}

	// 设置查询选项
	opts := options.Find().
		SetSort(bson.D{{Key: "receivedAt", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := sensorDataColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("查询传感器数据失败: %v", err)
	}
	defer cursor.Close(ctx)

	var results []SensorDataDocument
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

	// 总记录数
	totalRecords, err := sensorDataColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("查询总记录数失败: %v", err)
	}
	stats["totalRecords"] = totalRecords

	// 设备数量
	deviceCount, err := deviceInfoColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("查询设备数量失败: %v", err)
	}
	stats["deviceCount"] = deviceCount

	// 传感器类型数量
	sensorTypes, err := sensorDataColl.Distinct(ctx, "sensorType", bson.M{})
	if err != nil {
		return nil, fmt.Errorf("查询传感器类型失败: %v", err)
	}
	stats["sensorTypeCount"] = len(sensorTypes)
	stats["sensorTypes"] = sensorTypes

	// 最新数据时间
	var latestRecord SensorDataDocument
	opts := options.FindOne().SetSort(bson.D{{Key: "receivedAt", Value: -1}})
	err = sensorDataColl.FindOne(ctx, bson.M{}, opts).Decode(&latestRecord)
	if err == nil {
		stats["latestDataTime"] = latestRecord.ReceivedAt
	}

	Logger.Debug("统计信息查询完成", 
		slog.Int64("total_records", totalRecords),
		slog.Int64("device_count", deviceCount),
		slog.Int("sensor_types", len(sensorTypes)))

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
