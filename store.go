package main

import (
	"sync"
)

// ThreadSafeDataStore 线程安全的数据存储
type ThreadSafeDataStore struct {
	data  []ParsedSensorData
	mutex sync.RWMutex
}

// NewThreadSafeDataStore 创建新的线程安全数据存储
func NewThreadSafeDataStore() *ThreadSafeDataStore {
	return &ThreadSafeDataStore{
		data: make([]ParsedSensorData, 0),
	}
}

// Add 添加数据
func (ts *ThreadSafeDataStore) Add(data ParsedSensorData) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.data = append(ts.data, data)
}

// Get 获取所有数据的副本
func (ts *ThreadSafeDataStore) Get() []ParsedSensorData {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	result := make([]ParsedSensorData, len(ts.data))
	copy(result, ts.data)
	return result
}

// GetLatest 获取最新的N条数据
func (ts *ThreadSafeDataStore) GetLatest(n int) []ParsedSensorData {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	if len(ts.data) == 0 {
		return []ParsedSensorData{}
	}

	if n >= len(ts.data) {
		result := make([]ParsedSensorData, len(ts.data))
		copy(result, ts.data)
		return result
	}

	result := make([]ParsedSensorData, n)
	copy(result, ts.data[len(ts.data)-n:])
	return result
}

// GetLatestOne 获取最新的一条数据
func (ts *ThreadSafeDataStore) GetLatestOne() (ParsedSensorData, bool) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	if len(ts.data) == 0 {
		return ParsedSensorData{}, false
	}

	return ts.data[len(ts.data)-1], true
}

// Len 获取数据长度
func (ts *ThreadSafeDataStore) Len() int {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return len(ts.data)
}

// IsEmpty 检查是否为空
func (ts *ThreadSafeDataStore) IsEmpty() bool {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return len(ts.data) == 0
}

// TrimToSize 保持数据量在指定大小内
func (ts *ThreadSafeDataStore) TrimToSize(maxSize int) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if len(ts.data) > maxSize {
		ts.data = ts.data[len(ts.data)-maxSize:]
	}
}

// GetAllForRead 获取所有数据用于只读操作（不复制，需要外部保证只读）
func (ts *ThreadSafeDataStore) GetAllForRead() []ParsedSensorData {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return ts.data
}
