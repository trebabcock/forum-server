package audit

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Auditor struct {
	DB *gorm.DB
}

type Audit struct {
	ID          string `gorm:"UNIQUE"`
	UserID      string
	DateTime    time.Time
	Action      string
	ErrorString string
	Message     string
}

func (a *Auditor) Init() {
	dsn := os.Getenv("DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Could not connect to database")
		os.Exit(1)
	}
	a.DB = db
}

func (a *Auditor) Log(userId, action, errorString, message string) {
	id, _ := uuid.NewUUID()
	auditLog := Audit{
		ID:          id.String(),
		UserID:      userId,
		DateTime:    time.Now(),
		Action:      action,
		ErrorString: errorString,
		Message:     message,
	}
	a.DB.Save(&auditLog)
	log.Println(message)
}
