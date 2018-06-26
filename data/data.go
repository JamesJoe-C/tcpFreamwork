package data

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"reflect"

	"tcp.fresh.com/mylog"

	connpool "tcp.fresh.com/connpool"
	"tcp.fresh.com/core"
)

type TcpData struct {
	request_tree  map[int]RequestInterface
	response_tree map[int]ResponseInterface
}

//路由注册
func Register() (s *TcpData) {

	//tcp请求发起
	s = &TcpData{request_tree: make(map[int]RequestInterface), response_tree: make(map[int]ResponseInterface)}
	s.request_tree[HEARTBEAT] = Heartbeat_request{}

	s.response_tree[HEARTBEAT] = &Heartbeat_response{}
	s.response_tree[MACHINE_SET_PARAM] = &MachineSetParamResponse{}
	return s
}

//协议解析
func DataResolve(conn net.Conn, key string) (map[string]interface{}, error, []interface{}) {

	//解析协议头：

	/*
		1.消息长度：
			数据包长度：0-3
			标注整个数据包长度，包括数据包头和数据包内容，范围8~UINT32_MAX
	*/
	var log_param []interface{} = make([]interface{}, 26)

	data := make([]byte, 4)

	//从conn中读取数据
	_, err := conn.Read(data)
	//当长度为0，保持返回值为空，保持线程不断开
	if core.BytesToInt(data) == 0 {
		return nil, nil, log_param
	}
	if err != nil {
		return nil, err, log_param
	}

	data_len_yuan := core.BytesToInt(data)
	data_len := data_len_yuan - 8
	if data_len > 65535 {
		return nil, errors.New("消息长度超过最大允许范围:" + string(data_len)), log_param
	}

	/*
		2.消息类型：
			数据包长度：4-5
			标注该消息类型
			0x0001：DES加密后经过BASE64编码的JSON数据
	*/
	data_type := make([]byte, 2)

	//从conn中读取数据
	_, err2 := conn.Read(data_type)
	if err2 != nil {
		return nil, err2, log_param
	}

	chkError(err2)

	/*
		3.数据包校验
			数据包长度: 6-7
			采用和校验，校验范围包括整个数据包0-N字节，主要验证报文完整性，一旦出现校验错误需要断开连接，重新尝试连接再传输数据
	*/
	request_check := make([]byte, 2)
	//从conn中读取数据
	_, err3 := conn.Read(request_check)
	if err3 != nil {
		return nil, err3, log_param
	}

	chkError(err3)

	/*
		4.数据内容
			数据包长度：8-N
			任意字节数据
	*/

	request_data := make([]byte, data_len)
	//从conn中读取数据
	_, err4 := conn.Read(request_data)
	if err4 != nil {
		return nil, err4, log_param
	}
	chkError(err4)
	log_param[18] = "第四步读取的数据反序列化：" + string(request_data)

	///////////////////////////////////数据内容解密/////////////////////////

	//DES解密
	log_param[20] = "解密key为：" + key
	request_data_json := core.Undes(request_data, key)

	log_param[22] = "DES解密后数据包：" + request_data_json

	//数据包校验
	if !core.DataCheck(request_check, data, data_type, request_data, data_len) {
		return nil, errors.New("数据包校验不匹配"), log_param
	}
	// ////////////////////////////////////////////根据包体中传递的res（请求类型）来选择接收结构体///////////////////////////////////////////

	//请求json包解析
	request := dataForMap(request_data_json)
	log_param[24] = "请求包解析结果："
	log_param[25] = request

	return request, nil, log_param
}

//机器密钥获取（3des解密）
func machine_key_check(key string) (string, error) {
	//这里的解密密钥为共同约定的密钥
	machine_key := core.Undes([]byte(key), "c82d2e48")
	machine_key_byte := []byte(machine_key)
	//获取机器密钥的头两位
	machine_key_check := machine_key_byte[:2]
	//获取机器密钥的真实值
	machine_key_byte = machine_key_byte[2:]

	//fmt.Println("", string(machine_key_check))
	if string(machine_key_check) != "P:" {
		return "", errors.New("Machine password is incorrectly formatted")
	}
	machine_key = string(machine_key_byte)
	return machine_key, nil
}

//请求类型映射到对应结构体,启动request处理相应请求（业务层）
func AppStart(c *Client, tcp_data *TcpData, data map[string]interface{}, key string) {
	conn := c.Conn()
	did := c.Did

	//////////////////////////////////////////////////////将机器与tcp连接关联起来///////////////////////////////////////////////////////
	if _, ok := data["req"]; ok && data["req"] != nil && data["req"].(float64) == 1 {
		//客户端上线成功记录
		Machine_online(c)
		connpool.ConnPoolActualStorage[did] = connpool.ConnPoolNew(conn, key, c.Id)
	}
	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	//机器id放入数据包中
	data["did"] = did

	var log_param []interface{} = make([]interface{}, 26)

	log_param[1] = data

	var re int
	var request_or_respsone interface{}

	//判断是request还是respson
	if _, ok := data["req"]; ok {
		//请求类型信息断言
		re = int(data["req"].(float64))
		//映射到请求响应结构体
		request_or_respsone = tcp_data.request_tree[re]

	} else if _, ok := data["res"]; ok {
		//请求类型信息断言
		re = int(data["res"].(float64))
		//映射到请求响应结构体
		request_or_respsone = tcp_data.response_tree[re]
	} else {
		panic("req，res均为空，无法解析")
	}
	log_param[2] = "响应结构体：" + string(re)
	log_param[3] = request_or_respsone

	//结构体实例获取
	object := reflect.ValueOf(request_or_respsone)
	//调用结构体内方法,通过反射调用赋值方法对结构体进行赋值
	//获取结构体方法指针
	v := object.MethodByName("RunStart")
	//传递给方法的参数
	args := []reflect.Value{reflect.ValueOf(c), reflect.ValueOf(data)}
	//调用方法
	response := v.Call(args)

	//调用方法拿到的返回值
	response_string := response[0].String()

	if response_string == "" {
		return
	}
	//机器sn码累加
	SetMachineSn(did, int(data["sn"].(float64)))

	log_param[4] = ("拿到的response(返回给机器的数据):" + response_string)

	//这里用c82d2e48加密，因为通讯协议所有发送给机器的数据，全都用约定的key来加密

	response_des := core.Des(response_string, key)

	log_param[5] = "加密后发送给客户端的数据:" + response_des
	log_param[6] = "key为：" + key
	mylog.ToLog[c.Did].AddLog(log_param)
	//数据发送给客户端
	send([]byte(response_des), conn, did)
}

//发送数据到机器
func send(data []byte, conn net.Conn, did string) error {
	data_len := len(data) + 8
	data_send := make([]byte, data_len, data_len)
	//消息长度
	data_send[0] = 0
	data_send[1] = 0
	data_send[2] = 0
	data_send[3] = byte(data_len)
	//消息类型
	data_send[4] = 0
	data_send[5] = 1

	for k, v := range data {
		data_send[k+8] = v
	}

	dataCheck_byte := core.Sum_data(data_send)
	data_send[6] = dataCheck_byte[0]
	data_send[7] = dataCheck_byte[1]

	_, err := conn.Write(data_send)

	return err
}

//数据包解析
func dataForMap(data string) (dat map[string]interface{}) {
	//log.Println("进行json解包的data值：", data)

	if err := json.Unmarshal([]byte(data), &dat); err != nil {
		log.Println("数据包内容非json结构: ", data, []byte(data))
		panic(err)
	}
	return dat
}

//报错处理
func Error(err string) {
	log.Println(err)
	//return err
}

//客户端上线的特殊处理
func Machine_online(c *Client) {

}

//客户端离线的特殊处理
func Machine_unline(c *Client) {

}
