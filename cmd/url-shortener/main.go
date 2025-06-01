package main

import (
	"log/slog"
	"net/http"
	"os"

	"main/internal/config"
	"main/internal/http-server/handlers/url/save"
	urlDelete "main/internal/http-server/handlers/url/delete"
	"main/internal/lib/logger/handlers/slogpretty"
	"main/internal/lib/logger/sl"
	"main/internal/storage/sqlite"
	"main/internal/http-server/handlers/redirect"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main() {
	// init config: cleanenv
	cfg := config.MustLoad()

	// init logger: slog
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))

	log.Info(
		"starting...", 
		slog.String("env" , cfg.Env),
		slog.String("version", "1"),
	)
	log.Debug("debug messages are enabled")
	// log.Error("error messages are enabled")
	
	// init storage: sqlite
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

	// init router: chi "chi render"
	router := chi.NewRouter()

	// middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger) 
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		// TODO: add DELETE /url/{id}
		r.Delete("/{alias}", urlDelete.New(log, storage))
	})
	router.Get("/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	server := &http.Server {
		Addr: cfg.Address,
		Handler: router,
		ReadTimeout: cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

	// TODO: run server
	
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
			)
	}
	return log
}


func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}


/*
запрос: curl -u admin:admin -X POST https://7ebfe213-b10c-45bb-814a-a5bac47a4c8a-00-2i33r3pc2loc0.kirk.replit.dev:8080/url/ -H "Content-Type: application/json" -d '{"url":"https://github.com/wolfychanOwO", "alias":"OwO"}'
запрос: curl -u admin:admin -X POST http://localhost:8080/url/ -H "Content-Type: application/json" -d '{"url":"https://github.com/wolfychanOwO", "alias":"OwO"}'

OwO, nicie: github
lala3: google

statistic:
695 lines - не считая тестов
*/
