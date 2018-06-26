package connpool

import (
	"errors"
	"fmt"
	"net"
)

type MachineConn struct {
	MConn net.Conn
	Key   string
	Id    int64 //当前连接id
}

//获取ConnnPool实例
func ConnPoolNew(conn net.Conn, key string, id int64) *MachineConn {
	connPool := &MachineConn{MConn: conn, Key: key, Id: id}
	return connPool
}

var ConnPoolActualStorage map[string]*MachineConn

//跨包访问ConnPoolActualStorage（tcp连接池实例）
func GetConn(did string) (net.Conn, string, error) {
	conn, ok := ConnPoolActualStorage[did]

	fmt.Println("ConnPoolActualStorage:", ConnPoolActualStorage)
	if ok != true {
		fmt.Println("did:", did)
		return nil, "", errors.New("机器当前不在线")
	}
	return conn.MConn, conn.Key, nil
}

func FreeConn(did string) {
	if _, ok := ConnPoolActualStorage[did]; ok == true {
		//链接池中存在该连接，删除该元素

		delete(ConnPoolActualStorage, did)
	}
}
