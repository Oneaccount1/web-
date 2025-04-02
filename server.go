package main

import (
	"ChatRoom/util/protocol"
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type Client struct {
	conn    net.Conn
	name    string
	enterAt time.Time
	id      string
}

type ChatServer struct {
	clients     map[*Client]bool
	clientsLock sync.RWMutex
	flags       map[string]bool
	rdb         *redis.Client

	// 消息队列相关字段
	messageQueue chan protocol.Message
	workerCount  int
}

var rdb = redis.NewClient(&redis.Options{
	Addr:     "192.168.30.128:6379",
	Password: "",
	DB:       0,
})

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:      make(map[*Client]bool),
		flags:        make(map[string]bool),
		rdb:          rdb,
		messageQueue: make(chan protocol.Message, 1000),
		workerCount:  5,
	}
}

func (s *ChatServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	client := &Client{
		conn:    conn,
		enterAt: time.Now(),
	}

	if err := s.readUserName(client); err != nil {
		fmt.Println("读取用户名失败:", err)
		return
	}

	s.registerClient(client)
	defer s.unregisterClient(client)
	s.sendWelcomeMessage(client)

	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))

		msg, err := protocol.DecodeMessage(reader)
		if err != nil {
			s.handleReadError(client, err)
			return
		}

		switch msg.Type {
		case protocol.MessageTypeNormal:
			s.incrementUserActivity(client.id)
			s.asyncBroadcast(msg)

		case protocol.MessageTypeRank:
			s.handleRankRequest(client)

		case protocol.MessageTypeLeave:
			s.leave(*client)
			return

		case protocol.MessageTypeHeartbeat:
			fmt.Printf("[%s] 心跳保持\n", client.name)
		}
	}
}

// 优化消息广播方法
func (s *ChatServer) asyncBroadcast(msg protocol.Message) {
	// 将消息放入队列
	s.messageQueue <- msg
}

// 消息处理工作协程
func (s *ChatServer) startMessageWorkers() {
	for i := 0; i < s.workerCount; i++ {
		go s.messageWorker(i)
	}
}

func (s *ChatServer) messageWorker(id int) {
	for msg := range s.messageQueue {
		// 直接发送给所有客户端(除发送者外)
		s.clientsLock.RLock()
		for client := range s.clients {
			// 使用ID验证用户唯一性，只发送给其他客户端
			if client.id != msg.Id {
				if err := s.sendMessage(client, msg); err != nil {
					fmt.Printf("发送消息到 %s(%s) 失败: %v\n", client.name, client.id, err)
				}
			}
		}
		s.clientsLock.RUnlock()
	}
}

// 活跃度更新方法
func (s *ChatServer) incrementUserActivity(userID string) {
	ctx := context.Background()
	pipe := s.rdb.Pipeline()

	// 增加活跃度分数
	pipe.ZIncrBy(ctx, "user:activity", 1, userID)

	// 更新最后活跃时间
	pipe.HSet(ctx, "user:"+userID, "last_active", time.Now().Unix())

	// 执行管道命令
	if _, err := pipe.Exec(ctx); err != nil {
		fmt.Printf("更新用户活跃度失败: %v\n", err)
	}
}

// 排行榜处理方法
func (s *ChatServer) handleRankRequest(client *Client) {
	ctx := context.Background()
	pipe := s.rdb.Pipeline()

	// 获取排行榜数据
	topUsersCmd := pipe.ZRevRangeWithScores(ctx, "user:activity", 0, 9)
	userScoreCmd := pipe.ZScore(ctx, "user:activity", client.id)
	userRankCmd := pipe.ZRevRank(ctx, "user:activity", client.id)

	// 执行管道命令
	if _, err := pipe.Exec(ctx); err != nil {
		fmt.Printf("查询排行榜失败: %v\n", err)
		return
	}

	topUsers, _ := topUsersCmd.Result()
	userScore, _ := userScoreCmd.Result()
	userRank, _ := userRankCmd.Result()

	var sb strings.Builder
	sb.WriteString("=== 活跃度排行榜 ===\n")

	for i, item := range topUsers {
		userID, ok := item.Member.(string)
		if !ok {
			continue
		}

		// 获取用户信息
		userInfo, err := s.rdb.HGetAll(ctx, "user:"+userID).Result()
		if err != nil {
			continue
		}

		name := userInfo["name"]
		lastActive := "刚刚"
		if lastActiveTime, ok := userInfo["last_active"]; ok {
			if lastActiveInt, err := strconv.ParseInt(lastActiveTime, 10, 64); err == nil {
				lastActive = time.Unix(lastActiveInt, 0).Format("15:04")
			}
		}

		mark := ""
		if userID == client.id {
			mark = " ← 您"
		}

		sb.WriteString(fmt.Sprintf("%d. %-15s (%s) (积分: %4.0f) [最后活跃: %s]%s\n",
			i+1, name, userID, item.Score, lastActive, mark))
	}

	rankInfo := "您尚未上榜"
	if userRank >= 0 {
		rankInfo = fmt.Sprintf("您的排名: %d位", userRank+1)
	}

	sb.WriteString(fmt.Sprintf("\n%s [%s(%s)] 当前积分: %.0f\n",
		rankInfo, client.name, client.id, userScore))

	rankMsg := protocol.Message{
		Type:    protocol.MessageTypeRank,
		Content: sb.String(),
		Time:    time.Now().Unix(),
	}
	s.sendMessage(client, rankMsg)
}

func (s *ChatServer) registerClient(client *Client) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	// 生成唯一ID
	client.id = uuid.New().String()[:4]
	s.clients[client] = true
	s.flags[client.id] = true

	ctx := context.Background()
	pipe := s.rdb.Pipeline()

	// 初始化活跃度
	pipe.ZAddNX(ctx, "user:activity", &redis.Z{
		Score:  0,
		Member: client.id,
	})

	// 存储用户信息
	pipe.HSet(ctx, "user:"+client.id,
		"name", client.name,
		"join_time", client.enterAt.Unix(),
		"last_active", time.Now().Unix(),
	)

	if _, err := pipe.Exec(ctx); err != nil {
		fmt.Printf("注册用户信息失败: %v\n", err)
	}
}

func (s *ChatServer) unregisterClient(client *Client) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	if _, exists := s.clients[client]; exists {
		delete(s.clients, client)
		delete(s.flags, client.id)
		go s.removeUserActivity(client.id)
	}
}

// 删除用户活跃度
func (s *ChatServer) removeUserActivity(userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pipe := s.rdb.Pipeline()

	// 删除活跃度记录
	pipe.ZRem(ctx, "user:activity", userID)

	// 删除用户信息
	pipe.Del(ctx, "user:"+userID)

	if _, err := pipe.Exec(ctx); err != nil {
		fmt.Printf("删除用户活跃度失败: %v\n", err)
	}
}

// 消息发送方法
func (s *ChatServer) sendMessage(client *Client, msg protocol.Message) error {
	data, err := protocol.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("编码失败: %w", err)
	}

	// 添加写入超时
	client.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = client.conn.Write(data)
	if err != nil {
		return fmt.Errorf("发送失败: %w", err)
	}

	return nil
}

func (s *ChatServer) sendWelcomeMessage(client *Client) {
	joinMsg := protocol.Message{
		Type:    protocol.MessageTypeJoin,
		From:    client.name,
		Content: fmt.Sprintf("Welcome %s!", client.name),
		Time:    time.Now().Unix(),
		Id:      client.id,
	}
	s.sendMessage(client, joinMsg)

	systemMsg := protocol.Message{
		Type:    protocol.MessageTypeSystem,
		From:    "系统",
		Content: fmt.Sprintf("%s(%s) 上线，当前在线人数 %d", client.name, client.id, s.onlineUsers()),
		Time:    time.Now().Unix(),
	}
	s.broadcastMessage(systemMsg)
}

func (s *ChatServer) broadcastMessage(msg protocol.Message) {
	s.clientsLock.RLock()
	defer s.clientsLock.RUnlock()

	for client := range s.clients {
		go func(c *Client) {
			if err := s.sendMessage(c, msg); err != nil {
				fmt.Printf("发送消息到 %s(%s) 失败: %v\n", c.name, c.id, err)
			}
		}(client)
	}
}

func (s *ChatServer) leave(client Client) {
	leaveMsg := protocol.Message{
		Type:    protocol.MessageTypeSystem,
		From:    "系统",
		Content: fmt.Sprintf("用户 %s(%s) 下线，当前剩余人数 %d", client.name, client.id, s.onlineUsers()-1),
		Time:    time.Now().Unix(),
	}
	s.broadcastMessage(leaveMsg)
}

func (s *ChatServer) onlineUsers() int {
	s.clientsLock.RLock()
	defer s.clientsLock.RUnlock()
	return len(s.clients)
}

func (s *ChatServer) readUserName(client *Client) error {
	msg, err := protocol.DecodeMessage(client.conn)
	if err != nil {
		return fmt.Errorf("decode message error: %w", err)
	}
	client.name = msg.From
	return nil
}

func (s *ChatServer) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Chat server started on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// 处理连接错误方法
func (s *ChatServer) handleReadError(client *Client, err error) {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		fmt.Printf("[%s(%s)] 连接超时\n", client.name, client.id)
	} else {
		fmt.Printf("[%s(%s)] 连接异常: %v\n", client.name, client.id, err)
	}
	s.leave(*client)
	s.unregisterClient(client)
}

func main() {
	server := NewChatServer()

	// 启动消息处理工作协程
	server.startMessageWorkers()

	if err := server.Start(":8888"); err != nil {
		panic(err)
	}
}
