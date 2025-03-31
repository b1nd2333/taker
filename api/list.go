package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"taker/common"
	"time"
)

type ListResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		AssignmentId   int         `json:"assignmentId"`
		Title          string      `json:"title"`
		Describe       string      `json:"describe"`
		Url            string      `json:"url"`
		Done           bool        `json:"done"`
		AssignmentType string      `json:"assignmentType"`
		Reward         string      `json:"reward"`
		Project        string      `json:"project"`
		Logo           string      `json:"logo"`
		Complete       interface{} `json:"complete"`
		Top            bool        `json:"top"`
		Timestamp      interface{} `json:"timestamp"`
		CompleteTime   *string     `json:"completeTime"`
		Sort           int         `json:"sort"`
		CfVerify       bool        `json:"cfVerify"`
	} `json:"data"`
}

// 获取所有任务
func list(taker Taker, token string) error {
	proxyAddress := "socks5://" + taker.Username + ":" + taker.Password + "@" + taker.IP + ":" + taker.Port

	// 创建 HTTP 客户端
	client, err := newHTTPClientWithProxy(proxyAddress)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://lightmining-api.taker.xyz/assignment/list", nil)
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
	var responseData ListResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查返回的 nonce 是否存在
	if responseData.Code != 200 {
		return fmt.Errorf("get time response format: %v", responseData)
	}

	for _, v := range responseData.Data {
		// 每日任务
		if v.Done == false && v.AssignmentId == 13 {
			// 完成每日任务
			time.Sleep(3 * time.Second)
			err = task(taker, token, v.AssignmentId)
			if err != nil {
				//fmt.Printf("地址%s完成任务%s失败\n", taker.Address, v.Title)
				return err
			}
		}
	}

	return nil
}
