package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/viper"

	"github.com/Sinbad-HQ/kyc/config"
)

func startHTTPServer() {
	app, err := NewApp()
	if err != nil {
		log.Fatalln(err)
	}

	apiAddress := fmt.Sprintf("0.0.0.0:%s", viper.GetString("port"))
	server := &http.Server{
		Addr:    apiAddress,
		Handler: app.NewHandler(),
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for a process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		shutdownCtx, _ := context.WithTimeout(serverCtx, 10*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				slog.Warn("graceful shutdown timed out.. forcing exit.")
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Warn(err.Error())
		}
		app.Shutdown()
		serverStopCtx()
	}()

	log.Printf("Serving on: %s\n", apiAddress)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	<-serverCtx.Done()
}

func Run() error {
	if err := config.ReadConfiguration(); err != nil {
		return err
	}
	startHTTPServer()
	return nil
}
