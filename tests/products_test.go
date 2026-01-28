package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang-starter/handlers"
	"golang-starter/utils"
)

// 初始化测试数据库
func setupTestDB() func() {
	utils.InitDB()
	return func() {
		utils.CloseDB()
	}
}

// TestHealthCheck 测试健康检查接口
func TestHealthCheck(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	handlers.HealthCheck(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", result["status"])
	}
}

// TestCreateProduct 测试创建产品接口
func TestCreateProduct(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	product := `{"name": "Test Product", "price": 99.99, "stock": 10}`
	req := httptest.NewRequest("POST", "/api/products", strings.NewReader(product))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 使用路由处理函数
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusCreated, resp.StatusCode, body)
	}
}

// TestGetAllProducts 测试获取所有产品接口
func TestGetAllProducts(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	// 先创建一个产品
	createProduct(t)

	req := httptest.NewRequest("GET", "/api/products", nil)
	w := httptest.NewRecorder()

	// 使用路由处理函数
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}
}

// TestGetProduct 测试获取单个产品接口
func TestGetProduct(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	// 先创建一个产品
	productID := createProduct(t)

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/products/%d", productID), nil)
	w := httptest.NewRecorder()

	// 使用路由处理函数
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}
}

// TestUpdateProduct 测试更新产品接口
func TestUpdateProduct(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	// 先创建一个产品
	productID := createProduct(t)

	updatedProduct := `{"name": "Updated Product", "price": 199.99, "stock": 5}`
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/products/%d", productID), strings.NewReader(updatedProduct))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	// 使用路由处理函数
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}
}

// TestDeleteProduct 测试删除产品接口
func TestDeleteProduct(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	// 先创建一个产品
	productID := createProduct(t)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/products/%d", productID), nil)
	w := httptest.NewRecorder()

	// 使用路由处理函数
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}

	// 验证产品是否已删除
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/products/%d", productID), nil)
	w = httptest.NewRecorder()
	// 使用路由处理函数
	mux = http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected product not found, got status %d", w.Code)
	}
}

// 辅助函数：创建测试产品
func createProduct(t *testing.T) int {
	t.Helper()

	product := `{"name": "Test Product", "price": 99.99, "stock": 10}`
	req := httptest.NewRequest("POST", "/api/products", strings.NewReader(product))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	var result map[string]interface{}
	json.NewDecoder(w.Result().Body).Decode(&result)

	data := result["data"].(map[string]interface{})
	return int(data["id"].(float64))
}