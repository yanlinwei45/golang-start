package models

import (
	"database/sql"
	"errors"
	"time"
)

// Product 产品模型
type Product struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetProductByID 根据 ID 获取产品
func GetProductByID(db *sql.DB, id int) (*Product, error) {
	query := `SELECT id, name, price, stock, created_at, updated_at FROM products WHERE id = ?`
	row := db.QueryRow(query, id)

	var product Product
	err := row.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("product not found")
	} else if err != nil {
		return nil, err
	}

	return &product, nil
}

// GetAllProducts 获取所有产品
func GetAllProducts(db *sql.DB) ([]*Product, error) {
	query := `SELECT id, name, price, stock, created_at, updated_at FROM products ORDER BY id ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// CreateProduct 创建产品
func CreateProduct(db *sql.DB, product *Product) (*Product, error) {
	query := `INSERT INTO products (name, price, stock, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	result, err := db.Exec(query, product.Name, product.Price, product.Stock, time.Now(), time.Now())
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	product.ID = int(id)
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	return product, nil
}

// UpdateProduct 更新产品
func UpdateProduct(db *sql.DB, product *Product) (*Product, error) {
	query := `UPDATE products SET name = ?, price = ?, stock = ?, updated_at = ? WHERE id = ?`
	result, err := db.Exec(query, product.Name, product.Price, product.Stock, time.Now(), product.ID)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, errors.New("product not found")
	}

	product.UpdatedAt = time.Now()
	return product, nil
}

// DeleteProduct 删除产品
func DeleteProduct(db *sql.DB, id int) error {
	query := `DELETE FROM products WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}