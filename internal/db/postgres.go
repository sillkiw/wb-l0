package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func Connect(url string) (*DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DB{conn: db}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

// Сохраняем заказ (для примера: просто ключ+значение)
func (d *DB) SaveMessage(key, value []byte) error {
	_, err := d.conn.Exec(
		`INSERT INTO orders (kafka_key, value) VALUES ($1, $2)`,
		string(key), string(value),
	)
	if err != nil {
		log.Printf("insert failed: %v", err)
	}
	return err
}
