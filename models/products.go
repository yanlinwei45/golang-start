package models

// import 块：模型/数据访问层依赖。
import (
	// database/sql：提供 Query/Exec/Row/Rows 等，用于与具体 driver（sqlite3）交互。
	"database/sql"
	// errors：构造简单错误（本项目用 error message 来区分 not found）。
	"errors"
	// time：生成 created_at/updated_at 时间戳。
	"time"
)

// Product 产品模型
type Product struct {
	// ID：主键，自增；JSON 输出为 "id"。
	ID int `json:"id"`
	// Name：产品名称；JSON 输出为 "name"。
	Name string `json:"name"`
	// Price：产品价格；学习示例用 float64（生产环境常用整数分/decimal 以避免浮点误差）。
	Price float64 `json:"price"`
	// Stock：库存数量（非负整数）。
	Stock int `json:"stock"`
	// CreatedAt：创建时间；time.Time 会被 encoding/json 序列化为 RFC3339 格式字符串。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt：更新时间。
	UpdatedAt time.Time `json:"updated_at"`
}

var ErrProductNotFound = errors.New("product not found")

// GetProductByID 根据 ID 获取产品
func GetProductByID(db *sql.DB, id int) (*Product, error) {
	// query：参数化查询；使用 ? 占位符由 driver 安全绑定参数，避免 SQL 注入。
	query := `SELECT id, name, price, stock, created_at, updated_at FROM products WHERE id = ?`
	// QueryRow：预期最多返回一行；没有数据时 Scan 会返回 sql.ErrNoRows。
	row := db.QueryRow(query, id)

	// product：用于接收扫描结果。
	var product Product
	// Scan：按 SELECT 字段顺序把列值写入变量；必须传指针。
	err := row.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt)
	if err == sql.ErrNoRows {
		// 没有找到对应 id：返回业务层可识别的 not found 错误。
		return nil, ErrProductNotFound
	} else if err != nil {
		// 其他错误：例如 DB 连接问题、列类型不匹配等，直接向上返回。
		return nil, err
	}

	// 返回查询到的产品指针。
	return &product, nil
}

// GetAllProducts 获取所有产品
func GetAllProducts(db *sql.DB) ([]*Product, error) {
	// ORDER BY id ASC：保证返回顺序稳定（便于测试与客户端展示）。
	query := `SELECT id, name, price, stock, created_at, updated_at FROM products ORDER BY id ASC`
	// Query：返回多行结果集。
	rows, err := db.Query(query)
	if err != nil {
		// 查询失败：返回错误给上层处理（通常会转成 500）。
		return nil, err
	}
	// 关闭 rows 释放资源；defer 确保函数返回时执行。
	defer rows.Close()

	// products：用切片累积所有产品；这里存指针以减少复制开销（也符合常见 Go 写法）。
	var products []*Product
	// rows.Next：逐行迭代结果集。
	for rows.Next() {
		// product：每一行创建一个新的结构体变量用于接收 Scan。
		var product Product
		// Scan：读取当前行各列到结构体字段。
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		// 取地址追加到切片（此处 product 是每次循环的新变量，因此地址不会互相覆盖）。
		products = append(products, &product)
	}

	// rows.Err：检查迭代过程是否发生错误（例如中途读取失败）。
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 返回产品列表。
	return products, nil
}

// CreateProduct 创建产品
func CreateProduct(db *sql.DB, product *Product) (*Product, error) {
	// INSERT：写入 name/price/stock，同时写入 created_at 与 updated_at。
	query := `INSERT INTO products (name, price, stock, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	// Exec：执行写操作；返回 sql.Result 可用于获取 LastInsertId/RowsAffected。
	result, err := db.Exec(query, product.Name, product.Price, product.Stock, time.Now(), time.Now())
	if err != nil {
		return nil, err
	}

	// LastInsertId：获取 SQLite 自增主键值。
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// 回填主键到传入的结构体。
	product.ID = int(id)
	// 回填时间字段（注意：这里再次 time.Now()，与写入 DB 的时间可能有极小差异）。
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	// 返回创建后的产品对象（复用传入指针）。
	return product, nil
}

// UpdateProduct 更新产品
func UpdateProduct(db *sql.DB, product *Product) (*Product, error) {
	// UPDATE：根据 id 更新 name/price/stock，并更新 updated_at。
	query := `UPDATE products SET name = ?, price = ?, stock = ?, updated_at = ? WHERE id = ?`
	// Exec：执行更新操作。
	result, err := db.Exec(query, product.Name, product.Price, product.Stock, time.Now(), product.ID)
	if err != nil {
		return nil, err
	}

	// RowsAffected：更新影响的行数；若为 0 说明 id 不存在（或数据完全相同但仍可能返回 1，取决于 driver）。
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		// 没有任何行被更新：按业务语义认为资源不存在。
		return nil, ErrProductNotFound
	}

	// 回填更新时间（同样可能与 DB 中写入的时间有极小差异）。
	product.UpdatedAt = time.Now()
	// 返回更新后的对象（未重新 SELECT，因此其他字段以请求体为准）。
	return product, nil
}

// DeleteProduct 删除产品
func DeleteProduct(db *sql.DB, id int) error {
	// DELETE：按 id 删除一行。
	query := `DELETE FROM products WHERE id = ?`
	// Exec：执行删除操作。
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	// RowsAffected：删除影响的行数；为 0 表示 id 不存在。
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		// 没有删除到任何行：按业务语义认为资源不存在。
		return ErrProductNotFound
	}

	// 删除成功返回 nil error。
	return nil
}
