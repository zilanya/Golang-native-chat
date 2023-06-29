package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	Ip   string
	Port int
	Name string
	conn net.Conn
	flag int //当前client模式
}

func NewClient(Ip string, Port int) *Client {

	//创建对象
	client := &Client{
		Ip:   Ip,
		Port: Port,
		flag: 999,
	}

	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", Ip, Port))
	if err != nil {
		fmt.Println("net.Dial error!")
		return nil
	}

	client.conn = conn
	return client
}

var serverIp string
var serverPort int

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公共聊天")
	fmt.Println("2.私聊模式")
	fmt.Println("3.修改名字")
	fmt.Println("0.退出聊天")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>违法数字<<<<<")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入名字...")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"

	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("con error is :", err)
		return false
	}

	return true
}

// 处理server回应的消息，直接显示到标准输出即可
func (client *Client) DealResponse() {
	//一旦监听到client.conn有数据,就直接copy嗷stdout标准输出，永久阻塞监听
	io.Copy(os.Stdout, client.conn)

}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}
		//根据不同模式，处理不同业务
		switch client.flag {
		case 1:
			//公共聊天
			client.PublicChat()
			break
		case 2:
			//私聊
			client.PrivateChat()
			break
		case 3:
			//改名
			client.UpdateName()
			break
		}
	}
}

func (client *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("con error is", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteUser string
	var chatMsg string

	client.SelectUser()
	fmt.Println("请输入私聊用户姓名，exit退出")
	fmt.Scanln(&remoteUser)

	if remoteUser != "exit" {
		fmt.Println("请输入私聊内容")
		fmt.Scanln(&chatMsg)

		if chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to" + "|" + remoteUser + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("con error is", err)

				}
			}

			chatMsg = ""
			fmt.Println("请输入私聊内容")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUser()
		fmt.Println("请输入私聊用户姓名，exit退出")
		fmt.Scanln(&remoteUser)
	}
}

func (client *Client) PublicChat() {
	//提示发消息
	var chatMsg string

	fmt.Println("请输入消息")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("con error is", err)
				break
			}
		}

		chatMsg := ""
		fmt.Println("请输入消息")
		fmt.Scanln(&chatMsg)
	}

}
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "这是ip，默认为127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "这是端口号，默认为8888")
}
func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>连接服务器失败")
		return
	}

	go client.DealResponse()

	fmt.Println(">>>>>>>连接服务器成功")

	//启动业务
	client.Run()

}
