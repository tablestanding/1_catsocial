package main

import (
	"catsocial/pkg/env"
	"catsocial/pkg/jwt"
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

	port := ":" + cmp.Or(os.Getenv("PORT"), "5781")

	saltCountString := env.MustLoad("BCRYPT_SALT")
	saltCount, err := strconv.Atoi(saltCountString)
	if err != nil {
		log.Fatalf("parsing BCRYPT_SALT as int: %s\n", err.Error())
	}

	jwtSecret := env.MustLoad("JWT_SECRET")

	mux := http.NewServeMux()

	srv := &http.Server{}
	srv.Addr = port
	srv.Handler = mux

	userSQL := user.NewSQL(dbPool)
	userSvc := user.NewService(userSQL, saltCount, jwtSecret)
	userCtrl := user.NewController(userSvc)

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTQ1MTM5MjR9.m6tHDomfsAmy3dbWmpSA4zXpTSUQCpoDdPpC3efvL6w"
	log.Println("is token valid")
	log.Println(jwt.IsTokenValid(token, "secret"))

	mux.Handle("POST /v1/user/register", http.HandlerFunc(userCtrl.RegisterHandler))

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
