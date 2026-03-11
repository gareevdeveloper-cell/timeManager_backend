// @title           timeManager API
// @version         1.0
// @description     Backend API для работы с командой над проектом.
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"log"

	"github.com/joho/godotenv"
	_ "testik/docs"

	"testik/internal/app"
	"testik/internal/config"
)

func main() {
	// Загружаем .env в переменные окружения (игнорируем ошибку, если файла нет)
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatalf("app: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}
