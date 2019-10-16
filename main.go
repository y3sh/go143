package main

import (
	"flag"
	"os"

	"github.com/y3sh/go143/http"
	"github.com/y3sh/go143/twitter"

	log "github.com/sirupsen/logrus"
)

var (
	serverPort  = flag.String("port", getEnv("PORT", "3000"), "Rest API Port, e.g. 3000")
	logLevelStr = flag.String("logLevel", getEnv("LOG_LEVEL", "debug"),
		"Log level: trace, debug, info, warn, error, fatal, panic")
)

func main() {
	flag.Parse()
	setupLogger(logLevelStr)

	tweetService := twitter.NewTweetService()

	apiRouter := http.NewAPIRouter(tweetService)

	restAPIServer, err := http.NewServer(apiRouter, http.Port(*serverPort))
	if err != nil {
		log.Fatalf("Failed to create api server. \n%+v\n", err)
	}

	log.Info("143 API starting . . .")

	err = restAPIServer.Start()
	if err != nil {
		log.Infof("Server shutdown.  \n%+v\n", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func setupLogger(logLevelStr *string) {
	logLevel, err := log.ParseLevel(*logLevelStr)
	if err != nil {
		logLevel = log.TraceLevel
		log.Warn("Log level invalid or not provided, using trace level.")
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
		ForceColors:   true,
	})

	// Log filename and line number
	log.SetReportCaller(true)

	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel)
}