package initializers

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	Port      int
	Domain    string
	DebugMode bool
}

type DBConfig struct {
	Port int
	User string
	Pswd string
	Name string
	Host string
	DSN  string
}

func FormServerConfig() ServerConfig {
	var config ServerConfig

	log.Info("Начинаем формировать серверный конфиг") // Info-лог

	port, err := strconv.ParseInt(os.Getenv("PORT"), 10, 64)
	if err != nil {
		fmt.Printf("error: ошибка парсинга %s использую дефолтное значение\n", os.Getenv("PORT"))
		config.Port = 8080
	}
	config.Port = int(port)

	log.WithField("Server config.Port", config.Port).Debug("Установлено значение порта сервера") // Debug-лог

	config.Domain = os.Getenv("DOMAIN_NAME")
	if len(config.Domain) == 0 {
		fmt.Printf("error: ошибка парсинга %s использую дефолтное значение\n", os.Getenv("DOMAIN_NAME"))
		config.Domain = "localhost"
	}

	log.WithField("Server config.Domain", config.Domain).Debug("Установлено значение домена сервера") // Debug-лог

	return config
}

func FormDBConfig() DBConfig {
	var config DBConfig

	log.Info("Начинаем формировать БД конфиг") // Info-лог

	port, err := strconv.ParseInt(os.Getenv("DB_PORT"), 10, 64)
	if err != nil {
		fmt.Printf("error: ошибка парсинга %s использую дефолтное значение\n", os.Getenv("DB_PORT"))
		config.Port = 5432
	}
	config.Port = int(port)

	log.WithField("Server config.Port", config.Port).Debug("Установлено значение порта БД") // Debug-лог

	config.User = os.Getenv("DB_USER")
	if len(config.User) == 0 {
		panic("error: DB_USER не указан или пуст")
	}

	log.WithField("DB config.User", config.User).Debug("Установлено значение пользователя БД") // Debug-лог

	config.Pswd = os.Getenv("DB_PSWD")
	if config.Pswd == "" {
		panic("error: DB_PSWD не указан или пуст")
	}

	log.WithField("DB config.Pswd", config.Pswd).Debug("Установлено значение пароля БД") // Debug-лог

	config.Name = os.Getenv("DB_NAME")
	if config.Name == "" {
		fmt.Printf("error:  DB_NAME не указан или пуст %s, использую дефолтное значение 'postgres'\n", os.Getenv("DB_NAME"))
		config.Name = "postgres"
	}

	log.WithField("DB config.Name", config.Name).Debug("Установлено значение имени БД") // Debug-лог

	config.Host = os.Getenv("DB_HOST")
	if config.Host == "" {
		fmt.Printf("error: DB_HOST не указан или пуст %s, использую дефолтное значение 'localhost'\n", os.Getenv("DB_HOST"))
		config.Host = "localhost"
	}

	log.WithField("DB config.Host", config.Host).Debug("Установлено значение хоста БД") // Debug-лог

	config.DSN = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", config.Host, config.User, config.Pswd, config.Name, config.Port)

	log.WithField("DB config.DSN", config.DSN).Debug("Сформирована DSN строка для подключения к БД") // Debug-лог

	return config
}
