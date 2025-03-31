package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"taker/common"
)

type TimeResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		LastMiningTime  int `json:"lastMiningTime"`
		TotalMiningTime int `json:"totalMiningTime"`
	} `json:"data"`
}

// 获取签到时间
func totalMiningTime(taker Taker, token string) (int, error) {
	proxyAddress := "socks5://" + taker.Username + ":" + taker.Password + "@" + taker.IP + ":" + taker.Port

	// 创建 HTTP 客户端
	client, err := newHTTPClientWithProxy(proxyAddress)
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP client: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("GET", "https://lightmining-api.taker.xyz/assignment/totalMiningTime", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range common.GetHeaders() {
		req.Header.Set(key, value)
	}
	req.Header.Add("authorization", "Bearer "+token)
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get nonce. Status code: %d", resp.StatusCode)
	}

	// 读取响应数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	// 解析 JSON 响应
	var responseData TimeResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查返回的 nonce 是否存在
	if responseData.Code != 200 {
		return 0, fmt.Errorf("get time response format: %v", responseData)
	}

	return responseData.Data.LastMiningTime, nil
}
