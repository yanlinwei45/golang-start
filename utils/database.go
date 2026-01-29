package utils

// import 块：数据库工具层依赖的包。
import (
	// database/sql：Go 标准库的通用 SQL 接口层（连接池、Query/Exec、Rows/Row 等）。
	"database/sql"
	// log：输出数据库初始化/关闭过程中的日志信息。
	"log"

	// sqlite3 driver：用下划线导入触发 init() 注册驱动到 database/sql。
	// 如果缺少这一行，sql.Open("sqlite3", ...) 会报 unknown driver。
	_ "github.com/mattn/go-sqlite3"
)

// DB 全局数据库连接
// 注意：*sql.DB 表示“连接池句柄”，不是单个连接；它是并发安全的。
var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() {
	// err：用于接收后续 Open/Ping/Exec 的错误。
	var err error
	// sql.Open：创建 *sql.DB 句柄；对不少驱动而言，此时未必真正建立连接，因此需要 Ping 验证。
	DB, err = sql.Open("sqlite3", "./database.db")
	if err != nil {
		// log.Fatal：打印日志并退出进程；用于“无法启动就直接失败”的场景。
		log.Fatal("Failed to connect to database:", err)
	}

	// 检查数据库连接
	// Ping：确认能够与数据库正常通信（SQLite 是文件型 DB，这里也可用于验证打开是否成功）。
	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// 提示：数据库连接建立成功。
	log.Println("Database connection established successfully")

	// 创建表
	// 启动时建表：使用 IF NOT EXISTS，重复启动不会报错（适合学习项目）。
	createTables()
}

// CloseDB 关闭数据库连接
func CloseDB() {
	// Close：关闭连接池并释放资源；关闭后 DB 不应再被使用。
	if err := DB.Close(); err != nil {
		// 关闭失败一般不致命，但应记录下来，便于排查资源泄露等问题。
		log.Println("Error closing database connection:", err)
	} else {
		// 关闭成功提示。
		log.Println("Database connection closed")
	}
}

func createTables() {
	// productTableSQL：建表语句（多行字符串用反引号，便于写 SQL）。
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

	// Exec：执行不返回结果集的 SQL（CREATE/INSERT/UPDATE/DELETE 常用）。
	_, err := DB.Exec(productTableSQL)
	if err != nil {
		// 建表失败说明应用无法正常工作（缺少表），直接终止启动更容易发现问题。
		log.Fatal("Failed to create products table:", err)
	}

	// 提示：products 表已初始化完成。
	log.Println("Products table initialized")
}
