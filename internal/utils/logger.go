package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Глобальный логгер
var Logger = logrus.New()

func InitLogger() {
	// Настраиваем формат вывода
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Указываем уровень логирования
	Logger.SetLevel(logrus.DebugLevel)


	Logger.SetOutput(os.Stdout)
}
