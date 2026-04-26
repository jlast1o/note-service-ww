// @title           Notes Service API
// @version         1.0
// @description     Микросервис для управления заметками.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://github.com/your-username/notes-service
// @contact.email  your-email@example.com

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey APIKeyAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"service/internal/config"
	"service/internal/handler"
	"service/internal/middleware"
	"service/internal/store"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	_ "service/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @Summary      Создать заметку
// @Description  Создаёт новую заметку с заданным title и опциональным content.
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        note body      model.Note true "Данные заметки (обязательное поле title)"
// @Success      201  {object}  model.Note "Успешное создание"
// @Failure      400  {object}  map[string]string "Ошибка валидации"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка"
// @Router       /notes [post]
func _createAnnotation() {}

// @Summary      Получить все заметки
// @Description  Возвращает список заметок с пагинацией.
// @Tags         notes
// @Produce      json
// @Param        page query     int false "Номер страницы (по умолчанию 1)" default(1)
// @Param        size query     int false "Размер страницы (по умолчанию 20, максимум 100)" default(20)
// @Success      200  {object}  PaginatedResponse "Страница заметок"
// @Failure      400  {object}  map[string]string "Некорректные параметры"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка"
// @Router       /notes [get]
func _getAllAnnotation() {}

// @Summary      Получить заметку по ID
// @Description  Возвращает одну заметку по её идентификатору.
// @Tags         notes
// @Produce      json
// @Param        id   path      int true "ID заметки"
// @Success      200  {object}  model.Note "Найденная заметка"
// @Failure      400  {object}  map[string]string "Некорректный ID"
// @Failure      404  {object}  map[string]string "Заметка не найдена"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка"
// @Router       /notes/{id} [get]
func _getByIDAnnotation() {}

// @Summary      Обновить заметку
// @Description  Полностью заменяет заметку с указанным id.
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        id   path      int true "ID заметки"
// @Param        note body      model.Note true "Новые данные заметки (обязательное поле title)"
// @Success      200  {object}  model.Note "Обновлённая заметка"
// @Failure      400  {object}  map[string]string "Ошибка валидации"
// @Failure      404  {object}  map[string]string "Заметка не найдена"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка"
// @Router       /notes/{id} [put]
func _updateAnnotation() {}

// @Summary      Удалить заметку
// @Description  Удаляет заметку с указанным id.
// @Tags         notes
// @Param        id   path      int true "ID заметки"
// @Success      204  "Без тела, успешное удаление"
// @Failure      400  {object}  map[string]string "Некорректный ID"
// @Failure      404  {object}  map[string]string "Заметка не найдена"
// @Failure      500  {object}  map[string]string "Внутренняя ошибка"
// @Router       /notes/{id} [delete]
func _deleteAnnotation() {}

func initTracer(ctx context.Context) func() {
	exporter, err := otlptracehttp.New(
		ctx, otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)

	if err != nil {
		slog.Error("OTLP exporter не законнектился:", "error", err)
		return func() {}
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("service"),
		semconv.ServiceVersion("1.0.0"),
		attribute.String("enviroment", "development"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("Tracer выключается", "error", err)
		}
		exporter.Shutdown(ctx)
	}
}

func main() {
	// env load
	if err := godotenv.Load(); err != nil {
		log.Println("Файл конфигурации не обнаружен")
	}
	// cfg load
	cfg, err := config.Load()
	if err != nil {
		log.Printf("ошибка загрузки конфига %s", err)
		os.Exit(1)
	}

	middleware.InitMetrics()
	ctx := context.Background()

	shutdownTracer := initTracer(ctx)
	defer shutdownTracer()
	// db load
	pool, err := store.NewPool(ctx, cfg)
	if err != nil {
		log.Printf("Ошибка подключения к БД: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	// db migrate
	if err := store.Migrate(ctx, pool); err != nil {
		log.Printf("Ошибка миграции: %v", err)
		os.Exit(1)
	}

	// redis load

	redisClient, err := store.NewRedisClient(ctx, cfg)

	if err != nil {
		slog.Error("Редиска не законнектилась", "error", err)
		os.Exit(1)
	}

	defer redisClient.Close()

	noteStore := store.NewNoteStore(pool)
	cachedStore := store.NewCachedStore(noteStore, redisClient)
	noteHandler := handler.NewNoteHandler(cachedStore)

	healthHandler := handler.NewHealthHandler(pool, redisClient)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.MetricsMiddleware)

	r.Post("/notes", noteHandler.Create)
	r.Get("/notes", noteHandler.GetAll)
	r.Get("/notes/{id}", noteHandler.GetByID)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	r.Get("/health", healthHandler.Check)
	r.Delete("/notes/{id}", noteHandler.Delete)
	r.Put("/notes/{id}", noteHandler.Update)

	otelHandler := otelhttp.NewHandler(r, "notes-service-http")

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      otelHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen error", "error", err)
			os.Exit(1)
		}
	}()
	// gracefull shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGABRT)
	sig := <-quit
	slog.Info("shutting down", "sygnal", sig.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server down error", "error", err)
	}

	pool.Close()

	slog.Info("server gracefully stopped")
}
