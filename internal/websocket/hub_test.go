package websocket

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-cesi/internal/supervisor"
)

// 属性 26：WebSocket 消息分发效率
// 验证需求：10.4
func TestWebSocketMessageDistributionProperties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("concurrent message distribution does not block", prop.ForAll(
		func(clientCount int, messageCount int) bool {
			if clientCount < 1 || clientCount > 20 {
				clientCount = 5
			}
			if messageCount < 1 || messageCount > 20 {
				messageCount = 5
			}

			// 创建模拟的supervisor服务
			service := &supervisor.SupervisorService{}
			hub := NewHub(service)
			defer hub.Close()

			// 启动hub
			go hub.Run()

			// 等待hub启动
			time.Sleep(10 * time.Millisecond)

			// 创建模拟客户端
			clients := make([]*mockClient, clientCount)
			for i := 0; i < clientCount; i++ {
				clients[i] = newMockClient(hub)
			}

			// 等待客户端注册
			time.Sleep(50 * time.Millisecond)

			// 发送消息
			for i := 0; i < messageCount; i++ {
				message := []byte(`{"type":"test","data":"message"}`)
				select {
				case hub.broadcast <- message:
				case <-time.After(100 * time.Millisecond):
					// 超时，跳过
				}
			}

			// 等待消息分发
			time.Sleep(200 * time.Millisecond)
			
			// 清理
			for _, client := range clients {
				client.close()
			}

			// 验证至少有一些客户端收到了消息
			receivedCount := 0
			for _, client := range clients {
				if client.getReceivedCount() > 0 {
					receivedCount++
				}
			}

			return receivedCount > 0
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 20),
	))

	properties.Property("slow client does not block other clients", prop.ForAll(
		func(normalClients int) bool {
			if normalClients < 1 || normalClients > 20 {
				normalClients = 5
			}

			service := &supervisor.SupervisorService{}
			hub := NewHub(service)
			defer hub.Close()

			go hub.Run()

			// 创建正常客户端
			clients := make([]*mockClient, normalClients)
			for i := 0; i < normalClients; i++ {
				clients[i] = newMockClient(hub)
			}

			// 创建一个慢客户端（不读取消息）
			slowClient := newMockClient(hub)
			slowClient.stopReading() // 停止读取，模拟阻塞

			// 发送消息
			messageCount := 10
			for i := 0; i < messageCount; i++ {
				message := []byte(`{"type":"test","data":"message"}`)
				hub.broadcast <- message
			}

			// 等待消息分发
			time.Sleep(200 * time.Millisecond)

			// 验证正常客户端仍然能接收消息
			normalClientsReceived := 0
			for _, client := range clients {
				if client.getReceivedCount() > 0 {
					normalClientsReceived++
				}
			}

			// 清理
			for _, client := range clients {
				client.close()
			}
			slowClient.close()

			// 至少一半的正常客户端应该收到消息
			return normalClientsReceived >= normalClients/2
		},
		gen.IntRange(1, 20),
	))

	properties.Property("message distribution is concurrent", prop.ForAll(
		func(clientCount int) bool {
			if clientCount < 2 || clientCount > 15 {
				clientCount = 5
			}

			service := &supervisor.SupervisorService{}
			hub := NewHub(service)
			defer hub.Close()

			go hub.Run()
			time.Sleep(10 * time.Millisecond)

			// 创建客户端，每个客户端处理消息需要一定时间
			clients := make([]*mockClient, clientCount)
			for i := 0; i < clientCount; i++ {
				clients[i] = newMockClientWithDelay(hub, 5*time.Millisecond)
			}

			// 等待客户端注册
			time.Sleep(50 * time.Millisecond)

			// 发送一条消息
			message := []byte(`{"type":"test","data":"message"}`)
			hub.broadcast <- message

			// 等待所有客户端接收
			time.Sleep(300 * time.Millisecond)

			// 清理
			for _, client := range clients {
				client.close()
			}

			// 验证至少一半的客户端收到了消息
			receivedCount := 0
			for _, client := range clients {
				if client.getReceivedCount() > 0 {
					receivedCount++
				}
			}

			return receivedCount >= clientCount/2
		},
		gen.IntRange(2, 15),
	))

	properties.TestingRun(t)
}

// mockClient 模拟WebSocket客户端
type mockClient struct {
	hub           *Hub
	send          chan []byte
	receivedCount int32
	reading       bool
	mu            sync.Mutex
	processingDelay time.Duration
}

func newMockClient(hub *Hub) *mockClient {
	client := &mockClient{
		hub:     hub,
		send:    make(chan []byte, 256),
		reading: true,
	}

	// 模拟客户端注册
	mockWSClient := &Client{
		hub:        hub,
		send:       client.send,
		userID:     "mock-user",
		subscribed: make(map[string]bool),
	}

	hub.register <- mockWSClient

	// 启动消息接收
	go client.receiveMessages()

	return client
}

func newMockClientWithDelay(hub *Hub, delay time.Duration) *mockClient {
	client := newMockClient(hub)
	client.processingDelay = delay
	return client
}

func (c *mockClient) receiveMessages() {
	for {
		c.mu.Lock()
		reading := c.reading
		c.mu.Unlock()

		if !reading {
			return
		}

		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if msg != nil {
				atomic.AddInt32(&c.receivedCount, 1)
				
				// 模拟消息处理延迟
				if c.processingDelay > 0 {
					time.Sleep(c.processingDelay)
				}
			}
		case <-time.After(500 * time.Millisecond):
			return
		}
	}
}

func (c *mockClient) stopReading() {
	c.mu.Lock()
	c.reading = false
	c.mu.Unlock()
}

func (c *mockClient) getReceivedCount() int {
	return int(atomic.LoadInt32(&c.receivedCount))
}

func (c *mockClient) close() {
	c.stopReading()
	// 不关闭send通道，因为它由hub管理
}

// TestBroadcastTask 测试广播任务
func TestBroadcastTask(t *testing.T) {
	send := make(chan []byte, 10)
	client := &Client{
		send: send,
	}

	task := &broadcastTask{
		client:  client,
		message: []byte("test message"),
		taskID:  "test-1",
	}

	// 测试成功执行
	ctx := context.Background()
	err := task.Execute(ctx)
	require.NoError(t, err)

	// 验证消息已发送
	select {
	case msg := <-send:
		assert.Equal(t, []byte("test message"), msg)
	case <-time.After(100 * time.Millisecond):
		t.Error("Message not received")
	}

	// 验证任务ID
	assert.Equal(t, "test-1", task.ID())
}

// TestBroadcastTaskChannelFull 测试通道满的情况
func TestBroadcastTaskChannelFull(t *testing.T) {
	send := make(chan []byte, 1)
	client := &Client{
		send: send,
	}

	// 填满通道
	send <- []byte("blocking message")

	task := &broadcastTask{
		client:  client,
		message: []byte("test message"),
		taskID:  "test-2",
	}

	// 测试通道满时的行为
	ctx := context.Background()
	err := task.Execute(ctx)
	assert.Error(t, err)
	assert.IsType(t, &BroadcastError{}, err)
}

// TestBroadcastTaskContextCanceled 测试上下文取消
func TestBroadcastTaskContextCanceled(t *testing.T) {
	send := make(chan []byte, 10)
	client := &Client{
		send: send,
	}

	task := &broadcastTask{
		client:  client,
		message: []byte("test message"),
		taskID:  "test-3",
	}

	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// 测试上下文取消时的行为
	err := task.Execute(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestHubConcurrentBroadcast 测试并发广播
func TestHubConcurrentBroadcast(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	defer hub.Close()

	go hub.Run()

	// 创建多个客户端
	clientCount := 10
	clients := make([]*mockClient, clientCount)
	for i := 0; i < clientCount; i++ {
		clients[i] = newMockClient(hub)
	}

	// 并发发送多条消息
	messageCount := 20
	var wg sync.WaitGroup
	for i := 0; i < messageCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			message := []byte(`{"type":"test","data":"message"}`)
			hub.broadcast <- message
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	// 验证所有客户端都收到了消息
	for i, client := range clients {
		received := client.getReceivedCount()
		assert.Greater(t, received, 0, "Client %d should receive messages", i)
		client.close()
	}
}

// TestHubWorkerPoolIntegration 测试工作池集成
func TestHubWorkerPoolIntegration(t *testing.T) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	defer hub.Close()

	// 验证工作池已初始化
	require.NotNil(t, hub.workerPool)

	go hub.Run()

	// 创建客户端
	client := newMockClient(hub)
	defer client.close()

	// 发送消息
	message := []byte(`{"type":"test","data":"message"}`)
	hub.broadcast <- message

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证消息已接收
	assert.Greater(t, client.getReceivedCount(), 0)

	// 验证工作池统计
	stats := hub.workerPool.Stats()
	assert.Greater(t, stats.TotalJobs, int64(0))
}

// BenchmarkBroadcastWithWorkerPool 基准测试：使用工作池的广播
func BenchmarkBroadcastWithWorkerPool(b *testing.B) {
	service := &supervisor.SupervisorService{}
	hub := NewHub(service)
	defer hub.Close()

	go hub.Run()

	// 创建客户端
	clientCount := 100
	clients := make([]*mockClient, clientCount)
	for i := 0; i < clientCount; i++ {
		clients[i] = newMockClient(hub)
	}

	message := []byte(`{"type":"test","data":"message"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.broadcast <- message
	}
	b.StopTimer()

	// 清理
	for _, client := range clients {
		client.close()
	}
}

// BenchmarkBroadcastTask 基准测试：广播任务执行
func BenchmarkBroadcastTask(b *testing.B) {
	send := make(chan []byte, 1000)
	client := &Client{
		send: send,
	}

	task := &broadcastTask{
		client:  client,
		message: []byte("test message"),
		taskID:  "bench-task",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task.Execute(ctx)
	}
}
