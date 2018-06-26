package mylog

import (
	"fmt"
	"time"
)

type MyLog struct {
	LogMap []interface{}
}

//用于存储日志信息
var ToLog map[string]*MyLog
var Log_enable bool

func NewMyLog() *MyLog {
	return &MyLog{}
}

//添加日志到存储器
func (l *MyLog) AddLog(pararm []interface{}) {
	for _, v := range pararm {
		if v != nil {
			l.LogMap = append(l.LogMap, v)
			// fmt.Println(v)
		}
	}
	// fmt.Println("l.logmap:", l.LogMap)
}

//打印日志
func Print(did string) {
	if !Log_enable {
		return
	}

	logTime := time.Unix(1389058332, 0).Format("2006-01-02 15:04:05")
	start_str := "------------------------------------------------------------------" + logTime + "  " + did + "------------------------------------------------------------------"
	end_str := ("end " + did)

	str_len := len(start_str)
	i_ := (str_len - len(did) - 4) / 2
	for i := 0; i < i_; i++ {
		end_str = "-" + end_str
		end_str += "-"
	}

	fmt.Printf("\n %c[1;37;41m%s%c[0m\n", 0x1B, start_str, 0x1B)
	for _, v := range ToLog[did].LogMap {
		if v == "" || v == nil {
			continue
		}
		//其中0x1B是标记，[开始定义颜色，1代表高亮，40代表黑色背景，32代表绿色前景，0代表恢复默认颜色
		fmt.Printf("\n %c[1;40;32m%v%c[0m\n", 0x1B, v, 0x1B)
	}
	fmt.Printf("\n %c[1;37;41m%s%c[0m\n", 0x1B, end_str, 0x1B)
}

//释放日志
func FreeMyLog(did string) {
	ToLog[did].LogMap = make([]interface{}, 0, 10)
}
