package parserApp

import (
	"fmt"
	"github.com/SeeingTheFire/tsg.statistics/internal/services"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Run() {
	initConfig()
	ctx := context.Background()
	logger := initLogger()

	parser := initService(logger, &ctx)

	initParser(logger, parser)
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

func initService(logger *logrus.Logger, c *context.Context) *services.Parser {
	parser := services.NewParser(logger, c)
	return parser
}

func initParser(logger *logrus.Logger, parser *services.Parser) {

	err, rows := parser.ParseRows()
	if err != nil {
		return
	}

	for _, row := range rows.Rows {
		err, replayInfo := parser.ParseReplay(row.Name)
		if err != nil {
			return
		}

		err, d := parser.ParseReplayInfo(replayInfo)
		if err != nil {
			return
		}

		println(d)
	}
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server exiting")
}
