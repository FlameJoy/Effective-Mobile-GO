package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"songLibrary/initializers"
	"songLibrary/models"
	"songLibrary/utils"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Handler struct{}

func NewHandler() *Handler {
	h := Handler{}
	return &h
}

// @Summary      Добавить песню
// @Description  **Добавить песню**
// @Tags         Song
// @Accept       json
// @Produce      json
// @Param        Request body  models.Input  true  "Информация о песне"
// @Success      200  {object}  utils.RespOK "Успешный ответ"
// @Failure      400  {object}  utils.ProblemDetails "Ошибка валидации"
// @Failure      409  {object}  utils.ProblemDetails "Песня уже существует"
// @Failure      500  {object}  utils.ProblemDetails "Internal Server Error"
// @Router       /api/v1/library/songs/add [post]
func (h *Handler) AddSong(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "AddSong")

	var input models.Input

	if err := c.Bind(&input); err != nil {
		return c.JSON(utils.HttpResErrorRFC9457("error", err, http.StatusBadRequest, log, c))
	}

	log.WithField("group", input.Group).Debug("Имя группы") // Debug-лог
	log.WithField("song", input.Song).Debug("Имя песни")    // Debug-лог

	log.Info("Валидация входных данных") // Info-лог

	if len(input.Group) < 1 || len(input.Group) > 60 {
		return c.JSON(utils.HttpResErrorRFC9457("error", fmt.Errorf("значение group пустое или слишком длинное: %s", input.Group), http.StatusBadRequest, log, c))
	}
	if len(input.Song) < 1 || len(input.Song) > 100 {
		return c.JSON(utils.HttpResErrorRFC9457("error", fmt.Errorf("значение song пустое или слишком длинное: %s", input.Song), http.StatusBadRequest, log, c))
	}

	log.Info("Проверка существования группы") // Info-лог

	var group models.Group
	if err := initializers.DB.Where("name = ?", input.Group).First(&group).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			log.Info("Группа не найдена, создаём новую") // Info-лог

			group = models.Group{Name: input.Group}
			if err := initializers.DB.Create(&group).Error; err != nil {
				return c.JSON(utils.HttpResErrorRFC9457("DB error", fmt.Errorf("не удалось создать группу: %s", err), http.StatusInternalServerError, log, c))
			}
		} else {
			return c.JSON(utils.HttpResErrorRFC9457("DB error", fmt.Errorf("не удалось найти группу: %s", input.Group), http.StatusInternalServerError, log, c))
		}
	}

	log.Info("Проверка существования песни") // Info-лог

	var song models.Song
	if err := initializers.DB.Where("title = ? AND group_id = ?", input.Song, group.ID).First(&song).Error; err != nil {
		if err == gorm.ErrRecordNotFound {

			log.Info("Песня не найдена, делаем запрос к внешнему API") // Info-лог

			apiURL := fmt.Sprintf("%s/info?group=%s&song=%s", os.Getenv("EXTERNAL_API_ADDR"), input.Group, input.Song)
			resp, err := http.Get(apiURL)
			if err != nil {

				log.Info("Не удалось получить положительный ответ от внешнего API") // Info-лог

				return c.JSON(utils.HttpResErrorRFC9457("error", fmt.Errorf("не удалось получить положительный ответ от внешнего API: %s", err), http.StatusInternalServerError, log, c))
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {

				log.WithField("status_code", resp.StatusCode).Error("Внешний API вернул ошибку")

				return c.JSON(utils.HttpResErrorRFC9457("error", fmt.Errorf("внешний API вернул ошибку: %d", resp.StatusCode), http.StatusBadRequest, log, c))
			}

			log.Info("Парсинг ответа от внешнего API") // Info-лог

			var songDetail struct {
				ReleaseDate string `json:"releaseDate"`
				Text        string `json:"text"`
				Link        string `json:"link"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
				log.Info("Не удалось распарсить ответ от внешнего API. Используем дефолтные значения") // Info-лог

				return c.JSON(utils.HttpResErrorRFC9457("error", fmt.Errorf("не удалось распарсить ответ от внешнего API: %s", err), http.StatusBadRequest, log, c))
			}

			// Эмуляция внешнего API
			// songDetail.ReleaseDate = "16.07.2006"
			// songDetail.Text = "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight"
			// songDetail.Link = "https://www.youtube.com/watch?v=Xsp3_a-PMTw"

			log.Info("Создание новой записи песни") // Info-лог

			song = models.Song{
				GroupID:     group.ID,
				Title:       input.Song,
				ReleaseDate: songDetail.ReleaseDate,
				Link:        songDetail.Link,
			}

			log.WithField("song.GroupID", song.GroupID).Debug("ID группы")                        // Debug-лог
			log.WithField("song.Title", song.Title).Debug("Имя песни")                            // Debug-лог
			log.WithField("song.ReleaseDate", songDetail.ReleaseDate).Debug("Дата выпуска песни") // Debug-лог
			log.WithField("song.Link", songDetail.Link).Debug("Ссылка на песню")                  // Debug-лог
			log.WithField("songDetail.Text", songDetail.Text).Debug("Текст песни")                // Debug-лог

			log.Info("Разбиение текста песни на куплеты") // Info-лог

			verses := utils.SplitIntoVerses(songDetail.Text)

			log.Info("Начало транзакции") // Info-лог

			err = initializers.DB.Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&song).Error; err != nil {
					return err
				}

				log.Info("Сохранение куплетов") // Info-лог

				for i, verse := range verses {
					lyrics := models.Lyrics{
						SongID: song.ID,
						Verse:  verse,
						Order:  i + 1,
					}

					log.WithField("SongID", lyrics.SongID).Debug("ID песни, для которой добаялются куплеты") // Debug-лог
					log.WithField("Verse", lyrics.Verse).Debug("Куплет песни")                               // Debug-лог
					log.WithField("Order", lyrics.Order).Debug("Порядок куплета")                            // Debug-лог

					if err := tx.Create(&lyrics).Error; err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				return c.JSON(utils.HttpResErrorRFC9457("Tx error", err, http.StatusInternalServerError, log, c))
			}

			log.Info("Завершение транзакции") // Info-лог

			return c.JSON(http.StatusOK, echo.Map{"message": "Песня добавлена"})
		} else {
			return c.JSON(utils.HttpResErrorRFC9457("DB error", err, http.StatusInternalServerError, log, c))
		}
	}

	return c.JSON(utils.HttpResErrorRFC9457("error", errors.New("песня уже существует"), http.StatusConflict, log, c))
}

// @Summary      Получение текста песни
// @Description  **Получение текста песни**
// @Tags         Song
// @Produce      json
// @Param        page query string true "Страница"
// @Param        limit query string true "Ограничение вывода"
// @Success      200  {object}  []models.Lyrics "Успешный ответ"
// @Failure      400  {object}  utils.ProblemDetails "Ошибка валидации"
// @Failure      500  {object}  utils.ProblemDetails "Internal Server Error"
// @Router       /api/v1/library/songs/:id/lyrics [get]
func (h *Handler) GetLyrics(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "GetLyrics")

	id := c.Param("id")
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	log.WithField("song.id", id).Debug("ID песни") // Debug-лог
	log.WithField("page", page).Debug("страница")  // Debug-лог
	log.WithField("limit", limit).Debug("лимит")   // Debug-лог

	log.Info("Проверяем, существует ли песня с данным ID") // Info-лог

	var song models.Song
	if err := initializers.DB.First(&song, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(utils.HttpResErrorRFC9457("DB error", errors.New("песня не найдена"), http.StatusNotFound, log, c))
		}
		return c.JSON(utils.HttpResErrorRFC9457("DB error", err, http.StatusInternalServerError, log, c))
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		if len(page) == 0 {
			return c.JSON(utils.HttpResErrorRFC9457("error", errors.New("page пустое, укажите значение параметра"), http.StatusBadRequest, log, c))
		}
		return c.JSON(utils.HttpResErrorRFC9457("error", fmt.Errorf("не получается преобразовать значение страницы: %s", page), http.StatusBadRequest, log, c))
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		if len(limit) == 0 {
			return c.JSON(utils.HttpResErrorRFC9457("error", errors.New("limit пустое, укажите значение параметра"), http.StatusBadRequest, log, c))
		}
		return c.JSON(utils.HttpResErrorRFC9457("error", fmt.Errorf("не получается преобразовать значение лимита: %s", limit), http.StatusBadRequest, log, c))
	}
	if limitInt == 0 {

		log.Info("limitInt == 0 устанавливаем дефолтный размер страницы") // Info-лог

		limitInt = 5
	}

	var lyrics []models.Lyrics
	result := initializers.DB.Where("song_id = ?", id).Order("\"order\"").Offset((pageInt - 1) * limitInt).Limit(limitInt).Find(&lyrics)
	if result.Error != nil {
		return c.JSON(utils.HttpResErrorRFC9457("DB error", result.Error, http.StatusNotFound, log, c))
	}

	return c.JSON(http.StatusOK, lyrics)
}

// @Summary      Удаление песни
// @Description  **Удаление песни**
// @Tags         Song
// @Produce      json
// @Success      200  {object}  utils.RespOK "Успешный ответ"
// @Failure      404  {object}  utils.ProblemDetails "Песня не найдена"
// @Failure      500  {object}  utils.ProblemDetails "Internal Server Error"
// @Router       /api/v1/library/songs/delete/:id [delete]
func (h *Handler) DeleteSong(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "DeleteSong")

	id := c.Param("id")

	log.WithField("song.id", id).Debug("ID песни") // Debug-лог

	log.Info("Проверяем, существует ли песня с данным ID") // Info-лог

	var song models.Song
	if err := initializers.DB.First(&song, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(utils.HttpResErrorRFC9457("DB error", errors.New("песня не найдена"), http.StatusNotFound, log, c))
		}
		return c.JSON(utils.HttpResErrorRFC9457("DB error", err, http.StatusInternalServerError, log, c))
	}

	log.Info("Начинаем транзакцию") // Info-лог

	err := initializers.DB.Transaction(func(tx *gorm.DB) error {

		log.Info("Удаление текстов песни") // Info-лог

		if err := tx.Where("song_id = ?", id).Delete(&models.Lyrics{}).Error; err != nil {
			log.WithError(err).Error("error: не удалось удалить текст")
			return err
		}

		log.Info("Удаление самой песни") // Info-лог

		if err := tx.Delete(&models.Song{}, id).Error; err != nil {
			log.WithError(err).Error("error: не удалось удалить песню")
			return err
		}

		return nil
	})
	if err != nil {
		return c.JSON(utils.HttpResErrorRFC9457("Tx error", err, http.StatusInternalServerError, log, c))
	}

	log.Info("Завершение транзакции") // Info-лог

	return c.JSON(http.StatusOK, echo.Map{"message": "Песня и слова песни удалены"})
}

// @Summary      Редактирование песни
// @Description  **Редактирование данных песни, текста**
// @Tags         Song
// @Accept       json
// @Produce      json
// @Param        Request body  models.Edit  true  "Новая информация о песне"
// @Success      200  {object}  models.Song "Успешный ответ"
// @Failure      400  {object}  utils.ProblemDetails "Ошибка валидации"
// @Failure      404  {object}  utils.ProblemDetails "песня не найдена"
// @Failure      500  {object}  utils.ProblemDetails "Internal Server Error"
// @Router       /api/v1/library/songs/edit/:id [put]
func (h *Handler) EditSong(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "EditSong")

	id := c.Param("id")

	log.WithField("song.id", id).Debug("ID песни") // Debug-лог

	log.Info("Проверяем, существует ли песня с данным ID") // Info-лог

	var song models.Song
	if err := initializers.DB.First(&song, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(utils.HttpResErrorRFC9457("DB error", errors.New("песня не найдена"), http.StatusNotFound, log, c))
		}
		return c.JSON(utils.HttpResErrorRFC9457("DB error", err, http.StatusInternalServerError, log, c))
	}

	log.Info("Парсинг данных из запроса") // Info-лог

	var input models.Edit

	if err := c.Bind(&input); err != nil {
		return c.JSON(utils.HttpResErrorRFC9457("error", errors.New("неверные данные"), http.StatusBadRequest, log, c))
	}

	log.Info("Открытие транзакции") // Info-лог

	tx := initializers.DB.Begin()

	log.Info("Обновление информации о песне") // Info-лог

	song.Title = input.Title
	song.ReleaseDate = input.ReleaseDate
	song.Link = input.Link

	log.WithField("song.Title", song.Title).Debug("Название песни")                 // Debug-лог
	log.WithField("song.ReleaseDate", song.ReleaseDate).Debug("Дата выпуска песни") // Debug-лог
	log.WithField("song.Link", song.Link).Debug("Ссылка на песню")                  // Debug-лог

	log.Info("Проверка существования группы") // Info-лог

	var group models.Group
	if err := tx.Where("name = ?", input.GroupName).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			log.Info("Группа не найдена, создаём новую") // Info-лог

			group = models.Group{Name: input.GroupName}
			if err := tx.Create(&group).Error; err != nil {
				tx.Rollback()
				return c.JSON(utils.HttpResErrorRFC9457("Tx error", err, http.StatusInternalServerError, log, c))
			}
		} else {
			tx.Rollback()
			return c.JSON(utils.HttpResErrorRFC9457("Tx error", err, http.StatusInternalServerError, log, c))
		}
	}
	song.GroupID = group.ID

	log.WithField("song.GroupID", song.GroupID).Debug("ID группы") // Debug-лог

	log.Info("Обновление песни в БД") // Info-лог

	if err := tx.Save(&song).Error; err != nil {
		tx.Rollback()
		return c.JSON(utils.HttpResErrorRFC9457("Tx error", err, http.StatusInternalServerError, log, c))
	}

	log.Info("Обновление лирики") // Info-лог

	for _, lyric := range input.Lyrics {
		var existingLyric models.Lyrics
		if err := tx.First(&existingLyric, lyric.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				return c.JSON(utils.HttpResErrorRFC9457("Tx error", err, http.StatusInternalServerError, log, c))
			}
			tx.Rollback()
			return c.JSON(utils.HttpResErrorRFC9457("Tx error", err, http.StatusInternalServerError, log, c))
		}

		existingLyric.Verse = lyric.Verse
		existingLyric.Order = lyric.Order

		log.WithField("existingLyric.Order", existingLyric.Order).Debug("Порядок нового абзаца") // Debug-лог
		log.WithField("existingLyric.Verse", existingLyric.Verse).Debug("Новый абзац текста")    // Debug-лог

		if err := tx.Save(&existingLyric).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
	}

	log.Info("Завершение транзакции") // Info-лог

	tx.Commit()

	return c.JSON(http.StatusOK, song)
}

// @Summary      Получения списка песен
// @Description  **Получения списка песен**
// @Tags         Song
// @Produce      json
// @Param        groupName query string false "Название группы"
// @Param        songTitle query string false "Название песни"
// @Param        releaseDate query string false "Дата выпуска песни"
// @Param        link query string false "Ссылка на песню"
// @Param        lyrics query string false "Фрагмент текста песни"
// @Param        page query string true "Страница"
// @Param        limit query string true "Ограничение вывода"
// @Success      200  {object}  models.SongsList "Успешный ответ"
// @Failure      400  {object}  utils.ProblemDetails "Ошибка валидации"
// @Failure      500  {object}  utils.ProblemDetails "Internal Server Error"
// @Router       /api/v1/library/songs [get]
func (h *Handler) GetSongsList(c echo.Context) error {
	ctx := c.Request().Context()
	log := log.WithContext(ctx).WithField("prefix", "GetSongsList")

	log.Info("Получаем параметры фильтрации и пагинации из запроса") // Info-лог

	groupName := c.QueryParam("group_name")
	songTitle := c.QueryParam("song_title")
	releaseDate := c.QueryParam("release_date")
	link := c.QueryParam("link")
	lyrics := c.QueryParam("lyrics")
	limit := c.QueryParam("limit")
	page := c.QueryParam("page")

	log.Info("Преобразуем пагинацию в числа") // Info-лог

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	log.WithField("pageInt", pageInt).Debug("Номер страницы") // Debug-лог

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 10
	}
	offset := (pageInt - 1) * limitInt

	log.WithField("offset", offset).Debug("Смещение") // Debug-лог

	log.Info("Проверка формата даты") // Info-лог

	if releaseDate != "" {
		_, err := time.Parse("02.01.2006", releaseDate)
		if err != nil {
			return c.JSON(utils.HttpResErrorRFC9457("error", fmt.Errorf("release_date должен быть в формате dd.MM.yyyy: %v", err), http.StatusBadRequest, log, c))
		}
	}

	var songs []models.Song
	var totalCount int64

	query := initializers.DB.Preload("Lyrics")

	log.Info("Применяем фильтры") // Info-лог

	if groupName != "" && len(groupName) < 255 {
		query = query.Joins("JOIN groups ON groups.id = songs.group_id").Where("LOWER(groups.name) LIKE LOWER(?)", "%"+groupName+"%")
	}
	if songTitle != "" && len(songTitle) < 255 {
		query = query.Where("LOWER(songs.title) LIKE LOWER(?)", "%"+songTitle+"%")
	}
	if releaseDate != "" && len(songTitle) == 10 {
		query = query.Where("songs.release_date = ?", releaseDate)
	}
	if link != "" {
		query = query.Where("LOWER(songs.link) LIKE LOWER(?)", "%"+link+"%")
	}
	if lyrics != "" {
		query = query.Joins("JOIN lyrics ON lyrics.song_id = songs.id").Where("LOWER(lyrics.verse) LIKE LOWER(?)", "%"+lyrics+"%")
	}

	log.WithField("query", query).Debug("итоговый запрос") // Debug-лог

	log.Info("Получаем количество записей для пагинации") // Info-лог

	if err := query.Model(&models.Song{}).Count(&totalCount).Error; err != nil {
		return c.JSON(utils.HttpResErrorRFC9457("DB error", errors.New("не удалось посчитать кол-во песен"), http.StatusInternalServerError, log, c))
	}

	log.WithField("totalCount", totalCount).Debug("количество записей для пагинации") // Debug-лог

	log.Info("Получаем данные с пагинацией") // Info-лог

	if err := query.Offset(offset).Limit(limitInt).Find(&songs).Error; err != nil {
		return c.JSON(utils.HttpResErrorRFC9457("DB error", errors.New("не удалось получить песни"), http.StatusInternalServerError, log, c))
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data":        songs,
		"total_count": totalCount,
		"page":        pageInt,
		"limit":       limitInt,
	})
}
