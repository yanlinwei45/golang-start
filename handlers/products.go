package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"golang-starter/models"
	"golang-starter/utils"
)

// RegisterRoutes 注册所有 API 路由
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health", HealthCheck)
	mux.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			GetAllProducts(w, r)
		case "POST":
			CreateProduct(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/products/", HandleProduct)
}

// HandleProduct 处理产品相关的请求（GET / PUT / DELETE）
func HandleProduct(w http.ResponseWriter, r *http.Request) {
	// 从路径中提取 ID（如 /api/products/123 中的 123）
	path := strings.TrimPrefix(r.URL.Path, "/api/products/")
	idStr := strings.SplitN(path, "/", 2)[0]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		GetProduct(w, r, id)
	case "PUT":
		UpdateProduct(w, r, id)
	case "DELETE":
		DeleteProduct(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HealthCheck 健康检查接口
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetAllProducts 获取所有产品列表
func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := models.GetAllProducts(utils.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    http.StatusOK,
		"message": "success",
		"data":    products,
	})
}

// GetProduct 根据 ID 获取产品
func GetProduct(w http.ResponseWriter, r *http.Request, id int) {
	product, err := models.GetProductByID(utils.DB, id)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    http.StatusOK,
		"message": "success",
		"data":    product,
	})
}

// CreateProduct 创建产品
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 验证必填字段
	if product.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if product.Price <= 0 {
		http.Error(w, "price must be greater than 0", http.StatusBadRequest)
		return
	}

	if product.Stock < 0 {
		http.Error(w, "stock cannot be negative", http.StatusBadRequest)
		return
	}

	createdProduct, err := models.CreateProduct(utils.DB, &product)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    http.StatusCreated,
		"message": "product created",
		"data":    createdProduct,
	})
}

// UpdateProduct 更新产品
func UpdateProduct(w http.ResponseWriter, r *http.Request, id int) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	product.ID = id

	// 验证必填字段
	if product.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if product.Price <= 0 {
		http.Error(w, "price must be greater than 0", http.StatusBadRequest)
		return
	}

	if product.Stock < 0 {
		http.Error(w, "stock cannot be negative", http.StatusBadRequest)
		return
	}

	updatedProduct, err := models.UpdateProduct(utils.DB, &product)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    http.StatusOK,
		"message": "product updated",
		"data":    updatedProduct,
	})
}

// DeleteProduct 删除产品
func DeleteProduct(w http.ResponseWriter, r *http.Request, id int) {
	err := models.DeleteProduct(utils.DB, id)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    http.StatusOK,
		"message": "product deleted",
	})
}