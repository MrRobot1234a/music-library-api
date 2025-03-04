package main

import (
	"music-library/internal/db"
	"music-library/internal/handlers"
	"music-library/internal/utils"

	"github.com/gin-gonic/gin"

	_ "music-library/docs" // Подключаем автоматически сгенерированную документацию
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

func main() {
	
	utils.InitLogger()
	utils.Logger.Info("Запуск сервера...")
	
	db.InitDB()
	utils.Logger.Info("База данных успешно подключена")
	
	r := gin.Default()
	r.GET("/songs", handlers.GetSongs)
	r.POST("/songs", handlers.AddSong)
	r.GET("/songs/:id/lyrics", handlers.GetLyrics)
	r.DELETE("/songs/:id", handlers.DeleteSong)
	r.PUT("/songs/:id", handlers.UpdateSong)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	utils.Logger.Info("Сервер слушает порт 8080")
	r.Run(":8080")
}
