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

	_ "testik/docs"

	"github.com/joho/godotenv"

	"testik/internal/app"
	"testik/internal/config"
)

func main() {
	// Загружаем .env в переменные окружения (в Docker переменные приходят через env_file)
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env not found, using environment variables: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	a, err := app.New(cfg)
	log.Println("app: %v", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("app: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Fatalf("run: %v", err)
	}
}
