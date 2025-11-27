package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go-cesi/internal/logger"
	"go-cesi/internal/supervisor"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	MaxConnections     int           // 最大连接数
	RateLimit         float64       // 每秒消息限制
	RateBurst         int           // 突发消息数量
	HeartbeatInterval time.Duration // 心跳间隔
	ReadTimeout       time.Duration // 读取超时
	WriteTimeout      time.Duration // 写入超时
	AllowedOrigins    []string      // 允许的来源
	MaxMessageSize    int64         // 最大消息大小
	MaxViolations     int           // 最大违规次数
}

// GetDefaultWebSocketConfig 获取默认WebSocket配置
func GetDefaultWebSocketConfig() *WebSocketConfig {
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:8081",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:8081",
	}

	// 从环境变量读取允许的来源
	if origins := os.Getenv("WEBSOCKET_ALLOWED_ORIGINS"); origins != "" {
		allowedOrigins = strings.Split(origins, ",")
	}

	return &WebSocketConfig{
		MaxConnections:     100,              // 最大100个连接
		RateLimit:         10.0,             // 每秒10条消息
		RateBurst:         20,               // 突发20条消息
		HeartbeatInterval: 30 * time.Second, // 30秒心跳
		ReadTimeout:       60 * time.Second, // 60秒读取超时
		WriteTimeout:      10 * time.Second, // 10秒写入超时
		AllowedOrigins:    allowedOrigins,
		MaxMessageSize:    1024,             // 1KB最大消息大小
		MaxViolations:     5,                // 最大5次违规
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		config := GetDefaultWebSocketConfig()

		for _, allowed := range config.AllowedOrigins {
			if origin == allowed {
				return true
			}
		}

		logger.Warn("WebSocket connection rejected", zap.String("origin", origin))
		return false
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	clients       map[*Client]bool
	broadcast     chan []byte
	register      chan *Client
	unregister    chan *Client
	service       *supervisor.SupervisorService
	config        *WebSocketConfig
	connectionCount int64 // 原子操作的连接计数
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	userID     string
	subscribed map[string]bool // subscribed nodes
	limiter    *rate.Limiter   // 速率限制器
	lastPong   time.Time       // 最后一次pong时间
	mu         sync.RWMutex
	violationCount int          // 违规计数
	closed     bool            // 连接是否已关闭
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type NodeUpdateMessage struct {
	NodeName  string      `json:"node_name"`
	Processes interface{} `json:"processes"`
	Timestamp time.Time   `json:"timestamp"`
}

type ProcessStatusMessage struct {
	NodeName    string    `json:"node_name"`
	ProcessName string    `json:"process_name"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

type SystemStatsMessage struct {
	TotalNodes       int       `json:"total_nodes"`
	ConnectedNodes   int       `json:"connected_nodes"`
	RunningProcesses int       `json:"running_processes"`
	StoppedProcesses int       `json:"stopped_processes"`
	Timestamp        time.Time `json:"timestamp"`
}

func NewHub(service *supervisor.SupervisorService) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		service:    service,
		config:     GetDefaultWebSocketConfig(),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// NewHubWithConfig 使用自定义配置创建Hub
func NewHubWithConfig(service *supervisor.SupervisorService, config *WebSocketConfig) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		service:    service,
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Close 关闭Hub
func (h *Hub) Close() {
	h.cancel()
}

// GetConnectionCount 获取当前连接数
func (h *Hub) GetConnectionCount() int64 {
	return atomic.LoadInt64(&h.connectionCount)
}

func (h *Hub) Run() {
	// Start periodic updates
	go h.startPeriodicUpdates()
	// Start heartbeat checker
	go h.startHeartbeatChecker()

	for {
		select {
		case <-h.ctx.Done():
			logger.Info("Hub shutting down")
			return

		case client := <-h.register:
			// 检查连接数限制
			if atomic.LoadInt64(&h.connectionCount) >= int64(h.config.MaxConnections) {
				logger.Warn("Connection limit reached, rejecting new connection",
					zap.String("userID", client.userID),
					zap.Int64("current_connections", atomic.LoadInt64(&h.connectionCount)),
					zap.Int("max_connections", h.config.MaxConnections))
				client.conn.Close()
				continue
			}

			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			atomic.AddInt64(&h.connectionCount, 1)
			logger.Info("WebSocket client connected",
				zap.String("user_id", client.userID),
				zap.Int64("total_connections", atomic.LoadInt64(&h.connectionCount)))

			// Send initial data to new client
			go h.sendInitialData(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				atomic.AddInt64(&h.connectionCount, -1)
			}
			h.mu.Unlock()
			logger.Info("WebSocket client disconnected",
				zap.String("user_id", client.userID),
				zap.Int64("total_connections", atomic.LoadInt64(&h.connectionCount)))

		case message := <-h.broadcast:
			// 直接分发消息到所有客户端
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
					// 发送成功
				default:
					// 客户端阻塞，异步移除
					go func(c *Client) {
						h.mu.Lock()
						if _, ok := h.clients[c]; ok {
							close(c.send)
							delete(h.clients, c)
							atomic.AddInt64(&h.connectionCount, -1)
						}
						h.mu.Unlock()
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) startPeriodicUpdates() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.broadcastNodesUpdate()
			h.broadcastSystemStats()
		}
	}
}

// startHeartbeatChecker 启动心跳检测
func (h *Hub) startHeartbeatChecker() {
	ticker := time.NewTicker(h.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.checkHeartbeats()
		}
	}
}

// checkHeartbeats 检查所有客户端的心跳状态
func (h *Hub) checkHeartbeats() {
	h.mu.RLock()
	deadClients := make([]*Client, 0)
	now := time.Now()
	heartbeatTimeout := h.config.HeartbeatInterval * 3 // 更严格的超时时间

	for client := range h.clients {
		client.mu.RLock()
		// 检查客户端是否已关闭或超时
		if client.closed || now.Sub(client.lastPong) > heartbeatTimeout {
			if !client.closed {
				deadClients = append(deadClients, client)
			}
		}
		client.mu.RUnlock()
	}
	h.mu.RUnlock()

	// 安全地断开超时的客户端
	for _, client := range deadClients {
		client.mu.Lock()
		if !client.closed {
			client.closed = true
			logger.Warn("Client heartbeat timeout, disconnecting",
				zap.String("userID", client.userID),
				zap.Duration("timeout", now.Sub(client.lastPong)))
			
			// 通过unregister channel安全地移除客户端
			select {
			case h.unregister <- client:
			default:
				// 如果channel满了，直接关闭连接
				client.conn.Close()
			}
		}
		client.mu.Unlock()
	}

	// 发送ping消息给所有活跃客户端
	pingMessage := map[string]interface{}{
		"type": "ping",
		"timestamp": now.Unix(),
	}

	pingData, err := json.Marshal(pingMessage)
	if err != nil {
		logger.Error("Failed to marshal ping message", zap.Error(err))
		return
	}

	// 使用超时机制发送ping
	select {
	case h.broadcast <- pingData:
	case <-time.After(1 * time.Second):
		logger.Warn("Ping broadcast timeout, channel may be blocked")
	}
}

func (h *Hub) sendInitialData(client *Client) {
	// Check if client is still registered
	h.mu.RLock()
	_, exists := h.clients[client]
	h.mu.RUnlock()
	
	if !exists {
		logger.Debug("Client already disconnected, skipping initial data",
			zap.String("user_id", client.userID))
		return
	}

	// Send current nodes data
	nodes := h.service.GetAllNodes()
	nodesData := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		nodesData[i] = node.Serialize()
	}

	message := Message{
		Type: "nodes_update",
		Data: nodesData,
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Error marshaling initial data",
			zap.Error(err),
			zap.String("user_id", client.userID))
		return
	}

	// Try to send with timeout, but don't panic if channel is closed
	select {
	case client.send <- data:
		// Successfully sent
	case <-time.After(1 * time.Second):
		// Timeout - client may have disconnected
		logger.Warn("Failed to send initial data to client, timeout",
			zap.String("user_id", client.userID))
	}
}

func (h *Hub) broadcastNodesUpdate() {
	nodes := h.service.GetAllNodes()
	nodesData := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		nodesData[i] = node.Serialize()
	}

	message := Message{
		Type: "nodes_update",
		Data: nodesData,
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Error marshaling nodes update", zap.Error(err))
		return
	}

	select {
	case h.broadcast <- data:
	default:
		logger.Warn("Broadcast channel full, dropping nodes update")
	}
}

func (h *Hub) broadcastSystemStats() {
	nodes := h.service.GetAllNodes()
	totalNodes := len(nodes)
	connectedNodes := 0
	runningProcesses := 0
	stoppedProcesses := 0

	for _, node := range nodes {
		if node.IsConnected {
			connectedNodes++
			for _, process := range node.Processes {
				if process.State == 20 { // RUNNING state in supervisor
					runningProcesses++
				} else {
					stoppedProcesses++
				}
			}
		}
	}

	stats := SystemStatsMessage{
		TotalNodes:       totalNodes,
		ConnectedNodes:   connectedNodes,
		RunningProcesses: runningProcesses,
		StoppedProcesses: stoppedProcesses,
		Timestamp:        time.Now(),
	}

	message := Message{
		Type: "system_stats",
		Data: stats,
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Error marshaling system stats", zap.Error(err))
		return
	}

	select {
	case h.broadcast <- data:
	default:
		logger.Warn("Broadcast channel full, dropping system stats")
	}
}

func (h *Hub) BroadcastProcessStatusChange(nodeName, processName, status string) {
	message := Message{
		Type: "process_status_change",
		Data: ProcessStatusMessage{
			NodeName:    nodeName,
			ProcessName: processName,
			Status:      status,
			Timestamp:   time.Now(),
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		logger.Error("Error marshaling process status change", zap.Error(err))
		return
	}

	select {
	case h.broadcast <- data:
	default:
		logger.Warn("Broadcast channel full, dropping process status change")
	}
}

// handleViolation 处理客户端违规行为
func (c *Client) handleViolation(reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.violationCount++
	logger.Warn("Client violation detected",
		zap.String("userID", c.userID),
		zap.String("reason", reason),
		zap.Int("violationCount", c.violationCount),
		zap.Int("maxViolations", c.hub.config.MaxViolations))

	// 如果违规次数超过阈值，强制断开连接
	if c.violationCount >= c.hub.config.MaxViolations {
		c.closed = true
		logger.Error("Client exceeded max violations, force disconnecting",
			zap.String("userID", c.userID),
			zap.Int("violationCount", c.violationCount))
		
		// 异步关闭连接以避免阻塞
		go func() {
			c.conn.Close()
		}()
	}
}



func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade error", zap.Error(err))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	// 设置连接超时
	conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(h.config.WriteTimeout))

	// 创建速率限制器
	limiter := rate.NewLimiter(rate.Limit(h.config.RateLimit), h.config.RateBurst)

	client := &Client{
		hub:            h,
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         userID,
		subscribed:     make(map[string]bool),
		limiter:        limiter,
		lastPong:       time.Now(),
		violationCount: 0,
		closed:         false,
	}

	// 设置pong处理器
	conn.SetPongHandler(func(string) error {
		client.mu.Lock()
		client.lastPong = time.Now()
		client.mu.Unlock()
		conn.SetReadDeadline(time.Now().Add(h.config.ReadTimeout))
		return nil
	})

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(data []byte) {
	select {
	case h.broadcast <- data:
	default:
		logger.Warn("Broadcast channel full, message dropped")
	}
}
