package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"taker/common"
)

type TaskReq struct {
	AssignmentId int `json:"assignmentId"`
}

type TaskResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}

func task(taker Taker, token string, taskId int) error {
	taskReq := TaskReq{}
	taskReq.AssignmentId = taskId

	jsonData, err := json.Marshal(taskReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	proxyAddress := "socks5://" + taker.Username + ":" + taker.Password + "@" + taker.IP + ":" + taker.Port

	// 创建 HTTP 客户端
	client, err := newHTTPClientWithProxy(proxyAddress)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://lightmining-api.taker.xyz/assignment/do", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range common.GetHeaders() {
		req.Header.Set(key, value)
	}
	req.Header.Add("authorization", "Bearer "+token)
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get nonce. Status code: %d", resp.StatusCode)
	}

	// 读取响应数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// 解析 JSON 响应
	var responseData TaskResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查返回的 nonce 是否存在
	if responseData.Code != 200 {
		return fmt.Errorf("get time response format: %v", responseData)
	}

	return nil
}
