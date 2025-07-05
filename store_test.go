package main

import (
	"sync"
	"testing"
	"time"
)

// TestThreadSafeDataStore 测试线程安全数据存储的基本功能
func TestThreadSafeDataStore(t *testing.T) {
	store := NewThreadSafeDataStore()

	// 测试初始状态
	if store.Len() != 0 {
		t.Errorf("期望初始长度为0，实际为%d", store.Len())
	}

	if !store.IsEmpty() {
		t.Error("期望初始状态为空")
	}

	// 测试添加数据
	testData := ParsedSensorData{
		MessageID:     1,
		DeviceID:      "test-device",
		SessionID:     "test-session",
		TotalReadings: 5,
		ReceivedAt:    time.Now(),
	}

	store.Add(testData)

	if store.Len() != 1 {
		t.Errorf("期望长度为1，实际为%d", store.Len())
	}

	if store.IsEmpty() {
		t.Error("期望非空状态")
	}

	// 测试获取数据
	data := store.Get()
	if len(data) != 1 {
		t.Errorf("期望获取1条数据，实际为%d", len(data))
	}

	if data[0].MessageID != 1 {
		t.Errorf("期望消息ID为1，实际为%d", data[0].MessageID)
	}

	// 测试获取最新数据
	latest, exists := store.GetLatestOne()
	if !exists {
		t.Error("期望存在最新数据")
	}

	if latest.MessageID != 1 {
		t.Errorf("期望最新数据消息ID为1，实际为%d", latest.MessageID)
	}
}

// TestThreadSafeDataStoreTrimToSize 测试数据裁剪功能
func TestThreadSafeDataStoreTrimToSize(t *testing.T) {
	store := NewThreadSafeDataStore()

	// 添加10条数据
	for i := 0; i < 10; i++ {
		store.Add(ParsedSensorData{
			MessageID:     int64(i + 1),
			DeviceID:      "test-device",
			SessionID:     "test-session",
			TotalReadings: 1,
			ReceivedAt:    time.Now(),
		})
	}

	if store.Len() != 10 {
		t.Errorf("期望长度为10，实际为%d", store.Len())
	}

	// 裁剪到5条
	store.TrimToSize(5)

	if store.Len() != 5 {
		t.Errorf("期望裁剪后长度为5，实际为%d", store.Len())
	}

	// 验证保留的是最新的5条
	data := store.Get()
	for i, item := range data {
		expectedID := int64(i + 6) // 应该是6,7,8,9,10
		if item.MessageID != expectedID {
			t.Errorf("期望消息ID为%d，实际为%d", expectedID, item.MessageID)
		}
	}
}

// TestThreadSafeDataStoreGetLatest 测试获取最新N条数据
func TestThreadSafeDataStoreGetLatest(t *testing.T) {
	store := NewThreadSafeDataStore()

	// 添加5条数据
	for i := 0; i < 5; i++ {
		store.Add(ParsedSensorData{
			MessageID:     int64(i + 1),
			DeviceID:      "test-device",
			SessionID:     "test-session",
			TotalReadings: 1,
			ReceivedAt:    time.Now(),
		})
	}

	// 测试获取最新3条
	latest3 := store.GetLatest(3)
	if len(latest3) != 3 {
		t.Errorf("期望获取3条数据，实际为%d", len(latest3))
	}

	// 验证是最新的3条
	expectedIDs := []int64{3, 4, 5}
	for i, item := range latest3 {
		if item.MessageID != expectedIDs[i] {
			t.Errorf("期望消息ID为%d，实际为%d", expectedIDs[i], item.MessageID)
		}
	}

	// 测试获取超过总数量的数据
	latest10 := store.GetLatest(10)
	if len(latest10) != 5 {
		t.Errorf("期望获取5条数据，实际为%d", len(latest10))
	}
}

// TestThreadSafeDataStoreConcurrent 测试并发访问
func TestThreadSafeDataStoreConcurrent(t *testing.T) {
	store := NewThreadSafeDataStore()
	var wg sync.WaitGroup

	// 并发添加数据
	numGoroutines := 10
	numAdds := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numAdds; j++ {
				store.Add(ParsedSensorData{
					MessageID:     int64(id*numAdds + j),
					DeviceID:      "concurrent-device",
					SessionID:     "concurrent-session",
					TotalReadings: 1,
					ReceivedAt:    time.Now(),
				})
			}
		}(i)
	}

	wg.Wait()

	// 验证总数量
	expectedTotal := numGoroutines * numAdds
	if store.Len() != expectedTotal {
		t.Errorf("期望总数量为%d，实际为%d", expectedTotal, store.Len())
	}

	// 并发读取数据
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 读取操作
			store.Get()
			store.Len()
			store.IsEmpty()
			store.GetLatest(10)
			store.GetLatestOne()
		}()
	}

	wg.Wait()
}

// TestThreadSafeDataStoreEmptyOperations 测试空数据时的操作
func TestThreadSafeDataStoreEmptyOperations(t *testing.T) {
	store := NewThreadSafeDataStore()

	// 测试空数据时的GetLatestOne
	_, exists := store.GetLatestOne()
	if exists {
		t.Error("期望空数据时GetLatestOne返回false")
	}

	// 测试空数据时的GetLatest
	latest := store.GetLatest(5)
	if len(latest) != 0 {
		t.Errorf("期望空数据时GetLatest返回空切片，实际长度为%d", len(latest))
	}

	// 测试空数据时的Get
	data := store.Get()
	if len(data) != 0 {
		t.Errorf("期望空数据时Get返回空切片，实际长度为%d", len(data))
	}
}

// BenchmarkThreadSafeDataStoreAdd 基准测试：添加数据
func BenchmarkThreadSafeDataStoreAdd(b *testing.B) {
	store := NewThreadSafeDataStore()
	testData := ParsedSensorData{
		MessageID:     1,
		DeviceID:      "bench-device",
		SessionID:     "bench-session",
		TotalReadings: 1,
		ReceivedAt:    time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Add(testData)
	}
}

// BenchmarkThreadSafeDataStoreGet 基准测试：获取数据
func BenchmarkThreadSafeDataStoreGet(b *testing.B) {
	store := NewThreadSafeDataStore()

	// 预先添加一些数据
	for i := 0; i < 1000; i++ {
		store.Add(ParsedSensorData{
			MessageID:     int64(i),
			DeviceID:      "bench-device",
			SessionID:     "bench-session",
			TotalReadings: 1,
			ReceivedAt:    time.Now(),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get()
	}
}
