package data

import (
	"errors"
	"log"
	"net"
	"sync/atomic"
	"time"

	connpool "tcp.fresh.com/connpool"

	ml "tcp.fresh.com/mylog"
)

// Client holds info about connection
type Client struct {
	Id     int64  //唯一标识
	Did    string //机器id
	conn   net.Conn
	Server *server
}

// TCP server
type server struct {
	timeout                  int
	address                  string // Address to open connection: localhost:9999
	onNewClientCallback      func(c *Client)
	onClientConnectionClosed func(c *Client, err error)
	onNewMessage             func(c *Client, message map[string]interface{}, key string)
	ClientId                 int64
}

// Read client data from channel
func (c *Client) listen() {
	//机器首次连接服务器的加密key
	password_key := "1234bcda"
	for {
		//协议解析
		message, err, mylog := DataResolve(c.conn, password_key)

		if _, ok := message["did"]; ok {
			c.Did = message["did"].(string)
		}

		//数据包解析错误，直接断开连接
		if err != nil {
			log.Println("tcpServer error：", err)
			c.Close(err)
			return
		}

		//根据conn.write判断tcp连接超时,断开连接，发起断开连接回调
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			connpool.FreeConn(c.Did)

			log.Println("tcpServer error：", err)
			c.Close(err)
			return
		}

		//当消息为空时，代表当前连接在超时时间内没有发送任何数据包，此时则认为该链接已经超时
		if message == nil {
			//log.Println("客户端无数据传递，已经超时，断开连接")
			c.Close(errors.New("客户端无数据传递，已经超时，断开连接"))
			return
		}

		//查看机器在日志系统中是否已经注册，未注册则注册
		if _, ok := ml.ToLog[c.Did]; !ok {
			ml.ToLog[c.Did] = ml.NewMyLog()
		}
		ml.ToLog[c.Did].AddLog(mylog)
		//在机器上线请求时，获取设备控制密钥，并存储
		if req, ok := message["req"].(float64); ok && req == 1 {
			password_key = GetMachineKey(c.Did)
			if password_key == "" {
				//这里表示数据库中没有该机器的密钥，或数据库异常.所以这里要断开连接
				c.Close(errors.New("数据库中没有该机器的密钥，或数据库异常"))
				return
			}
		}
		//////////////////
		c.Server.onNewMessage(c, message, password_key)
	}
}

//set tcp conn time out
func (c *Client) SetTimeOut(timeout int) {
	c.conn.SetDeadline(time.Now().Add(time.Second * time.Duration(timeout)))
}

// Send text message to client
func (c *Client) Send(message string) error {
	_, err := c.conn.Write([]byte(message))
	return err
}

// Send bytes to client
func (c *Client) SendBytes(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (c *Client) Close(err error) error {
	c.Server.onClientConnectionClosed(c, err)
	return c.conn.Close()
}

// Called right after server starts listening new client
func (s *server) OnNewClient(callback func(c *Client)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *server) OnClientConnectionClosed(callback func(c *Client, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Client receives new message
func (s *server) OnNewMessage(callback func(c *Client, message map[string]interface{}, key string)) {
	s.onNewMessage = callback
}

// Start network server
func (s *server) Listen() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		client := &Client{
			conn:   conn,
			Server: s,
			Id:     atomic.AddInt64(&s.ClientId, 1),
		}

		go client.listen()
		s.onNewClientCallback(client)
	}
}

// Creates new tcp server instance
func New(address string) *server {
	log.Println("Creating server with address", address)
	//在服务器启动时，实例化tcp连接存储池
	connpool.ConnPoolActualStorage = make(map[string]*connpool.MachineConn)
	server := &server{
		address:  address,
		ClientId: 0,
	}

	server.OnNewClient(func(c *Client) {})
	server.OnNewMessage(func(c *Client, message map[string]interface{}, key string) {})
	server.OnClientConnectionClosed(func(c *Client, err error) {})

	return server
}
