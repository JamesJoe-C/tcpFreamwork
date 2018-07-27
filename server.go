package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"tcpServer/mylog"

	"tcpServer/core"
	"tcpServer/data"
)

func main() {

	mylog.Log_enable = true

	server := data.New(":8900")

	//日志系统初始化
	mylog.ToLog = make(map[string]*mylog.MyLog)

	server.OnNewClient(func(c *data.Client) {
		c.SetTimeOut(35)
		//log.Println("new Client", c)
		time := "【" + time.Now().Format("2006-01-02 15:04:05") + "】"
		fmt.Printf("\n %c[1;34;47m%s%s%v%c[0m\n\n", 0x1B, time, "new Client", c, 0x1B)
	})
	server.OnNewMessage(func(c *data.Client, message map[string]interface{}, key string) {
		//超时时间重置
		c.SetTimeOut(35)

		//log.Println("接收到消息，发送的机器id为：", c.Did)
		tcp_data := data.Register()

		//数据包对应处理器启动
		data.AppStart(c, tcp_data, message, key)
		//打印日志
		mylog.Print(c.Did)
		mylog.FreeMyLog(c.Did)
	})
	server.OnClientConnectionClosed(func(c *data.Client, err error) {
		// connection with client lost
		data.Machine_unline(c)
		log.Println("connection with client lost", c, err)
	})

	////////////////////////////////////////////////////////启动内部rpc服务////////////////////////////////////////////////////////
	machineRPC := new(data.MachineRPC)
	rpc.Register(machineRPC)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8901")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Println("Creating RPCserver with address :8901 ")
	go http.Serve(l, nil)
	//rpcServerStart()
	////////////////////////////////////////////////////////////end/////////////////////////////////////////////////////////////

	//数据库连接池实例化
	core.DbNew(10)

	server.Listen()
}

//内部RPC服务
func rpcServerStart() {
	machineRPC := new(data.MachineRPC)
	rpc.Register(machineRPC)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8901")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Println("Creating RPCserver with address :8901 ")
	go http.Serve(l, nil)

}
