package main

import (
	"net/http"

	_ "externalAPI/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type SongDetail struct {
	ReleaseDate string `json:"releaseDate" example:"16.07.2006"`
	Text        string `json:"text" example:"Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight"`
	Link        string `json:"link" example:"https://www.youtube.com/watch?v=Xsp3_a-PMTw"`
}

// @title Music info API
// @version 0.0.1
// @description API для получения информации о песнях.
// @BasePath /
func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Роуты
	e.GET("/info", GetSongInfo)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Logger.Fatal(e.Start(":8081"))
}

// GetSongInfo возвращает информацию о песне по группе и названию.
// @Summary Получить информацию о песне
// @Description Получить данные о песне на основе группы и названия песни.
// @Tags songs
// @Accept json
// @Produce json
// @Param group query string true "Название группы"
// @Param song query string true "Название песни"
// @Success 200 {object} SongDetail
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /info [get]
func GetSongInfo(c echo.Context) error {
	group := c.QueryParam("group")
	song := c.QueryParam("song")

	// Проверка входных параметров
	if group == "" || song == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "требуются параметры группы и песни"})
	}

	// Затычка
	songDetail := SongDetail{
		ReleaseDate: "16.07.2006",
		Text:        "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
		Link:        "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
	}

	return c.JSON(http.StatusOK, songDetail)
}
