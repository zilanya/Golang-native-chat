package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建一个用户的api
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	//启动监听当前user channel的gourtine
	go user.ListenMessage()
	return user
}

// 监听当前User  channel的方法，一旦有消息，就直接发送
func (this *User) ListenMessage() {

	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}

}

func (this *User) Online() {

	//用户上线，将用户加入到OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

func (this *User) Offline() {

	//用户下线，将用户从onlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户下线消息
	this.server.BroadCast(this, "已下线")

}

// 给当前user对应的客户端发消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 查询在线用户、改名、聊天
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户都有谁
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线....\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//修改名字
		newName := strings.Split(msg, "|")[1]

		//判断是否存在
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.sendMsg("已被使用")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.sendMsg("修改成功,用户名:" + this.Name)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式 : to|user|消息内容

		//1获取用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.sendMsg("消息格式不正确\n")
			return
		}

		//2根据用户名 得到对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.sendMsg("私聊用户不存在\n")
			return
		}

		//3发送私聊内容
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.sendMsg("内容不正确，请重新发送私聊内容\n")
			return
		}

		remoteUser.sendMsg(this.Name + "私聊你说：" + content)
	} else {
		this.server.BroadCast(this, msg)
	}

}
