package main

import (
	"jemref_go/internal/config"
	"jemref_go/internal/domain/id"
	"jemref_go/internal/infrastructure/mock"
	"jemref_go/internal/interface/handler"
	"jemref_go/internal/middleware"
	"jemref_go/internal/usecase"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初期化DI
	log.Print("initializing config")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("DB_HOST=%s, DB_Port=%s, DB_Name=%s, DB_User=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBUser,
	)

	log.Print("finish configuration")

	// DBはまだ使用しない
	//	db, err := infrastructure.NewDB(cfg)
	// if err != nil {
	// 	log.Fatal(err)

	log.Print("initializing routing")

	// ハンドラ呼び出し
	recordHandler := NewRecordHandler()
	userHandler := NewUserHandler()

	// ルーティング設定
	r := gin.Default()

	// 認証なしルート（health check）
	r.GET("/health", handler.Health)

	// 認証ありルート
	auth := r.Group("/api")
	auth.Use(middleware.Auth())

	recordHandler.RegisterRoutes(auth)
	userHandler.RegisterRoutes(auth)

	log.Print("finsh routing")

	log.Print("server starting")
	// 実行
	r.Run(":8080")
}

func NewRecordHandler() *handler.RecordHandler {
	// repositoryはmock
	repo := mock.NewRecordRepositoryMock()
	uc := usecase.NewRecordUsecase(repo)
	return handler.NewRecordHandler(uc)
}

func NewUserHandler() *handler.UserHandler {
	// repositoryはmock
	repo := mock.NewUserRepositoryMock()
	idGen := id.NewUlidGenerator()
	uc := usecase.NewUserUsecase(repo, idGen)
	return handler.NewUserHandler(uc)
}
