package api

import (
	"errors"
	"strings"
	"taker/common"
	"time"
)

// 签到流程
func check(taker Taker) (int, error) {
	// 获取message
	message, err := generateNonce(taker)
	if err != nil {
		return 0, err
	}

	// sign message
	privateKeyHex := taker.PrivateKey
	if strings.Contains(taker.PrivateKey, "0x") {
		privateKeyHex = taker.PrivateKey[2:]
	}
	signMessage, err := common.SignMessage(message, privateKeyHex)
	if err != nil {
		return 0, err
	}

	// 登录
	token, err := login(taker, message, signMessage)
	if err != nil {
		return 0, err
	}

	// 每日任务获取失败
	err = list(taker, token)
	if err != nil {
		return 0, err
	}

	miningTime := 0
	// 数据库为空，就获取比较
	if !taker.LastRunTime.Valid {
		miningTime, err = totalMiningTime(taker, token)
		if err != nil {
			return 0, err
		}
		if miningTime == 0 {
			return 0, errors.New("请先绑定推特")
		}

		// 如果不为0 就比较
		secondsIn24Hours := 24 * 60 * 60
		// 将时间戳加上 24 小时
		newTimestamp := miningTime + secondsIn24Hours
		t := time.Unix(int64(newTimestamp), 0).Local()
		if t.Before(time.Now()) { // 应该签到
			err = startMining(taker, token)
			if err != nil {
				return 0, err
			}

			// 获取下次签到时间
			miningTime, err = totalMiningTime(taker, token)
			if err != nil {
				return 0, err
			}
		}
		return miningTime, nil
	}

	// 如果比当前时间早
	if taker.LastRunTime.Time.Before(time.Now()) {
		// 获取下次签到时间
		miningTime, err = totalMiningTime(taker, token)
		if err != nil {
			return 0, err
		}
		// 如果不为0 就比较
		secondsIn24Hours := 24 * 60 * 60
		// 将时间戳加上 24 小时
		newTimestamp := miningTime + secondsIn24Hours
		t := time.Unix(int64(newTimestamp), 0).Local()
		if t.Before(time.Now()) { // 应该签到
			err = startMining(taker, token)
			if err != nil {
				return 0, err
			}

			// 获取下次签到时间
			miningTime, err = totalMiningTime(taker, token)
			if err != nil {
				return 0, err
			}
		} else {
			return newTimestamp, nil
		}
	}

	return miningTime, nil
}
