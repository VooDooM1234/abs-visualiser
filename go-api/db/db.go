package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
	Ctx  context.Context
}

type ABS_CPI struct {
	TIME_PERIOD string
	VALUE       float64
}

type ABSDataflow struct {
	ID                  string
	Version             string
	AgencyID            string
	IsExternalReference bool
	IsFinal             bool
	Name                string
}

type ABSDataflowList struct {
	Dataflows []ABSDataflow
}

// Create a new database pool
func NewDatabase(ctx context.Context) (*Database, error) {
	dbURL := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	log.Print("Connected to database using pool")

	return &Database{
		Pool: pool,
		Ctx:  ctx,
	}, nil
}

// Close the pool
func (db *Database) Close() {
	db.Pool.Close()
}

// Upsert CPI data
func (d *Database) UpsertDataABSCPI(data interface{}) error {
	log.Print("INSERTING ABS CPI data")

	cpiData, ok := data.(ABS_CPI)
	if !ok {
		return fmt.Errorf("invalid type: expected ABS_CPI, got %T", data)
	}

	_, err := d.Pool.Exec(d.Ctx,
		`INSERT INTO "ABS_CPI" (TIME_PERIOD, VALUE)
         VALUES ($1, $2)
         ON CONFLICT (TIME_PERIOD)
         DO UPDATE SET VALUE = EXCLUDED.VALUE`,
		cpiData.TIME_PERIOD, cpiData.VALUE,
	)
	if err != nil {
		return fmt.Errorf("upsert failed: %w", err)
	}

	return nil
}

// Get ABS dataflows
func (d *Database) GetABSDataflow(query string) ([]ABSDataflow, error) {
	log.Print("Fetching ABS dataflow list with query: ", query)

	var absDataflows []ABSDataflow
	rows, err := d.Pool.Query(d.Ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dataflow ABSDataflow
		if err := rows.Scan(
			&dataflow.ID,
			&dataflow.Version,
			&dataflow.AgencyID,
			&dataflow.IsExternalReference,
			&dataflow.IsFinal,
			&dataflow.Name,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		absDataflows = append(absDataflows, dataflow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return absDataflows, nil
}
