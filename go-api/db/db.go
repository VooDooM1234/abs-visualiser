package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	Conn *pgx.Conn
	Ctx  context.Context
}

type ABS_CPI struct {
	TIME_PERIOD string
	VALUE       float64
}

// start postgres server
func StartUp() {

}

func NewDatabase(ctx context.Context) (*Database, error) {
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	log.Print("Connected to database")

	return &Database{
		Conn: conn,
		Ctx:  ctx,
	}, nil
}

func (db *Database) Close() {
	db.Conn.Close(db.Ctx)
}

// boiler plate query for refrene
func (db *Database) boilerplater() error {
	rows, err := db.Conn.Query(db.Ctx, "SELECT * FROM users")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		// process rows
	}
	return nil
}

func (d *Database) UpsertDataABSCPI(data interface{}) error {
	log.Print("INSERTING ABS CPI data")

	cpiData, ok := data.(ABS_CPI)
	if !ok {
		return fmt.Errorf("invalid type: expected ABS_CPI, got %T", data)
	}

	_, err := d.Conn.Exec(d.Ctx,
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
