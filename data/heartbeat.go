package data

import (
	"encoding/json"
	"strconv"
)

//心跳包请求接受
type Heartbeat_request struct {
	//请求类型:2
	request
}

//心跳包请求回复
type Heartbeat_response struct {
	//响应类型:2
	response
}

///实现Data接口，返回数据包内容
func (f Heartbeat_request) getData() string {
	//结构体转换为json
	json, err := json.Marshal(f)
	if err != nil {
		panic(err)
	}
	return string(json)
}

//实现Data接口,返回数据包类型
func (f Heartbeat_request) getType() int {
	return 2
}

//标识此结构体为Request
func (f Heartbeat_request) ThisIsRequest() {
}

//实现request接口
func (f Heartbeat_request) RunStart(c *Client, data map[string]interface{}) string {

	sn := strconv.FormatFloat(data["sn"].(float64), 'f', -1, 64)

	hrs := Heartbeat_response{}
	response := hrs.getData(sn)
	return response
}

//////////////////////////////////////////////////////////////下方为response//////////////////////////////////////////////////////////////////////////////
//发送数据包
/*
 */
func (f Heartbeat_response) getData(param ...string) string {
	f.Res = 2
	sn, err := strconv.Atoi(param[0])
	chkError(err)
	f.Sn = sn
	data := dataFormat(f)
	return data
}

//标识此结构体为Response
func (f Heartbeat_response) ThisIsResponse() {
}
