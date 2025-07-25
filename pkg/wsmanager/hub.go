package wsmanager

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// Connection WebSocket连接
type Connection struct {
	conn     *websocket.Conn
	userID   int64
	send     chan []byte
	hub      *Hub
	lastSeen time.Time
}

// Hub WebSocket连接管理器
type Hub struct {
	// 注册的连接
	connections map[int64][]*Connection
	
	// 注册连接的channel
	register chan *Connection
	
	// 注销连接的channel
	unregister chan *Connection
	
	// 广播消息的channel
	broadcast chan *BroadcastMessage
	
	// 用于并发安全的读写锁
	mutex sync.RWMutex
}

// BroadcastMessage 广播消息结构
type BroadcastMessage struct {
	UserID int64
	Data   []byte
}

// MessageData WebSocket消息数据结构
type MessageData struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

var globalHub *Hub
var once sync.Once

// GetHub 获取全局Hub实例（单例模式）
func GetHub() *Hub {
	once.Do(func() {
		globalHub = &Hub{
			connections: make(map[int64][]*Connection),
			register:   make(chan *Connection),
			unregister: make(chan *Connection),
			broadcast:  make(chan *BroadcastMessage),
		}
		go globalHub.run()
	})
	return globalHub
}

// run 运行Hub
func (h *Hub) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case conn := <-h.register:
			h.registerConnection(conn)
		case conn := <-h.unregister:
			h.unregisterConnection(conn)
		case message := <-h.broadcast:
			h.broadcastToUser(message)
		case <-ticker.C:
			h.cleanupStaleConnections()
		}
	}
}

// registerConnection 注册连接
func (h *Hub) registerConnection(conn *Connection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if h.connections[conn.userID] == nil {
		h.connections[conn.userID] = make([]*Connection, 0)
	}
	h.connections[conn.userID] = append(h.connections[conn.userID], conn)
	
	logx.Infof("User %d connected, total connections: %d", conn.userID, len(h.connections[conn.userID]))
}

// unregisterConnection 注销连接
func (h *Hub) unregisterConnection(conn *Connection) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if connections, ok := h.connections[conn.userID]; ok {
		for i, c := range connections {
			if c == conn {
				// 关闭发送channel
				close(c.send)
				// 从切片中移除连接
				h.connections[conn.userID] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
		
		// 如果用户没有任何连接了，删除用户记录
		if len(h.connections[conn.userID]) == 0 {
			delete(h.connections, conn.userID)
		}
	}
	
	logx.Infof("User %d disconnected", conn.userID)
}

// broadcastToUser 向指定用户广播消息
func (h *Hub) broadcastToUser(message *BroadcastMessage) {
	h.mutex.RLock()
	connections := h.connections[message.UserID]
	h.mutex.RUnlock()
	
	if connections == nil {
		logx.Infof("User %d is not online", message.UserID)
		return
	}
	
	// 向用户的所有连接发送消息
	for _, conn := range connections {
		select {
		case conn.send <- message.Data:
		default:
			// 发送失败，关闭连接
			h.unregister <- conn
		}
	}
}

// cleanupStaleConnections 清理过期连接
func (h *Hub) cleanupStaleConnections() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	now := time.Now()
	for userID, connections := range h.connections {
		activeConnections := make([]*Connection, 0)
		for _, conn := range connections {
			if now.Sub(conn.lastSeen) < 5*time.Minute {
				activeConnections = append(activeConnections, conn)
			} else {
				close(conn.send)
				conn.conn.Close()
				logx.Infof("Cleaned up stale connection for user %d", userID)
			}
		}
		
		if len(activeConnections) == 0 {
			delete(h.connections, userID)
		} else {
			h.connections[userID] = activeConnections
		}
	}
}

// SendToUser 向指定用户发送消息
func (h *Hub) SendToUser(userID int64, messageType string, content interface{}) bool {
	data := MessageData{
		Type:    messageType,
		Content: content,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		logx.Errorf("Marshal message failed: %v", err)
		return false
	}
	
	message := &BroadcastMessage{
		UserID: userID,
		Data:   jsonData,
	}
	
	select {
	case h.broadcast <- message:
		return true
	default:
		logx.Errorf("Broadcast channel is full")
		return false
	}
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	connections, exists := h.connections[userID]
	return exists && len(connections) > 0
}

// GetOnlineUserCount 获取在线用户数
func (h *Hub) GetOnlineUserCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	return len(h.connections)
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request, userID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("WebSocket upgrade failed: %v", err)
		return
	}
	
	connection := &Connection{
		conn:     conn,
		userID:   userID,
		send:     make(chan []byte, 256),
		hub:      GetHub(),
		lastSeen: time.Now(),
	}
	
	// 注册连接
	GetHub().register <- connection
	
	// 启动读写协程
	go connection.writePump()
	go connection.readPump()
}

// writePump 写消息到WebSocket
func (c *Connection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			
			// 批量发送队列中的其他消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump 从WebSocket读消息
func (c *Connection) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.lastSeen = time.Now()
		return nil
	})
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Errorf("WebSocket error: %v", err)
			}
			break
		}
		
		c.lastSeen = time.Now()
		
		// 处理客户端发送的消息（如心跳、状态更新等）
		var msgData MessageData
		if err := json.Unmarshal(message, &msgData); err == nil {
			c.handleClientMessage(&msgData)
		}
	}
}

// handleClientMessage 处理客户端消息
func (c *Connection) handleClientMessage(msgData *MessageData) {
	switch msgData.Type {
	case "ping":
		// 心跳响应
		response := MessageData{
			Type:    "pong",
			Content: time.Now().Unix(),
		}
		if data, err := json.Marshal(response); err == nil {
			select {
			case c.send <- data:
			default:
				close(c.send)
			}
		}
	case "typing":
		// 处理正在输入状态
		logx.Infof("User %d is typing", c.userID)
	case "read_receipt":
		// 处理已读回执
		logx.Infof("User %d sent read receipt", c.userID)
	}
}