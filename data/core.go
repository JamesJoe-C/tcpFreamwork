package data

import (
	"fmt"
	"log"

	"encoding/json"

	"tcp.fresh.com/core"
)

const (
	_ = iota
	FIRST_DATA
	HEARTBEAT
	MACHINE_INFO
	MACHINE_SET_PARAM
	MACHINE_SEND_INFO
	MACHINE_FIRMWARE_UPDATE
)

//获取数据包内容与类型接口
type Data interface {
	GetData() string //获取数据包体内容
	GetType() int    //获取数据包类型
}

//发送数据包，设置数据包
type SendData interface {
	SetData(data map[string]interface{}) bool
}

//标识出发送包
type RequestInterface interface {
	ThisIsRequest()
}

//标识出响应包
type ResponseInterface interface {
	ThisIsResponse()
}

type request struct {
	Req int `json:"req"` //请求类型
	Sn  int `json:"sn"`  //请求序号:0 ~ 2147483647 数值递增，用于响应同步
}

type response struct {
	Res int `json:"res"` //响应类型
	Sn  int `json:"sn"`  //响应序号:对应请求序号
}

//机器操控--rpc方法提供
type MachineRPC struct {
}

type Args struct {
	//调用的请求名
	RequestId int
	//调用的请求内参数,json格式
	RequestArgs string
}

type machineid struct {
	Id string `json:"-"` //机器id
}

/*
	args  传递的参数，json格式
	reply 返回的值,同样json格式
*/
func (t *MachineRPC) Multiply(args Args, reply *string) error {
	fmt.Println("args=", args)
	fmt.Println("reply=", reply)
	reply_str := "this is MachineRPC.Multiply "
	reply = &reply_str
	return nil
}

//协议包封装。（目前为json协议，故结构体转json）
func dataFormat(a_struct interface{}) (jsonp string) {
	jsons, errs := json.Marshal(a_struct) //转换成JSON返回的是byte[]
	chkError(errs)
	jsonp = string(jsons)
	return jsonp
}

//统一报错处理
func chkError(err error) {
	if err != nil {
		fmt.Println("chkError:", err)
		//log.Fatal(err)
		//panic(err)
	}
}

//统一报错处理
func chkErrorPanic(err error) {
	if err != nil {
		//fmt.Println(err)
		log.Fatal(err)
		panic(err)
	}
}

type NullSn struct {
	Sn    int
	Valid bool // Valid is true if Time is not NULL
}

//获取机器的SN码值
//@param did 机器编号
//@param requestId 请求编号
func GetMachineSn(did string) (sn int) {

	db := core.GetDb()
	defer core.FreeDb(db)

	//fmt.Println("GetMachineSn-> did:", did)

	var sn_c interface{}
	//获取机器当前sn（请求编号）

	err := db.QueryRow("select sn from machine where number = ?", did).Scan(&sn_c)
	chkErrorPanic(err)
	if sn_c != nil {
		sn = int(sn_c.(int64))
		// use sn_c.Sn
	}

	//机器SN码累加
	SetMachineSn(did, sn)
	return sn
}

//自定义日志输出
func MyLog(logtype string, param []interface{}) {
	start_str := ("------------------------------------------------------------------" + logtype + "------------------------------------------------------------------")
	end_str := ("end " + logtype)

	str_len := len(start_str)
	i_ := (str_len - len(logtype) - 4) / 2
	for i := 0; i < i_; i++ {
		end_str = "-" + end_str
		end_str += "-"
	}
	if len(param) == 0 || param == nil {
		return
	}
	fmt.Printf("\n %c[1;37;41m%s%c[0m\n", 0x1B, start_str, 0x1B)
	for _, v := range param {
		if v == "" || v == nil {
			continue
		}
		//其中0x1B是标记，[开始定义颜色，1代表高亮，40代表黑色背景，32代表绿色前景，0代表恢复默认颜色
		fmt.Printf("\n %c[1;40;32m%s%c[0m\n", 0x1B, v, 0x1B)
	}
	fmt.Printf("\n %c[1;37;41m%s%c[0m\n", 0x1B, end_str, 0x1B)
}

//设置机器SN码值
//@param did 机器编号
//@param requestId 请求编号
//@param sn 请求次数序号
func SetMachineSn(did string, sn int) bool {
	if sn >= 65535 {
		//sn码的最大值
		sn = 1
	} else {
		sn++
	}
	db := core.GetDb()
	defer core.FreeDb(db)
	stmtIns, err := db.Prepare("update  machine set sn= ? where number = ? ") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close()
	_, err1 := stmtIns.Exec(sn, did)
	if err1 == nil {
		return true
	} else {
		return false
	}
}

//获取机器通讯密钥
func GetMachineKey(did string) string {
	db := core.GetDb()
	defer core.FreeDb(db)

	//fmt.Println("GetMachineSn-> did:", did)

	var pwd interface{}
	//获取机器当前sn（请求编号）
	pwd_str := ""
	err := db.QueryRow("select pwd from machine where number = ?", did).Scan(&pwd)
	//chkErrorPanic(err)
	if err != nil {
		return ""
	}
	if pwd != nil {
		pwd_str = string(pwd.([]uint8))
		// use sn_c.Sn
	}
	return pwd_str
}
