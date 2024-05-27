package repository

import "user-service/infrastructure/database"

type DB struct {
	conn *database.SqlDB
}

func NewRepository(conn *database.SqlDB) *DB {
	return &DB{conn: conn}
}
