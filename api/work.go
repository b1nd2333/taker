package api

import (
	"database/sql"
	"fmt"
	"github.com/fatih/color"
	"math/rand"
	"os"
	db2 "taker/common"
	"time"
)

type Taker struct {
	ID          int
	Address     string
	PrivateKey  string
	IP          string
	Port        string
	Username    string
	Password    string
	LastRunTime sql.NullTime // 使用 sql.NullTime 来处理空值
}

func Work() {
	color.Blue("开始执行新的任务，当前时间%s", time.Now().Local())
	// 打开 SQLite 数据库
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		color.Red("无法连接到数据库:%s", err)
		os.Exit(0)
	}
	defer db.Close()

	// 查询所有代理
	takers, err := fetchProxies(db)
	if err != nil {
		color.Red("获取账号失败:%s", err)
		os.Exit(0)
	}

	// 遍历代理并依次登录
	for _, taker := range takers {
		//fmt.Printf("检查代理 %s:%s...\n", taker.IP, taker.Port)

		// 如果 lastRunTime 为空或小于当前时间，则开始登录
		if shouldCheck(taker.LastRunTime) {
			// 登录
			miningTime, err := check(taker)
			if err != nil {
				color.Red("第%d行,地址%s使用代理 %s:%s 登录失败: %v", taker.ID, taker.Address, taker.IP, taker.Port, err)
			} else {
				// 更新 lastRunTime
				secondsIn16Hours := 24 * 60 * 60
				// 将时间戳加上 24 小时
				newTimestamp := miningTime + secondsIn16Hours
				t := time.Unix(int64(newTimestamp), 0).Local()

				// 将 time.Time 转换为 DATETIME 格式（YYYY-MM-DD HH:MM:SS）
				datetimeFormat := t.Format("2006-01-02 15:04:05")
				db2.UpdateLastRunTime(db, taker.ID, datetimeFormat)

				secondsIn24Hours := 32 * 60 * 60
				// 将时间戳加上 24 小时
				newTimestamp1 := miningTime + secondsIn24Hours
				t1 := time.Unix(int64(newTimestamp1), 0).Local()

				// 将 time.Time 转换为 DATETIME 格式（YYYY-MM-DD HH:MM:SS）
				datetimeFormat1 := t1.Format("2006-01-02 15:04:05")
				color.Green("第%d行，地址%s签到成功，下次签到时间为%v\n", taker.ID, taker.Address, datetimeFormat1)
			}
		}
		//} else {
		//	("地址%s暂不需要签到，下次签到时间为%v\n", taker.Address, taker.LastRunTime.Time)
		//}
	}
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 生成一个 10 到 60 之间的随机分钟数
	randomMinutes := rand.Intn(51) + 10 // 51是上限，10是下
	randomSeconds := rand.Intn(60)

	fmt.Printf("程序将休眠 %d 分 %d 秒...\n", randomMinutes, randomSeconds)
	time.Sleep(time.Duration(randomMinutes)*time.Minute + time.Duration(randomSeconds)*time.Second)
}

// 判断是否需要签到
func shouldCheck(lastRunTime sql.NullTime) bool {
	// 如果 lastRunTime 为空，表示没有记录，或者 lastRunTime 小于当前时间
	//fmt.Println(lastRunTime.Time, time.Now(), lastRunTime.Time.Before(time.Now()))
	//fmt.Println(lastRunTime.Time, time.Now().Local(), lastRunTime.Time.Before(time.Now().Local()))
	if !lastRunTime.Valid || lastRunTime.Time.Before(time.Now().Local()) {
		return true
	}
	return false
}

// 从数据库获取代理
func fetchProxies(db *sql.DB) ([]Taker, error) {
	query := "SELECT id, address,privateKey, ip, port, username, password, lastRunTime FROM taker"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []Taker
	for rows.Next() {
		var taker Taker
		var lastRunTime sql.NullTime
		err = rows.Scan(&taker.ID, &taker.Address, &taker.PrivateKey, &taker.IP, &taker.Port, &taker.Username, &taker.Password, &lastRunTime)
		if err != nil {
			return nil, err
		}
		taker.LastRunTime = lastRunTime // 保存 LastRunTime
		proxies = append(proxies, taker)
	}
	return proxies, nil
}
