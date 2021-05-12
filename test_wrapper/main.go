package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drone/ff-golang-server-sdk/test_wrapper/handlers"
	"github.com/drone/ff-golang-server-sdk/test_wrapper/restapi"
	"github.com/drone/ff-golang-server-sdk/test_wrapper/wrapperconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	config := wrapperconfig.Config{}

	viper.SetDefault("WrapperHostname", "")
	viper.SetDefault("WrapperPort", "4000")
	viper.SetDefault("BaseURL", "http://localhost/api/1.0")

	viper.SetDefault("SdkKey", "1bca46aa-0abe-4b22-a5f0-904422db288b")
	viper.SetDefault("EnableStreaming", true)
	viper.BindEnv("WrapperHostname", "WRAPPER_HOSTNAME")
	viper.BindEnv("WrapperPort", "WRAPPER_PORT")
	viper.BindEnv("BaseURL", "SDK_BASE_URL")
	viper.BindEnv("SdkKey", "SDK_KEY")
	viper.BindEnv("EnableStreaming", "ENABLE_STREAMING")

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to viper decode into struct, %v", err)
	}

	log.Infof("Starting up with config: \n%+v\n", config)
	setupHandlers(config)
}

func setupHandlers(config wrapperconfig.Config) {
	myServer := handlers.NewServer(config)
	e := echo.New()

	apiGroup := e.Group("/api/1.0")
	//apiGroup.Use(oapimiddleware.OapiRequestValidatorWithOptions(swagger, &oapimiddleware.Options{}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "Cache-Control"},
	}))

	restapi.RegisterHandlers(apiGroup, &myServer)

	go func() {
		serverAddress := fmt.Sprintf("%s:%s", config.WrapperHostname, config.WrapperPort)
		fmt.Printf("Starting api wrapper on: %s", serverAddress)
		err := e.Start(serverAddress)
		if err != nil && err != http.ErrServerClosed {
			log.Fatal("shutting down http server")
		}
	}()

	waitForShutdown(e)
}

func waitForShutdown(e *echo.Echo) {
	// Handle sigterm and shutdown gracefully
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	log.Debug("Waiting for Shutdown")

	<-termChan // Blocks here until interrupted

	// Handle shutdown
	log.Info("Shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
