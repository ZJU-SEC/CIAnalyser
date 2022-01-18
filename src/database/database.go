package database

import (
	"CIHunter/src/config"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	DBSection, err := config.Config.GetSection("DB")
	if err != nil {
		panic(err)
	}

	host := config.ParseKey(DBSection, "HOST")
	user := config.ParseKey(DBSection, "USER")
	password := config.ParseKey(DBSection, "PASSWORD")
	dbname := config.ParseKey(DBSection, "DBNAME")
	port := config.ParseKey(DBSection, "PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", host, user, password, dbname, port)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic(err)
	}

	err = DB.AutoMigrate(Repo{})
	if err != nil {
		panic(err)
	}
}
