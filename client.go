package main

import (
	"ChatRoom/util/protocol"
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

type ChatClient struct {
	conn     net.Conn
	user     protocol.Message
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
	stopChan chan struct{}

	// 活跃度相关字段
	lastActivityTime time.Time
	activityScore    float64
}

func NewChatClient() *ChatClient {
	return &ChatClient{
		stopChan:         make(chan struct{}),
		lastActivityTime: time.Now(),
		activityScore:    0,
	}
}

func (c *ChatClient) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("连接服务器失败: %w", err)
	}
	c.conn = conn
	c.ctx, c.cancel = context.WithCancel(context.Background())
	return nil
}

func (c *ChatClient) Login(username string) error {
	loginMsg := protocol.Message{
		Type: protocol.MessageTypeJoin,
		From: username,
		Time: time.Now().Unix(),
	}

	if err := c.SendMessage(loginMsg); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}

	// 等待服务器返回用户ID
	msg, err := c.waitForLoginResponse()
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.user = msg
	c.mu.Unlock()

	return nil
}

func (c *ChatClient) waitForLoginResponse() (protocol.Message, error) {
	reader := bufio.NewReader(c.conn)
	for {
		msg, err := protocol.DecodeMessage(reader)
		if err != nil {
			return protocol.Message{}, err
		}

		if msg.Type == protocol.MessageTypeJoin && msg.Id != "" {
			return msg, nil
		}
	}
}

func (c *ChatClient) Start() {
	c.wg.Add(2)
	go c.receiveMessages()
	go c.sendHeartbeat()

	fmt.Printf("%s已成功登录，输入消息开始聊天（输入/help查看帮助）%s\n", colorGreen, colorReset)
	c.printHelp()
	c.handleInput()
}

func (c *ChatClient) handleInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-c.stopChan:
			return
		default:
			if !scanner.Scan() {
				return
			}

			text := strings.TrimSpace(scanner.Text())
			switch {
			case text == "/quit":
				c.Shutdown()
				return
			case text == "/help":
				c.printHelp()
			case text == "/rank":
				c.requestRank()
			case text == "/clear":
				fmt.Print("\033[H\033[2J") // 清屏
			case text == "/stats":
				c.showStats()
			case text == "":
				continue
			default:
				c.sendNormalMessage(text)
			}
		}
	}
}

func (c *ChatClient) printHelp() {
	fmt.Printf(`
%s可用命令：
  /help    - 显示帮助信息
  /rank    - 查看活跃度排行榜
  /quit    - 退出聊天室
  /clear   - 清屏
  /stats   - 查看个人统计信息
%s`, colorCyan, colorReset)
}

func (c *ChatClient) requestRank() {
	msg := protocol.Message{
		Type: protocol.MessageTypeRank,
		From: c.user.From,
		Id:   c.user.Id,
		Time: time.Now().Unix(),
	}

	if err := c.SendMessage(msg); err != nil {
		fmt.Printf("%s排行榜请求失败: %v%s\n", colorRed, err, colorReset)
	}
}

func (c *ChatClient) sendNormalMessage(text string) {
	msg := protocol.Message{
		Type:    protocol.MessageTypeNormal,
		From:    c.user.From,
		Content: text,
		Time:    time.Now().Unix(),
		Id:      c.user.Id,
	}

	if err := c.SendMessage(msg); err != nil {
		fmt.Printf("%s消息发送失败: %v%s\n", colorRed, err, colorReset)
	}
}

func (c *ChatClient) receiveMessages() {
	defer c.wg.Done()

	reader := bufio.NewReader(c.conn)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := protocol.DecodeMessage(reader)
			if err != nil {
				fmt.Printf("%s连接断开: %v%s\n", colorRed, err, colorReset)
				c.Shutdown()
				return
			}
			c.displayMessage(msg)
		}
	}
}

func (c *ChatClient) displayMessage(msg protocol.Message) {
	timestamp := time.Unix(msg.Time, 0).Format("15:04:05")

	switch msg.Type {
	case protocol.MessageTypeJoin:
		fmt.Printf("%s[%s][系统] %s <%s> 加入了聊天室%s\n",
			colorYellow, timestamp, msg.From, msg.Id, colorReset)

	case protocol.MessageTypeLeave:
		fmt.Printf("%s[%s][系统] %s <%s> 离开了聊天室%s\n",
			colorYellow, timestamp, msg.From, msg.Id, colorReset)

	case protocol.MessageTypeNormal:
		color := colorBlue
		prefix := fmt.Sprintf("%s(%s)", msg.From, msg.Id)
		if msg.Id == c.user.Id {
			color = colorGreen
			prefix = "您"
		}
		fmt.Printf("%s[%s][%s]%s %s\n",
			color, timestamp, prefix, colorReset, msg.Content)

	case protocol.MessageTypeSystem:
		fmt.Printf("%s[%s][系统] %s%s\n",
			colorYellow, timestamp, msg.Content, colorReset)

	case protocol.MessageTypeRank:
		fmt.Printf("\n%s%s%s\n", colorCyan, msg.Content, colorReset)
	}
}

func (c *ChatClient) sendHeartbeat() {
	defer c.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			msg := protocol.Message{
				Type:    protocol.MessageTypeHeartbeat,
				From:    c.user.From,
				Id:      c.user.Id,
				Content: "PING",
				Time:    time.Now().Unix(),
			}

			if err := c.SendMessage(msg); err != nil {
				fmt.Printf("%s心跳发送失败: %v%s\n", colorRed, err, colorReset)
				c.Shutdown()
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *ChatClient) SendMessage(msg protocol.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 设置消息时间戳
	msg.Time = time.Now().Unix()

	// 更新活跃度
	if msg.Type == protocol.MessageTypeNormal {
		c.updateActivity()
	}

	data, err := protocol.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("编码失败: %w", err)
	}

	// 设置写入超时
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("发送失败: %w", err)
	}

	return nil
}

func (c *ChatClient) updateActivity() {
	c.lastActivityTime = time.Now()
	c.activityScore++
}

func (c *ChatClient) Shutdown() {
	c.cancel()
	close(c.stopChan)

	// 发送离开消息
	if c.conn != nil {
		leaveMsg := protocol.Message{
			Type: protocol.MessageTypeLeave,
			From: c.user.From,
			Id:   c.user.Id,
			Time: time.Now().Unix(),
		}
		c.SendMessage(leaveMsg)
		c.conn.Close()
	}

	c.wg.Wait()
	fmt.Printf("%s已安全退出聊天室%s\n", colorGreen, colorReset)
}

func main() {
	client := NewChatClient()

	// 连接服务器
	if err := client.Connect("localhost:8888"); err != nil {
		fmt.Println(err)
		return
	}

	// 获取用户名
	fmt.Printf("%s请输入用户名：%s", colorCyan, colorReset)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := strings.TrimSpace(scanner.Text())

	// 用户登录
	if err := client.Login(username); err != nil {
		fmt.Printf("%s登录失败: %v%s\n", colorRed, err, colorReset)
		return
	}

	// 启动客户端
	client.Start()
}

func (c *ChatClient) showStats() {
	fmt.Printf(`
%s=== 个人统计信息 ===
用户名: %s
ID: %s
活跃度: %.0f
最后活跃: %s
%s`, colorCyan, c.user.From, c.user.Id, c.activityScore,
		c.lastActivityTime.Format("2006-01-02 15:04:05"), colorReset)
}
