package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/MatthewTabatneck/stock-screener/internal/screener"
	"github.com/MatthewTabatneck/stock-screener/internal/store"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Summary api/main.go: This file will start a interrupt that is activated by ctrl+c, loads .env, opens csv of tickers whle parsing and cleaning data,
// opens database and pings to check if online, deletes any tickers in tickers table that has already been processed and removes from table,
// takes []string output from csv parse and inputs values into tickers table for experimentation/parsing.

func main() {
	//Starts an os.Interrupt to terminate the program at any point, important to add ctx to all functions needed termination
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	//loads .env file to take out env variables so they are not hardcoded when commiting changes
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	//opens ticker csv file and defers its closing until done with the file
	file, err := os.Open("tickers.csv")
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}
	defer file.Close()
	//LoadtickersCSV will take the file and parse its contents while converting to all caps and checking for duplicates
	tickers, err := screener.LoadtickersCSV(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	//takes database url and confirms exsistence
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	//sql.Open is accessing database and defering close till completion
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	//pings database url to confirm if online
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	//Deletes tickers from the tickers table that were already processed
	err = store.CleanupProcessedTickers(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	//InsertTickers will take the []string and input those values into the tickers table
	store.InsertTickers(ctx, db, tickers)

}
