package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/fatih/set"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// 消息
type Message struct {
	gorm.Model
	FormId   int64  // 发送者
	TargetId int64  // 接收者
	Type     int    // 发送类型 群聊 私聊 广播
	Media    int    // 消息类型 文字 图片 音频
	Content  string // 消息内容
	Pic      string
	Url      string
	Desc     string
	Amount   int // 其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

// 发送消息 需要 发送者ID、接收者ID、消息类型、消息内容
// 接收消息
type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

func Chat(wrtier http.ResponseWriter, request *http.Request) {
	// 1. 获取参数 并且 校检 token 等合法性
	// token := query.Get("token")
	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	// msgType := query.Get("type")
	// targetId := query.Get("targetId")
	// context := query.Get("context")
	isvalidate := true // checkToken() 待实现
	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := (&websocket.Upgrader{
		// token 校检
		CheckOrigin: func(r *http.Request) bool {
			return isvalidate // 允许跨域请求
		},
	}).Upgrade(wrtier, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 2. 获取连接conn
	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}
	// 3. 用户关系
	// 4. userid 跟 node 绑定 并且加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()
	// 5. 完成发送逻辑
	go sendProc(node)
	// 6. 完成接受逻辑
	go recvProc(node)
	sendMsg(userId, []byte("欢迎来到聊天室!")) // 发送欢迎消息
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

}

func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		broadMsg(data)
		fmt.Println("[ws] <<<<<< ", data)
	}
}

var upsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	upsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP: net.IPv4(192, 168, 0, 255),
		// IP:   net.IPv4(172, 28, 80, 1),
		Port: 3000,
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		select {
		case data := <-upsendChan:
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 完成udp数据接收协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close()
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := &Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case 1: // 私信
		sendMsg(msg.TargetId, data)
		// case 2: // 群发
		// 	sendGroupMsg()
		// case 3: // 广播
		// 	sendAllMsg()
		// default:
	}
}

func sendMsg(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	if ok {
		node.DataQueue <- msg
	} else {
		fmt.Println("用户未在线，无法发送消息")
		return
	}
}
