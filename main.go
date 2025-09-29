package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/labring/aiproxy-free/config"
	"github.com/labring/aiproxy-free/db"
	"github.com/labring/aiproxy-free/middleware"
	"github.com/labring/aiproxy-free/server"
	"github.com/labring/aiproxy-free/utils"
	"github.com/labring/aiproxy-free/utils/pprof"
	log "github.com/sirupsen/logrus"
)

var (
	listen    string
	pprofPort int
)

func init() {
	flag.StringVar(&listen, "listen", "0.0.0.0:3000", "http server listen")
	flag.IntVar(&pprofPort, "pprof-port", 15000, "pport http server port")
}

func initializePprof() {
	go func() {
		err := pprof.RunPprofServer(pprofPort)
		if err != nil {
			log.Errorf("run pprof server error: %v", err)
		}
	}()
}

var loadedEnvFiles []string

func loadEnv() {
	envfiles := []string{
		".env",
		".env.local",
	}
	for _, envfile := range envfiles {
		absPath, err := filepath.Abs(envfile)
		if err != nil {
			panic(
				fmt.Sprintf(
					"failed to get absolute path of env file: %s, error: %s",
					envfile,
					err.Error(),
				),
			)
		}

		file, err := os.Stat(absPath)
		if err != nil {
			continue
		}

		if file.IsDir() {
			continue
		}

		if err := godotenv.Overload(absPath); err != nil {
			panic(fmt.Sprintf("failed to load env file: %s, error: %s", absPath, err.Error()))
		}

		loadedEnvFiles = append(loadedEnvFiles, absPath)
	}
}

func printLoadedEnvFiles() {
	for _, envfile := range loadedEnvFiles {
		log.Infof("loaded env file: %s", envfile)
	}
}

func setupHTTPServer() (*http.Server, *gin.Engine) {
	initializePprof()

	e := gin.New()

	e.Use(
		gin.RecoveryWithWriter(log.StandardLogger().Writer()),
		middleware.NewLog(log.StandardLogger()),
	)
	server.SetRouter(e)

	listenEnv := os.Getenv("LISTEN")
	if listenEnv != "" {
		listen = listenEnv
	}

	return &http.Server{
		Addr:              listen,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           e,
	}, e
}

func listenAndServe(srv *http.Server) {
	if err := srv.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal("failed to start HTTP server: " + err.Error())
	}
}

func main() {
	flag.Parse()

	loadEnv()

	config.ReloadEnv()

	utils.InitLog(log.StandardLogger(), config.DebugEnabled)

	printLoadedEnvFiles()

	err := db.InitDatabase(config.DSN)
	if err != nil {
		log.Fatalf("init database failed: %v", err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv, _ := setupHTTPServer()
	log.Infof("server started on http://%s", srv.Addr)
	go listenAndServe(srv)

	<-ctx.Done()

	shutdownSrvCtx, shutdownSrvCancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer shutdownSrvCancel()

	log.Info("shutting down http server...")
	if err := srv.Shutdown(shutdownSrvCtx); err != nil {
		log.Error("server forced to shutdown: " + err.Error())
		return
	}
	log.Info("server shutdown successfully")
}
