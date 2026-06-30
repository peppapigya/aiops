package configs

import (
	"fmt"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	GORMDB *gorm.DB
)

func NewDB() *gorm.DB {
	if Config == nil {
		log.Printf("database init failed: config not loaded")
		return nil
	}

	databaseConfig := Config.Database.MySQL
	var err error
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		databaseConfig.Username,
		databaseConfig.Password,
		databaseConfig.Host,
		databaseConfig.Port,
		databaseConfig.Database,
	)

	GORMDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel()),
	})
	if err != nil {
		log.Printf("database init failed: %v", err)
		GORMDB = nil
		return nil
	}

	sqlDB, _ := GORMDB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return GORMDB
}

func CloseDB() {
	if GORMDB == nil {
		return
	}
	sqlDB, _ := GORMDB.DB()
	if err := sqlDB.Close(); err != nil {
		log.Printf("database close failed: %v", err)
		return
	}
}

func gormLogLevel() logger.LogLevel {
	if Config == nil {
		return logger.Warn
	}

	switch strings.ToLower(Config.Server.LogLevel) {
	case "debug":
		return logger.Info
	case "error", "silent":
		return logger.Silent
	default:
		return logger.Warn
	}
}
