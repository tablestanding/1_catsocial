package main

import (
	"catsocial/cat"
	"catsocial/match"
	"catsocial/pkg/env"
	"catsocial/pkg/pgxtrx"
	"catsocial/user"
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func runServer() {
	// === CONTEXT
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// === OTEL
	otelShutdown, err := setupOTelSDK(ctx)
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// === DB
	dbPool := initDB(ctx)
	defer dbPool.Close()

	pgxTrx := pgxtrx.New(dbPool)

	// === ENV VAR
	port := ":" + cmp.Or(os.Getenv("PORT"), "8080")

	saltCountString := env.MustLoad("BCRYPT_SALT")
	saltCount, err := strconv.Atoi(saltCountString)
	if err != nil {
		log.Fatalf("parsing BCRYPT_SALT as int: %s\n", err.Error())
	}

	jwtSecret := env.MustLoad("JWT_SECRET")

	// === HTTP MUX
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, h http.Handler) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, h)
		mux.Handle(pattern, handler)
	}

	srv := &http.Server{
		Addr:         port,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
	}

	var h http.Handler
	h = mux
	h = http.TimeoutHandler(h, 60*time.Second, "timeout")
	h = otelhttp.NewHandler(h, "/")
	srv.Handler = h

	// === USER
	userSQL := user.NewSQL(dbPool)
	userSvc := user.NewService(userSQL, saltCount, jwtSecret)
	userCtrl := user.NewController(userSvc)

	handleFunc("POST /v1/user/register", http.HandlerFunc(userCtrl.RegisterHandler))
	handleFunc("POST /v1/user/login", http.HandlerFunc(userCtrl.LoginHandler))

	// === CAT
	catSQL := cat.NewSQL(pgxTrx)
	catSvc := cat.NewService(catSQL, pgxTrx)
	catCtrl := cat.NewController(catSvc)

	createCatHandler := userCtrl.AuthMiddleware(http.HandlerFunc(catCtrl.CreateHandler))
	handleFunc("POST /v1/cat", createCatHandler)
	searchCatHandler := userCtrl.AuthMiddleware(http.HandlerFunc(catCtrl.SearchHandler))
	handleFunc("GET /v1/cat", searchCatHandler)
	updateCatHandler := userCtrl.AuthMiddleware(http.HandlerFunc(catCtrl.UpdateHandler))
	handleFunc("PUT /v1/cat/{id}", updateCatHandler)
	deleteCatHandler := userCtrl.AuthMiddleware(http.HandlerFunc(catCtrl.DeleteHandler))
	handleFunc("DELETE /v1/cat/{id}", deleteCatHandler)

	// === MATCH
	matchSQL := match.NewSQL(pgxTrx)
	matchSvc := match.NewService(matchSQL, catSvc, catSQL, pgxTrx)
	matchCtrl := match.NewController(matchSvc)

	createMatchHandler := userCtrl.AuthMiddleware(http.HandlerFunc(matchCtrl.CreateHandler))
	handleFunc("POST /v1/cat/match", createMatchHandler)
	getMatchHandler := userCtrl.AuthMiddleware(http.HandlerFunc(matchCtrl.GetHandler))
	handleFunc("GET /v1/cat/match", getMatchHandler)
	approveMatchHandler := userCtrl.AuthMiddleware(http.HandlerFunc(matchCtrl.ApproveHandler))
	handleFunc("POST /v1/cat/match/approve", approveMatchHandler)
	rejectMatchHandler := userCtrl.AuthMiddleware(http.HandlerFunc(matchCtrl.RejectHandler))
	handleFunc("POST /v1/cat/match/reject", rejectMatchHandler)
	deleteMatchHandler := userCtrl.AuthMiddleware(http.HandlerFunc(matchCtrl.DeleteHandler))
	handleFunc("DELETE /v1/cat/match/{id}", deleteMatchHandler)

	// === SERVE HTTP AND GRACE SHUTDOWN
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

func initDB(ctx context.Context) *pgxpool.Pool {
	dbName := env.MustLoad("DB_NAME")
	dbPort := env.MustLoad("DB_PORT")
	dbHost := env.MustLoad("DB_HOST")
	dbUsername := env.MustLoad("DB_USERNAME")
	dbPassword := env.MustLoad("DB_PASSWORD")
	dbParams := env.MustLoad("DB_PARAMS")
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?%s", dbUsername, dbPassword, dbHost, dbPort, dbName, dbParams)

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		log.Fatalf("Unable to parse postgres conn string: %s\n", err.Error())
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer()
	cfg.MaxConns = 70

	dbpool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("Unable to create db pool: %s\n", err.Error())
	}

	return dbpool
}
