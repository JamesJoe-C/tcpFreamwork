package core

//数据包校验:和校验
func DataCheck(check, data, data_type, data_s []byte, data_len int) bool {
	check_rs_byte := Sum_data(data, data_type, data_s)

	//fmt.Println("求和结果：", check_rs_byte)

	//fmt.Println("数据包内校验值：", check)
	for k, v := range check_rs_byte {
		if v != check[k] {
			return false
		}
	}
	return true
	//return true
}

//和校验算法
func Sum_data(args ...[]byte) []byte {
	sum := 0
	for _, v := range args {
		for _, v1 := range v {
			sum += int(v1)
		}
	}
	check_rs_byte := testBinaryWrite(int16(sum))
	return check_rs_byte
}
