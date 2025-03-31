package main

import (
	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
	"taker/api"
	db "taker/common"
)

func main() {
	db.InitDB()

	for {
		api.Work()
	}
}
