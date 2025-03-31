package common

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/fatih/color"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
)

type Proxy struct {
	Address     string
	IP          string
	Port        string
	Username    string
	Password    string
	LastRunTime time.Time
}

func InitDB() {
	// 初始化 SQLite 数据库
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		fmt.Println("无法连接到数据库:", err)
		return
	}
	defer db.Close()

	// 创建表
	err = createTable(db)
	if err != nil {
		fmt.Println("创建表失败:", err)
		return
	}

	// 读取文件内容
	account, err := readLines("account.txt")
	if err != nil {
		fmt.Println("读取 proxy.txt 出错:", err)
		return
	}

	//addresses, err := readLines("address.txt")
	//if err != nil {
	//	fmt.Println("读取 address.txt 出错:", err)
	//	return
	//}

	// 插入数据到 SQLite 数据库
	err = insertData(db, account)
	if err != nil {
		fmt.Println("插入数据失败:", err)
		return
	}

	//fmt.Println("数据插入完成！")
	return
}

func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS taker (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		address TEXT NOT NULL,
		privateKey TEXT NOT NULL,
		ip TEXT NOT NULL,
		port TEXT NOT NULL,
		username TEXT,
		password TEXT,
		lastRunTime DATETIME
	);
	`
	_, err := db.Exec(query)
	return err
}

func insertData(db *sql.DB, accounts []string) error {
	if len(accounts) == 0 {
		return fmt.Errorf("account.txt 文件为空")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO taker (address, privateKey, ip, port, username, password, lastRunTime)
		VALUES (?,?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, account := range accounts {
		ip, port, username, password, publicKey, privateKey := parseProxy(account)

		// 检查 address 是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM taker WHERE address = ?", publicKey).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			// 如果 address 已存在，跳过当前条数据
			//fmt.Printf("地址 %s 已存在，跳过插入\n", addrs[0])
			continue
		}

		// 插入新的数据
		_, err = stmt.Exec(publicKey, privateKey, ip, port, username, password, nil)
		if err != nil {
			return err
		}
	}

	// 提交事务
	return tx.Commit()
}

func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func parseProxy(account string) (ip, port, username, password, publicKey, privateKey string) {
	// 假设 proxy 格式为 "ip:port:username:password"
	parts := strings.Split(account, ":")
	ip, port = parts[0], parts[1]
	if len(parts) > 2 {
		username = parts[2]
	}
	if len(parts) > 3 {
		password = parts[3]
	}
	if len(parts) > 4 {
		publicKey = parts[4]
	}

	if len(parts) > 5 {
		privateKey = parts[5]
	}
	return
}

// UpdateLastRunTime 更新代理的 lastRunTime
func UpdateLastRunTime(db *sql.DB, proxyID int, datetimeFormat string) {
	query := "UPDATE taker SET lastRunTime = ? WHERE id = ?"
	// 将时间戳转换为 time.Time 类型

	_, err := db.Exec(query, datetimeFormat, proxyID)
	if err != nil {
		color.Red("更新代理 %d 的 lastRunTime 失败: %v\n", proxyID, err)
	}
}
