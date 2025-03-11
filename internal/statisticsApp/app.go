package statisticsApp

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/SeeingTheFire/tsg.statistics/internal/database"
	"github.com/SeeingTheFire/tsg.statistics/internal/database/migration"
	"github.com/SeeingTheFire/tsg.statistics/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	initConfig()

	logger := initLogger()

	dbPool, err := initDatabase()
	if err != nil {
		logger.Fatalf("%s: %v", "Error on connect to database", err)
	}

	redisClient := initRedis()

	initService(dbPool, redisClient, logger)

	initHandler(logger)
}

func initConfig() {
	viper.SetConfigType("toml")

	viper.AddConfigPath(".")
	viper.SetConfigName(".config")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file: ", viper.ConfigFileUsed())
	} else {
		log.Fatal(err)
	}
}

func initLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}

	return logger
}

func initDatabase() (*sql.DB, error) {
	dbConn := database.DatabaseConnector{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetInt("database.port"),
		Username: viper.GetString("database.username"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.name"),
		SSLMode:  viper.GetString("database.sslmode"),
	}

	dbPool, err := dbConn.Connect()
	if err != nil {
		return nil, err
	}

	// Migrate database schema
	err = migration.Up(dbPool)
	if err != nil {
		log.Fatalf("Error on migrate schema : %v", err)
	}

	return dbPool, nil
}

func initRedis() *redis.Client {
	client := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf(`%s:%d`, viper.GetString("redis.host"), viper.GetInt("redis.port")),
			Password: viper.GetString("redis.password"),
		},
	)

	return client
}

func initService(dbPool *sql.DB, redisClient *redis.Client, logger *logrus.Logger) *services.Parser {
	parser := services.NewParser(nil)
	return parser
}

func initHandler(logger *logrus.Logger) {
	ctx := context.Background()

	r := gin.Default()

	http.Handle("/", r)

	srv := &http.Server{
		Addr:         fmt.Sprintf(`:%d`, viper.GetInt("parserApp.port")),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
