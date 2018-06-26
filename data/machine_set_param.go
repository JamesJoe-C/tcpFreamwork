package data

//设置设备参数
type MachineSetParamRequest struct {
	//请求类型:4
	request
}

//设置设备参数回复
type MachineSetParamResponse struct {
	//响应类型：4
	//设置参数后的响应

}

func (mi MachineSetParamResponse) ThisIsResponse() {
}

//接收到机器回复后，调用的方法
func (mi *MachineSetParamResponse) RunStart(c *Client, data map[string]interface{}) string {
	//将数据存储到数据库中

	return ""
}

//机器回复数据落地到数据库
func (mi *MachineSetParamResponse) save(did string) {

}

/*
*设置机器状态（供rpc调用）
*@param machine_info_request 请求
*@param machine_info_response 机器返回数据
 */
func (t *MachineRPC) SetMachineParam(msprt MachineSetParamRequest, infoId *int) error {

	//数据发送给客户端
	//err := send([]byte(response_des), conn)
	return nil
}
