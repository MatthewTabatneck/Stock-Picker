# Stock Picker

Stock Screener ingests a list of tickers from `tickers.csv`, loads them into Postgres, and runs workers that fetch data for further analysis. The repository will also serve as a sandbox for experiment projects that test different auto-trading strategies using the tickers table as the source of which stocks to investigate.

## Prerequisites

- Go 1.21+
- PostgreSQL database
- `goose` CLI for running migrations
- An Alpha Vantage API key

## Configuration

Create a `.env` file (or set environment variables directly) with the following values:

- `DATABASE_URL`: Postgres connection string (e.g., `postgres://user:pass@localhost:5432/stocks?sslmode=disable`)
- `ALPHA_KEY`: Alpha Vantage API key
- `MIN_INTERVAL_SEC` (optional): Minimum seconds between API calls (default: `13`)
- `PROCESSORS` (optional): Number of concurrent processors (default: `3`)

## Database setup

Run migrations to create the tables used by the API and worker processes:

```sh
make migrate-up
```

Use `make migrate-down` if you need to roll back the schema.

## Running locally

The tickers importer reads from `tickers.csv` and stores the symbols in the `tickers` table for experimentation.

- Start the API ticker loader:
  ```sh
  make run-api
  ```
- Start the worker to process tickers and fetch data from Alpha Vantage:
  ```sh
  make run-worker
  ```

Both commands respect the environment variables in your `.env` file.

## Docker

You can build and run the services with Docker:

```sh
make docker-build
make docker-run-api
make docker-run-worker
```

Or bring up everything (API, worker, and dependencies) via Docker Compose:

```sh
make up
```

Use `make down` to stop the containers and `make logs` to tail output.

## Future experiments

The `tickers` table acts as the central queue of symbols to research. Future experimental projects will plug into this table to trial different automated trading strategies while reusing the same list of candidate stocks.

