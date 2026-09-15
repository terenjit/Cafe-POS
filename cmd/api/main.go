package main

import (
	"fmt"
	"os"

	"github.com/terenjit/Cafe-POS/config"
	"github.com/terenjit/Cafe-POS/internal/handler"
	"github.com/terenjit/Cafe-POS/pkg/database"
	pkgredis "github.com/terenjit/Cafe-POS/pkg/redis"
	"github.com/terenjit/Cafe-POS/pkg/validator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}

	db, err := database.NewMySQL(cfg.MysqlDSN())
	if err != nil {
		fmt.Println("mysql error:", err)
		os.Exit(1)
	}
	defer db.Close()
	fmt.Println("MySQL connected.")

	rdb, err := pkgredis.NewRedis(cfg.RedisAddr(), cfg.Redis.Password)
	if err != nil {
		fmt.Println("redis error:", err)
		os.Exit(1)
	}
	defer rdb.Close()
	fmt.Println("Redis connected.")

	v := validator.New()
	r := handler.NewRouter(db, cfg, v)

	fmt.Println("Starting server on port :" + cfg.App.Port)
	if err := r.Run(":" + cfg.App.Port); err != nil {
		fmt.Println("server error:", err)
		os.Exit(1)
	}
}
