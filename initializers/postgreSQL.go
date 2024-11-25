package initializers

import (
	"database/sql"
	"fmt"
	"songLibrary/models"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

var (
	DB *gorm.DB
)

func Migrate(config DBConfig) {
	var err error

	log.Info("Начинаем миграцию") // Info-лог

	if err = createDatabaseIfNotExists(config); err != nil {
		log.Fatal("Не удалось создать БД: " + err.Error())
	}

	log.Info("Открываем соединение") // Info-лог

	log.WithField("DB config.DSN", config.DSN).Debug("DSN для подключения") // Debug-лог

	DB, err = gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к БД: " + err.Error())
	}

	log.Info("Успешное подключение к БД") // Info-лог

	log.Info("Мигрирую модели через GORM") // Info-лог

	err = DB.AutoMigrate(&models.Group{}, &models.Song{}, &models.Lyrics{})
	if err != nil {
		log.Fatalf("Ошибка миграции: %s", err)
		return
	}

	log.Info("Миграция успешна") // Info-лог
}

func createDatabaseIfNotExists(config DBConfig) error {
	log.Info("Подключаемся к БД к дефолтной БД postgres") // Info-лог

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%d sslmode=disable", config.Host, config.User, config.Pswd, config.Port)

	log.WithField("DB dsn", dsn).Debug("DSN для подключения к дефолтной БД postgres") // Debug-лог

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %s", err)
		return err
	}
	defer db.Close()

	log.Info("Проверяем существует ли необходимая БД с именем ", config.Name) // Info-лог

	log.WithField("DB config.Name", config.Name).Debug("Имя необходимой БД") // Debug-лог

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", config.Name).Scan(&exists)
	if err != nil {
		log.Fatalf("Не удалось проверить существование БД: %s", err)
		return err
	}

	if !exists {

		log.Info("Нужная БД не найдена, создаём ", config.Name) // Info-лог

		_, err = db.Exec("CREATE DATABASE " + config.Name)
		if err != nil {
			log.Fatalf("Не удалось создать БД %s: %s", config.Name, err)
			return err
		}

		log.Info("БД успешно создана") // Info-лог
	}

	return nil
}
