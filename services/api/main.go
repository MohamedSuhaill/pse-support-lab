// Command api is the Invoicely Lite backend.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"invoicely/api/internal/directory"
	"invoicely/api/internal/httpapi"
	"invoicely/api/internal/importer"
	"invoicely/api/internal/store"
	"invoicely/api/internal/submit"
	"invoicely/api/internal/validate"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbURL := env("DATABASE_URL", "postgres://invoicely:invoicely@localhost:5432/invoicely?sslmode=disable")
	aspURL := env("ASP_URL", "http://localhost:8090")
	addr := ":" + env("PORT", "8080")

	log.Info("api starting", "addr", addr, "asp_url", aspURL)

	st, err := store.Open(ctx, dbURL)
	if err != nil {
		log.Error("database config", "error", err)
		os.Exit(1)
	}
	for attempt := 1; ; attempt++ {
		if err = st.Ping(ctx); err == nil {
			break
		}
		if attempt == 30 {
			log.Error("database unavailable", "error", err)
			os.Exit(1)
		}
		log.Warn("waiting for database", "attempt", attempt, "error", err)
		time.Sleep(time.Second)
	}

	submitter := submit.New(st, submit.NewASPClient(aspURL), log, 4)
	submitter.Start(ctx)

	sources := importer.NewSourceCache()
	validator := validate.New(directory.New(aspURL), st, log)
	imp := importer.New(st, validator, submitter, sources, log)

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.New(st, imp, sources, submitter, log).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()

	log.Info("api listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
