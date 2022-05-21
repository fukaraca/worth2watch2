package db

import "context"

//todo to be deleted
func Truncate() {
	conn.Exec(context.Background(), "TRUNCATE Table users,movies,episodes,favorite_movies,favorite_series,seasons,series cascade ;")
}
