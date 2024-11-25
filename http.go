package main

import (
	"songLibrary/handlers"

	_ "songLibrary/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func registerHandlers(e *echo.Echo, h *handlers.Handler) {
	api := e.Group("/api/v1")
	api.Use(middleware.CORS())

	// Swagger doc
	api.GET("/doc/*", echoSwagger.EchoWrapHandler())

	// Library
	library := api.Group("/library")

	library.GET("/songs", h.GetSongsList, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET},
	}))

	library.POST("/songs/add", h.AddSong, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.POST},
	}))

	library.GET("/songs/:id/lyrics", h.GetLyrics, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET},
	}))

	library.PUT("/songs/edit/:id", h.EditSong, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.PUT},
	}))

	library.DELETE("/songs/delete/:id", h.DeleteSong, middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.DELETE},
	}))
}
