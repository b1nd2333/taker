package common

import (
	"math/rand"
	"time"
)

// 获取随机的 User-Agent
func getRandomUserAgent() string {
	// 模拟随机 User-Agent，使用 github.com/mileusna/useragent 库
	// 你也可以使用自己的随机生成逻辑
	rand.Seed(time.Now().UnixNano())

	// 随机选择一个 User-Agent
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.128 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.85 Safari/537.36",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/54.0",
		"Mozilla/5.0 (Linux; Android 10; Pixel 3 XL Build/QP1A.191105.004) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.136 Mobile Safari/537.36",
	}

	// 随机选择一个 User-Agent
	return userAgents[rand.Intn(len(userAgents))]
}

func GetHeaders() map[string]string {
	// 随机生成 User-Agent
	ua := getRandomUserAgent()

	// 构建并返回请求头
	headers := map[string]string{
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "en-US,en;q=0.9",
		"Content-Type":    "application/json",
		"Origin":          "https://earn.taker.xyz",
		"Referer":         "https://earn.taker.xyz/",
		"User-Agent":      ua,
	}

	return headers
}
