package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	customMiddleware "github.com/olezhek28/microservices-course-examples/week_1/http_chi_ogen/internal/middleware"
	weatherV1 "github.com/olezhek28/microservices-course-examples/week_1/http_chi_ogen/pkg/openapi/weather/v1"
)

const (
	httpPort = "8080"
	// Таймауты для HTTP-сервера
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// WeatherStorage представляет потокобезопасное хранилище данных о погоде
type WeatherStorage struct {
	mu       sync.RWMutex
	weathers map[string]*weatherV1.Weather
}

// NewWeatherStorage создает новое хранилище данных о погоде
func NewWeatherStorage() *WeatherStorage {
	return &WeatherStorage{
		weathers: make(map[string]*weatherV1.Weather),
	}
}

// GetWeather возвращает информацию о погоде по имени города
func (s *WeatherStorage) GetWeather(city string) *weatherV1.Weather {
	s.mu.RLock()
	defer s.mu.RUnlock()

	weather, ok := s.weathers[city]
	if !ok {
		return nil
	}

	return weather
}

// UpdateWeather обновляет данные о погоде для указанного города
func (s *WeatherStorage) UpdateWeather(city string, weather *weatherV1.Weather) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.weathers[city] = weather
}

// WeatherHandler реализует интерфейс weatherV1.Handler для обработки запросов к API погоды
type WeatherHandler struct {
	storage *WeatherStorage
}

// NewWeatherHandler создает новый обработчик запросов к API погоды
func NewWeatherHandler(storage *WeatherStorage) *WeatherHandler {
	return &WeatherHandler{
		storage: storage,
	}
}

// GetWeatherByCity обрабатывает запрос на получение данных о погоде по названию города
func (h *WeatherHandler) GetWeatherByCity(_ context.Context, params weatherV1.GetWeatherByCityParams) (weatherV1.GetWeatherByCityRes, error) {
	weather := h.storage.GetWeather(params.City)
	if weather == nil {
		return &weatherV1.NotFoundError{
			Code:    404,
			Message: "Weather for city '" + params.City + "' not found",
		}, nil
	}

	return weather, nil
}

// UpdateWeatherByCity обрабатывает запрос на обновление данных о погоде по названию города
func (h *WeatherHandler) UpdateWeatherByCity(_ context.Context, req *weatherV1.UpdateWeatherRequest, params weatherV1.UpdateWeatherByCityParams) (weatherV1.UpdateWeatherByCityRes, error) {
	// Создаем объект погоды с полученными данными
	weather := &weatherV1.Weather{
		City:        params.City,
		Temperature: req.Temperature,
		UpdatedAt:   time.Now(),
	}

	// Обновляем данные в хранилище
	h.storage.UpdateWeather(params.City, weather)

	return weather, nil
}

// NewError создает новую ошибку в формате GenericError
func (h *WeatherHandler) NewError(_ context.Context, err error) *weatherV1.GenericErrorStatusCode {
	return &weatherV1.GenericErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: weatherV1.GenericError{
			Code:    weatherV1.NewOptInt(http.StatusInternalServerError),
			Message: weatherV1.NewOptString(err.Error()),
		},
	}
}

func main() {
	// Создаем хранилище для данных о погоде
	storage := NewWeatherStorage()

	// Создаем обработчик API погоды
	weatherHandler := NewWeatherHandler(storage)

	// Создаем OpenAPI сервер
	weatherServer, err := weatherV1.NewServer(weatherHandler)
	if err != nil {
		log.Fatalf("ошибка создания сервера OpenAPI: %v", err)
	}

	// Инициализируем роутер Chi
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(customMiddleware.RequestLogger)

	// Монтируем обработчики OpenAPI
	r.Mount("/", weatherServer)

	// Запускаем HTTP-сервер
	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атак - тип DDoS-атаки, при которой
		// атакующий умышленно медленно отправляет HTTP-заголовки, удерживая соединения открытыми и истощая
		// пул доступных соединений на сервере. ReadHeaderTimeout принудительно закрывает соединение,
		// если клиент не успел отправить все заголовки за отведенное время.
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("🚀 HTTP-сервер запущен на порту %s\n", httpPort)
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("❌ Ошибка запуска сервера: %v\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Завершение работы сервера...")

	// Создаем контекст с таймаутом для остановки сервера
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("❌ Ошибка при остановке сервера: %v\n", err)
	}

	log.Println("✅ Сервер остановлен")
}
