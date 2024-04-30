package main

import (
	"catsocial/cat"
	"catsocial/pkg/env"
	"catsocial/user"
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
)

func runServer() {
	dbPool := initDB()
	defer dbPool.Close()

	// === ENV VAR

	port := ":" + cmp.Or(os.Getenv("PORT"), "5781")

	saltCountString := env.MustLoad("BCRYPT_SALT")
	saltCount, err := strconv.Atoi(saltCountString)
	if err != nil {
		log.Fatalf("parsing BCRYPT_SALT as int: %s\n", err.Error())
	}

	jwtSecret := env.MustLoad("JWT_SECRET")

	// === HTTP MUX

	mux := http.NewServeMux()

	srv := &http.Server{}
	srv.Addr = port
	srv.Handler = mux

	// === USER

	userSQL := user.NewSQL(dbPool)
	userSvc := user.NewService(userSQL, saltCount, jwtSecret)
	userCtrl := user.NewController(userSvc)

	mux.Handle("POST /v1/user/register", http.HandlerFunc(userCtrl.RegisterHandler))
	mux.Handle("POST /v1/user/login", http.HandlerFunc(userCtrl.LoginHandler))

	// === CAT

	catSQL := cat.NewSQL(dbPool)
	catSvc := cat.NewService(catSQL)
	catCtrl := cat.NewController(catSvc)

	createCatHandler := userCtrl.AuthMiddleware(http.HandlerFunc(catCtrl.CreateHandler))
	mux.Handle("POST /v1/cat", createCatHandler)

	// === SERVE HTTP AND GRACE SHUTDOWN

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
	err = srv.Shutdown(context.Background())
	if err != nil {
		log.Printf("shutdown server: %v\n", err)
	}
}

func initDB() *pgxpool.Pool {
	dbName := env.MustLoad("DB_NAME")
	dbPort := env.MustLoad("DB_PORT")
	dbHost := env.MustLoad("DB_HOST")
	dbUsername := env.MustLoad("DB_USERNAME")
	dbPassword := env.MustLoad("DB_PASSWORD")
	dbParams := env.MustLoad("DB_PARAMS")
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?%s", dbUsername, dbPassword, dbHost, dbPort, dbName, dbParams)

	dbpool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to create db pool: %s\n", err.Error())
	}
	return dbpool
}
