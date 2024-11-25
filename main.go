package main

import (
	"flag"
	"fmt"
	"songLibrary/handlers"
	"songLibrary/initializers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

var debugMode = flag.Bool("debug", false, "Дебаг-режим")

// @title           songLibraryAPI
// @version         1.0
// @description     Song library API by Ilya Valentuikevich

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Дебаг режим
	flag.Parse()
	if *debugMode {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug-режим активирован")
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Env файл
	initializers.LoadEnv()

	// Инициализация config + DB
	serverConfig := initializers.FormServerConfig()
	dbConfig := initializers.FormDBConfig()

	// Автомиграция
	initializers.Migrate(dbConfig)

	// Echo
	e := echo.New()

	log.Info("Регистрируем middleware") // Info-лог

	e.Use(middleware.Recover(), middleware.Logger())

	h := handlers.NewHandler()

	log.Info("Регистрируем handlers") // Info-лог

	registerHandlers(e, h)

	log.Info("Запускаем сервер") // Info-лог

	log.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%v", serverConfig.Port)))
}
