package utils

import (
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

var (
	location, _ = time.LoadLocation("Europe/Moscow")
)

type RespOK struct {
	Message string `json:"message" example:"Запрос успешно выполнен"`
}

// RFC 9457
type ProblemDetails struct {
	StatusCode int    `json:"status_code,omitempty" example:"400"`
	Method     string `json:"method" example:"POST"`
	Time       string `json:"time" example:"2024-11-24 19:33:57"`
	Type       string `json:"type" example:"/api/v1/..."`
	Title      string `json:"title" example:"DB error"`
	Detail     string `json:"detail" example:"описание ошибки"`
}

func HttpResErrorRFC9457(title string, err error, statusCode int, log *log.Entry, c echo.Context) (int, ProblemDetails) {
	if log != nil {
		log.Error(err)
	}
	return statusCode, ProblemDetails{
		StatusCode: statusCode,
		Time:       time.Now().In(location).Format("2006-01-02 15:04:05"),
		Method:     c.Request().Method,
		Type:       c.Request().URL.String(),
		Title:      title,
		Detail:     err.Error(),
	}
}

// Хелпер для разбиения текста песни на куплеты
func SplitIntoVerses(text string) []string {
	return strings.Split(text, "\n\n") // Разделяем по двойным переносам строк
}
