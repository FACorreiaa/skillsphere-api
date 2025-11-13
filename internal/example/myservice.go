package example

import "time"

// MyServiceRecord represents a record in the database
type MyServiceRecord struct {
	ID        int64     `db:"id"`
	Input     string    `db:"input"`
	Output    string    `db:"output"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
