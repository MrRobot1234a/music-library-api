package handlers

import (
	"music-library/internal/db"
	"music-library/internal/models"
	"net/http"
	"strings"

	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// GetSongs возвращает список песен с фильтрацией и пагинацией
// @Summary Получить список песен
// @Description Возвращает список песен, поддерживает фильтрацию по группе, названию и дате релиза
// @Tags Songs
// @Accept json
// @Produce json
// @Param group query string false "Фильтр по группе"
// @Param song query string false "Фильтр по названию песни"
// @Param releaseDate query string false "Фильтр по дате релиза"
// @Param limit query int false "Лимит записей на страницу" default(10)
// @Param page query int false "Номер страницы" default(1)
// @Success 200 {array} models.Song
// @Failure 400 {object} map[string]string
// @Router /songs [get]
func GetSongs(c *gin.Context) {
	var songs []models.Song
	query := db.DB

	// Фильтрация по группе (если передан параметр ?group=)
	if group := c.Query("group"); group != "" {
		fmt.Println(group)
		query = query.Where("band ILIKE ?", "%"+group+"%")
	}

	// Фильтрация по названию песни (если передан параметр ?song=)
	if song := c.Query("song"); song != "" {
		query = query.Where("title ILIKE ?", "%"+song+"%")
	}

	// Фильтрация по дате релиза (если передан параметр ?releaseDate=)
	if releaseDate := c.Query("releaseDate"); releaseDate != "" {
		query = query.Where("release_date = ?", releaseDate)
	}

	// Пагинация (по умолчанию limit=10, page=1)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	// Выполняем запрос с фильтрацией и пагинацией
	query.Limit(limit).Offset(offset).Find(&songs)
	c.JSON(http.StatusOK, songs)
}

// GetLyrics возвращает текст песни по куплетам
// @Summary Получить текст песни
// @Description Возвращает определённый куплет песни (по номеру).
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Param verse query int false "Номер куплета" default(1)
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /songs/{id}/lyrics [get]
func GetLyrics(c *gin.Context) {
	var song models.Song

	// Получаем ID песни из URL
	id := c.Param("id")
	if err := db.DB.First(&song, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	// Разбиваем текст песни на куплеты
	verses := strings.Split(song.Text, "\n\n")

	// Получаем номер куплета из query (?verse=2)
	verseIndex, err := strconv.Atoi(c.DefaultQuery("verse", "1"))
	if err != nil || verseIndex < 1 || verseIndex > len(verses) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verse number"})
		return
	}

	// Отправляем конкретный куплет
	c.JSON(http.StatusOK, gin.H{"verse": verses[verseIndex-1]})
}


// UpdateSong обновляет данные о песне
// @Summary Обновить данные песни
// @Description Обновляет информацию о песне по её ID
// @Tags Songs
// @Accept json
// @Produce json
// @Param id path int true "ID песни"
// @Param song body models.Song true "Обновлённые данные песни"
// @Success 200 {object} models.Song
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /songs/{id} [put]
func UpdateSong(c *gin.Context) {
	id := c.Param("id")

	// Проверяем, есть ли такая песня
	var song models.Song
	if err := db.DB.First(&song, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	// Получаем новые данные
	var input models.Song
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем поля
	db.DB.Model(&song).Updates(input)
	c.JSON(http.StatusOK, song)
}

// DeleteSong удаляет песню по ID
// @Summary Удалить песню
// @Description Удаляет песню из базы по её ID
// @Tags Songs
// @Param id path int true "ID песни"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /songs/{id} [delete]
func DeleteSong(c *gin.Context) {
	id := c.Param("id")

	// Проверяем, есть ли такая песня
	var song models.Song
	if err := db.DB.First(&song, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
		return
	}

	// Удаляем песню
	db.DB.Delete(&song)
	c.JSON(http.StatusOK, gin.H{"message": "Song deleted"})
}

// AddSong добавляет новую песню и обогащает её данными из внешнего API
// @Summary Добавить песню (с обогащением)
// @Description Добавляет песню в базу данных, запрашивая дополнительные данные (текст, дата релиза, ссылка) из внешнего API
// @Tags Songs
// @Accept json
// @Produce json
// @Param song body models.Song true "Новая песня (group и song обязательны)"
// @Success 200 {object} models.Song
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /songs [post]
func AddSong(c *gin.Context) {
	var input struct {
		Band string `json:"group" binding:"required"`
		Title string `json:"song" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем дополнительные данные из API
	song, err := FetchSongDetails(input.Band, input.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch song details"})
		return
	}

	// Сохраняем в базу данных
	db.DB.Create(song)
	c.JSON(http.StatusOK, song)
}

func FetchSongDetails(group, title string) (*models.Song, error) {
	_ = godotenv.Load()

	apiBaseURL := os.Getenv("EXTERNAL_API_URL")
	if apiBaseURL == "" {
		return nil, fmt.Errorf("API base URL is not set")
	}

	url := fmt.Sprintf("%s/info?group=%s&song=%s", apiBaseURL, url.QueryEscape(group), url.QueryEscape(title))
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var songDetails struct {
		ReleaseDate string `json:"releaseDate"`
		Text string `json:"text"`
		Link string `json:"link"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&songDetails); err != nil {
		return nil, err
	}

	return &models.Song{
		Band: group,
		Title: title,
		ReleaseDate: songDetails.ReleaseDate,
		Text: songDetails.Text,
		Link: songDetails.Link,
	}, nil
}
