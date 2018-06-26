package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

//字节转换成整形
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}

func testBinaryWrite(x interface{}) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, x)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	//fmt.Printf("% x\n", buf.Bytes())
	return buf.Bytes()
}
