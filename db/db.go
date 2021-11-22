package db

import (
	"forum-server/app/model"
	"forum-server/audit"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(auditor *audit.Auditor) *gorm.DB {
	dsn := os.Getenv("DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Could not connect to database")
		auditor.Log("", "Connect To Database", "Error", err.Error())
		os.Exit(1)
	}
	auditor.Log("", "Connect To Database", "Success", "")
	mdb := migrate(db)
	auditor.Log("", "Migrate Database", "Success", "")
	return mdb
}

func migrate(db *gorm.DB) *gorm.DB {
	db.AutoMigrate(&model.User{}, &model.Post{}, &model.Comment{}, &model.Board{}, &audit.Audit{})
	return db
}
