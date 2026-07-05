// @title JemRef Beta API
// @version 1.0
// @license.name Capra
// @description JemRef Beta backend API
// @description This API requires a valid Firebase ID Token.
// @description Requests without authentication will be rejected.
// @host localhost:8080
// @BasePath /api/v0
package main

import (
	"context"
	"database/sql"
	"jemref_go/internal/config"
	"jemref_go/internal/domain/id"
	"jemref_go/internal/handler"
	"jemref_go/internal/infrastructure"
	infraDB "jemref_go/internal/infrastructure/db"
	"jemref_go/internal/infrastructure/mock"
	"jemref_go/internal/middleware"
	"jemref_go/internal/usecase"
	"log"

	_ "jemref_go/docs"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {

	log.Print("initializing config")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// DB初期化
	log.Print("initializing database")
	log.Printf("DB_HOST=%s, DB_Port=%s, DB_Name=%s, DB_User=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBUser,
	)

	db, err := infrastructure.NewDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// firebase初期化
	log.Print("initializing routing")
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error initializing firebase client: %v", err)
	}

	// ルーティング
	log.Print("initializing routing")
	healthHandler := handler.NewHealthHandler(db)
	generalHandler := newGeneralHandler(db)
	recordHandler := newRecordHandler()
	userHandler := newUserHandler(db, client)

	// 認証なしルート ------------------------------------------------------
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/api/v0/health", healthHandler.Health)
	r.GET("/api/v0/policies/:"+handler.ParamPolicyType, generalHandler.GetPolicies)

	// 認証ありルート ------------------------------------------------------

	authUC := newAuthUsecase(db, client)
	// 会員登録用
	join := r.Group("/api/v0")
	join.Use(
		middleware.FirebaseAuth(app),
		middleware.ChkUnregistered(authUC),
	)
	join.POST("/join", userHandler.CreateUser)

	// 共通認証
	auth := r.Group("/api/v0")
	auth.Use(
		middleware.FirebaseAuth(app),
		middleware.FindCurrentUser(authUC),
	)

	// ルーティング登録
	recordHandler.RegisterRoutes(auth)
	userHandler.RegisterRoutes(auth)
	log.Print("finsh routing")

	// 実行
	log.Print("server starting")
	r.Run("0.0.0.0:8080")
}

func newRecordHandler() *handler.RecordHandler {
	// repositoryはmock
	repo := mock.NewRecordRepositoryMock()
	uc := usecase.NewRecordUsecase(repo)
	return handler.NewRecordHandler(uc)
}

func newUserHandler(db *sql.DB, client *auth.Client) *handler.UserHandler {
	userRepo := infraDB.NewUserRepositoryImpl(db)
	termsRepo := infraDB.NewGeneralRepositoryImpl(db)
	firebaseRepo := infraDB.NewFirebaseRepositoryImpl(client)
	idGen := id.NewUlidGenerator()
	uc := usecase.NewUserUsecase(userRepo, termsRepo, firebaseRepo, idGen)
	return handler.NewUserHandler(uc)
}

func newGeneralHandler(db *sql.DB) *handler.GeneralHandler {
	repo := infraDB.NewGeneralRepositoryImpl(db)
	uc := usecase.NewGeneralUsecaseImpl(repo)
	return handler.NewGeneralHandler(uc)
}

func newAuthUsecase(db *sql.DB, c *auth.Client) *usecase.AuthUsecaseImpl {
	userRepo := infraDB.NewUserRepositoryImpl(db)
	FirebaseRepo := infraDB.NewFirebaseRepositoryImpl(c)
	return usecase.NewAuthUsecaseImpl(userRepo, FirebaseRepo)
}
