package utillib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func SendPostRequest(url string, data string) string {
	var jsonStr = []byte(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("SendPostRequest->error:", err.Error())
		return err.Error()
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
