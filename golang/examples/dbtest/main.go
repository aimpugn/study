package dbtest

import (
	"time"

	"github.com/go-sql-driver/mysql"
)

func GetDSN() {
	config := mysql.NewConfig()
	config.User = "username"
	config.Passwd = "password"
	config.Addr = "host" + ":" + "port"
	config.DBName = "dbname"
	config.Net = "tcp"
	config.ParseTime = true
	loc, _ := time.LoadLocation("Asia/Seoul")
	config.Loc = loc
}
