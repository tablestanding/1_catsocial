package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {
	mux := http.NewServeMux()

	srv := &http.Server{}
	srv.Addr = ":5781"
	srv.Handler = mux

	// speeches.InstallController(mux, speeches.NewService(speeches.NewSQLiteRepo(db), whisperx.New()))

	mux.Handle("/v1/", http.StripPrefix("/v1/", http.HandleFunc()))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		log.Printf("server has started listening on: %s\n", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http listen and serve: %v\n", err)
		}
	}()

	<-ctx.Done()
	err := srv.Shutdown(context.Background())
	if err != nil {
		log.Printf("shutdown server: %v\n", err)
	}
}
