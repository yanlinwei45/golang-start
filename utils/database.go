package utils

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// DB 全局数据库连接
var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 检查数据库连接
	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Database connection established successfully")

	// 创建表
	createTables()
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if err := DB.Close(); err != nil {
		log.Println("Error closing database connection:", err)
	} else {
		log.Println("Database connection closed")
	}
}

func createTables() {
	productTableSQL := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		price REAL NOT NULL,
		stock INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`

	_, err := DB.Exec(productTableSQL)
	if err != nil {
		log.Fatal("Failed to create products table:", err)
	}

	log.Println("Products table initialized")
}