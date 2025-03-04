package db

import (
	"log"
	"music-library/internal/models"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	err = DB.AutoMigrate(&models.Song{})
	if err != nil {
		log.Fatal("Ошибка миграции БД:", err)
	}

	log.Println("База данных успешно инициализирована")
}
