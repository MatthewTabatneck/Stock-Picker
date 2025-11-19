package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MatthewTabatneck/stock-screener/internal/config"
	"github.com/MatthewTabatneck/stock-screener/internal/provider"
	"github.com/MatthewTabatneck/stock-screener/internal/store"
	_ "github.com/lib/pq"
)

func main() {
	//Starts os.Interrupt to terminate program whenever user wants
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}
func run(ctx context.Context) error {
	//Loads configuration parameters for goroutines and api calls
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	store.SetDB(db)
	//Ends the worker program in a set amount of downtime
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return err
	}

	// Keep polling until program finds unprocessed tickers
	var symbols []string
	for {
		symbols, err = store.GetAllTickers(ctx, db)
		if err != nil {
			return err
		}

		if len(symbols) == 0 {
			log.Println("No unprocessed tickers found. Waiting...")
			time.Sleep(2 * time.Second)
			continue
		}

		break
	}

	log.Printf("loaded %d tickers\n", len(symbols))
	log.Println(symbols)

	err = provider.GetAlpha(ctx, cfg, symbols)
	if err != nil {
		return err
	}

	return nil
}
