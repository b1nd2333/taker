package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"taker/common"
	"time"
)

type MintResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

// 开始签到
func startMining(taker Taker, token string) error {
	proxyAddress := "socks5://" + taker.Username + ":" + taker.Password + "@" + taker.IP + ":" + taker.Port

	// 创建 HTTP 客户端
	client, err := newHTTPClientWithProxy(proxyAddress)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://lightmining-api.taker.xyz/assignment/startMining", nil)
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
	var responseData MintResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查返回的 nonce 是否存在
	//fmt.Println(responseData)
	if responseData.Code != 200 {
		if strings.Contains(responseData.Msg, "The contract is not activated") || strings.Contains(responseData.Msg, "startMining Fail, startMining has not reached refresh time") { // 激活节点
			err = activated(taker) // 激活
			if err != nil {
				return err
			}
			time.Sleep(3 * time.Second)
			err = startMining(taker, token)
			return err
		}
		return fmt.Errorf("startMint response format: %v", responseData)
	}

	return nil
}
