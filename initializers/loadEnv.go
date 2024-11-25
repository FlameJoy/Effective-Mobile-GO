package initializers

import (
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func LoadEnv() {

	log.Info("Загружаем env-файл") // Info-лог

	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	log.Info("env-файл загруженл") // Info-лог
}
