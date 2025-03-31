package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net/http"
	"net/url"
	"taker/common"
	"time"
)

type LoginRequest struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

type LoginResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Token    string `json:"token"`
		IsInvite bool   `json:"isInvite"`
		ErrMsg   string `json:"errMsg"`
	} `json:"data"`
}

// 用于生成 Nonce 的请求和响应结构体
type NonceRequest struct {
	WalletAddress string `json:"walletAddress"`
}

type NonceResponse struct {
	Data struct {
		Nonce string `json:"nonce"`
	} `json:"data"`
}

// 获取登录需要message
func generateNonce(taker Taker) (string, error) {
	nonceRequest := NonceRequest{
		WalletAddress: taker.Address,
	}
	jsonData, err := json.Marshal(nonceRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %v", err)
	}
	proxyAddress := "socks5://" + taker.Username + ":" + taker.Password + "@" + taker.IP + ":" + taker.Port

	// 创建 HTTP 客户端
	client, err := newHTTPClientWithProxy(proxyAddress)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://lightmining-api.taker.xyz/wallet/generateNonce", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range common.GetHeaders() {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get nonce. Status code: %d", resp.StatusCode)
	}

	// 读取响应数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// 解析 JSON 响应
	var responseData NonceResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查返回的 nonce 是否存在
	if responseData.Data.Nonce == "" {
		return "", fmt.Errorf("invalid nonce response format: %v", responseData)
	}

	return responseData.Data.Nonce, nil
}

// 登录
func login(taker Taker, message string, signMessage string) (string, error) {
	//uri := "https://lightmining-api.taker.xyz/wallet/login"
	loginRequest := LoginRequest{}
	loginRequest.Address = taker.Address
	loginRequest.Message = message
	loginRequest.Signature = signMessage

	jsonData, err := json.Marshal(loginRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %v", err)
	}
	proxyAddress := "socks5://" + taker.Username + ":" + taker.Password + "@" + taker.IP + ":" + taker.Port

	// 创建 HTTP 客户端
	client, err := newHTTPClientWithProxy(proxyAddress)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://lightmining-api.taker.xyz/wallet/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range common.GetHeaders() {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get nonce. Status code: %d", resp.StatusCode)
	}

	// 读取响应数据
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// 解析 JSON 响应
	var responseData LoginResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查返回的 nonce 是否存在
	if responseData.Data.Token == "" {
		return "", fmt.Errorf("invalid token response format: %v", responseData)
	}

	return responseData.Data.Token, nil

}

// 创建支持 SOCKS5 代理和身份验证的 HTTP 客户端
func newHTTPClientWithProxy(proxyAddress string) (*http.Client, error) {
	// 解析代理地址
	proxyURL, err := url.Parse(proxyAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy address: %v", err)
	}

	// 设置 SOCKS5 代理并进行身份验证
	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %v", err)
	}

	// 创建 HTTP Transport 使用 SOCKS5 代理
	transport := &http.Transport{
		Dial: dialer.Dial,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 忽略 HTTPS 错误
		},
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	return client, nil
}

// 完成任务
//func performTasks(taker Taker, token string) bool {
//	// 需要完成的任务id
//	proxyAddress := "socks5://" + taker.Username + ":" + taker.Password + "@" + taker.IP + ":" + taker.Port
//
//	// 创建 HTTP 客户端
//	client, err := newHTTPClientWithProxy(proxyAddress)
//	if err != nil {
//		return fmt.Errorf("failed to create HTTP client: %v", err)
//	}
//
//	// 创建请求
//	req, err := http.NewRequest("POST", "https://lightmining-api.taker.xyz/assignment/startMining", nil)
//	if err != nil {
//		return fmt.Errorf("failed to create request: %v", err)
//	}
//
//	for key, value := range getHeaders() {
//		req.Header.Set(key, value)
//	}
//	req.Header.Add("authorization", "Bearer "+token)
//	// 发送请求
//	resp, err := client.Do(req)
//	if err != nil {
//		return fmt.Errorf("request failed: %v", err)
//	}
//	defer resp.Body.Close()
//
//	// 检查响应状态码
//	if resp.StatusCode != http.StatusOK {
//		return fmt.Errorf("failed to get nonce. Status code: %d", resp.StatusCode)
//	}
//
//	// 读取响应数据
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return fmt.Errorf("failed to read response body: %v", err)
//	}
//
//	// 解析 JSON 响应
//	var responseData MintResponse
//	err = json.Unmarshal(body, &responseData)
//	if err != nil {
//		return fmt.Errorf("failed to unmarshal response: %v", err)
//	}
//
//	// 检查返回的 nonce 是否存在
//	if responseData.Code != 200 {
//		return fmt.Errorf("get time response format: %v", responseData)
//	}
//
//	return nil
//}
