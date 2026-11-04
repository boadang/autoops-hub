package storage

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

type DBManager struct {
	Conn *sql.DB
}

// InitDB khởi tạo kết nối SQLite và tạo bảng lịch sử nếu chưa có
var dbPath = "autoops.db"

func NewDBManager(dbPath string) (*DBManager, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Kiểm tra kết nối
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Tạo bảng lưu lịch sử thực thi task
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS execution_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT,
		name TEXT,
		output TEXT,
		error TEXT,
		elapsed_ms INTEGER,
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	log.Println("🗄️ Đã kết nối thành công với SQLite và khởi tạo bảng execution_logs.")
	return &DBManager{Conn: db}, nil
}

// SaveExecutionResult lưu kết quả chạy task vào database
func (db *DBManager) SaveExecutionResult(taskID, name, output, errStr string, elapsed time.Duration) error {
	query := `INSERT INTO execution_logs (task_id, name, output, error, elapsed_ms) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Conn.Exec(query, taskID, name, output, errStr, elapsed.Milliseconds())
	return err
}
