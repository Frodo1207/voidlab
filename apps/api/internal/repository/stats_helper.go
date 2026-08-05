package repository

import (
	"database/sql"
	"fmt"
)

func countTableRows(db *sql.DB, table string) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s", table)
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func countTableRowsByStatus(db *sql.DB, table string, status string) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE status = ?", table)
	if err := db.QueryRow(query, status).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}
