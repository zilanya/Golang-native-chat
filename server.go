package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

// 创建server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听message广播消息channel的goroutine。一旦有消息就全部发给在线user
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		//将msg发给全部在线user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()

	}
}

func (this *Server) Handler(conn net.Conn) {
	// 当前链接业务
	//fmt.Println("连接成功")

	user := NewUser(conn, this)

	//用户上线
	user.Online()

	//监听活跃channel
	isLive := make(chan bool)

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				//用户下线
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户消息去除\n
			msg := string(buf[:n-1])

			//针对用户对msg进行处理
			user.DoMessage(msg)

			//用户发送任何消息都代表活跃
			isLive <- true
		}
	}()

	//当前handle阻塞
	for {
		select {
		case <-isLive:
			//代表当前用户处于活跃，不做任何操作，激活select，重置下方定时器

			//10秒踢人
		case <-time.After(time.Second * 1000):
			//已经超时，踢出用户
			user.sendMsg("您已被移除聊天")

			//消除user资源
			close(user.C)

			//关闭连接
			conn.Close()

			//退出当前handle
			runtime.Goexit()
		}
	}
}

// 启动服务器的接口
func (this *Server) Start() {
	// 监听端口
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// 关闭监听
	defer listen.Close()

	//启动监听message的goroutine
	go this.ListenMessager()

	for {
		// 连接
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("list accept err:", err)
			continue
		}

		// 处理链接
		go this.Handler(conn)
	}

}
