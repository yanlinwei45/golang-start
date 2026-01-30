package handlers

// import 块：handler 层依赖的标准库与项目内包。
import (
	// encoding/json：用于 JSON 编解码（请求体解析、响应体输出）。
	"encoding/json"
	"errors"
	"fmt"

	// net/http：HTTP handler 所需的核心类型与工具函数（ResponseWriter、Request、StatusCode、http.Error）。
	"net/http"
	// strconv：字符串与数值转换；这里用 Atoi 把 path 中的 id 解析成 int。
	"strconv"
	// strings：字符串处理；这里用于从 URL path 中裁剪前缀与拆分片段。
	"strings"

	// models：数据模型与数据库 CRUD 操作（对 SQLite 的增删改查）。
	"golang-starter/models"
	// utils：提供全局数据库连接 utils.DB（在 main 启动时 InitDB 初始化）。
	"golang-starter/utils"
)

// RegisterRoutes 注册所有 API 路由。
// 约定：同一个 path 用不同 HTTP Method 表示不同动作（GET 列表 / POST 创建 / GET 单个 / PUT 更新 / DELETE 删除）。
func RegisterRoutes(mux *http.ServeMux) {
	// /api/health：健康检查接口（一般用于探活、负载均衡检查等）。
	mux.HandleFunc("/api/health", HealthCheck)
	// /api/products：集合资源路径；用 method 区分 GET（列表）与 POST（创建）。
	mux.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		// r.Method：HTTP 方法字符串，例如 "GET"、"POST"、"PUT"、"DELETE"。
		switch r.Method {
		case "GET":
			// GET /api/products：返回所有产品。
			GetAllProducts(w, r)
		case "POST":
			// POST /api/products：创建一个产品。
			CreateProduct(w, r)
		default:
			// 其他方法不支持：返回 405 Method Not Allowed。
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
	// /api/products/：注意以 "/" 结尾时，ServeMux 会做前缀匹配；例如 /api/products/123 会进入 HandleProduct。
	mux.HandleFunc("/api/products/search", SearchProducts)
	mux.HandleFunc("/api/products/bulk", ProductBulk)
	mux.HandleFunc("/api/products/", HandleProduct)
}

// HandleProduct 处理单个产品资源的请求（GET / PUT / DELETE）。
// 该 handler 负责：
// 1) 从 path 中解析 id
// 2) 根据 method 分发到具体处理函数
func HandleProduct(w http.ResponseWriter, r *http.Request) {
	// 从路径中提取 ID（如 /api/products/123 中的 123）
	// r.URL.Path：不包含 querystring（?a=b），只包含路径部分。
	path := strings.TrimPrefix(r.URL.Path, "/api/products/")
	// SplitN(..., 2)：最多拆两段；取 [0] 得到第一个 segment，即 id 字符串。
	idStr := strings.SplitN(path, "/", 2)[0]

	// Atoi：把十进制字符串转 int；失败则说明 id 不是合法数字。
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// 400 Bad Request：客户端传参不合法（id 不是数字）。
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// 按 method 分发到单个资源的 CRUD 操作。
	switch r.Method {
	case "GET":
		// GET /api/products/{id}：查询单个产品。
		GetProduct(w, r, id)
	case "PUT":
		// PUT /api/products/{id}：更新单个产品（body 提供 name/price/stock）。
		UpdateProduct(w, r, id)
	case "DELETE":
		// DELETE /api/products/{id}：删除单个产品。
		DeleteProduct(w, r, id)
	default:
		// 不支持的方法返回 405。
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type healthResponse struct {
	Status string `json:"status"`
}

// HealthCheck 健康检查接口：返回固定 JSON，表明服务可用。
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, http.StatusOK, successResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    healthResponse{Status: "ok"},
	})
}

// GetAllProducts 获取所有产品列表：调用 model 层查询 DB，并以 JSON 形式返回。
func GetAllProducts(w http.ResponseWriter, r *http.Request) {
	// utils.DB：全局数据库连接（*sql.DB，实际上是连接池句柄），并发安全。
	products, err := models.GetAllProducts(utils.DB)
	if err != nil {
		// 500：服务端错误（例如 DB 查询失败、SQL 语法错误、连接异常等）。
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, successResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    products,
	})
}

// GetProduct 根据 ID 获取产品：查到则 200 + data；不存在则 404。
func GetProduct(w http.ResponseWriter, r *http.Request, id int) {
	// 调用 model 层按 id 查询。
	product, err := models.GetProductByID(utils.DB, id)
	if err != nil {
		// 通过错误消息区分“未找到”和“内部错误”（学习项目的简化写法）。
		if errors.Is(err, models.ErrProductNotFound) {
			// 404 Not Found：资源不存在。
			writeError(w, http.StatusNotFound, "product not found")
		} else {
			// 500 Internal Server Error：数据库或其他内部错误。
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeSuccess(w, http.StatusOK, successResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    product,
	})

}

// CreateProduct 创建产品：
// 1) 从 body 解析 JSON 到 Product
// 2) 校验字段
// 3) 调用 model 层写入 DB
// 4) 返回 201 + 创建后的对象
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	// product：用于接收请求体 JSON 解码后的结果。
	var product models.Product
	// NewDecoder(r.Body)：从请求体流读取 JSON；Decode(&product) 需要传指针才能写入字段。
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		// 400：请求体不是合法 JSON，或字段类型不匹配导致解码失败。
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// 关闭请求体（释放资源）；defer 确保函数返回时执行。
	defer r.Body.Close()

	// 验证必填字段
	if product.Name == "" {
		// name 为空：业务校验失败，返回 400。
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if product.Price <= 0 {
		// price 必须为正数：避免无效数据进入 DB。
		writeError(w, http.StatusBadRequest, "price must be greater than 0")
		return
	}

	if product.Stock < 0 {
		// stock 不能为负：库存不能小于 0。
		writeError(w, http.StatusBadRequest, "stock cannot be negative")
		return
	}

	// 调用 model 层创建产品；成功后会填充 ID/时间字段。
	createdProduct, err := models.CreateProduct(utils.DB, &product)
	if err != nil {
		// 500：插入失败（例如数据库写入错误）。
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, successResponse{
		Code:    http.StatusCreated,
		Message: "success",
		Data:    createdProduct,
	})
}

// UpdateProduct 更新产品：
// - id 来自 URL path（避免客户端在 body 里伪造 id）
// - body 提供 name/price/stock
// - 若 id 不存在则返回 404
func UpdateProduct(w http.ResponseWriter, r *http.Request, id int) {
	// product：用于接收请求体中的更新字段。
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		// 400：JSON 解码失败。
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// 关闭请求体。
	defer r.Body.Close()

	// 强制使用 path 中的 id，覆盖 body 中可能存在的 id。
	product.ID = id

	// 验证必填字段
	if product.Name == "" {
		// name 为空：返回 400。
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if product.Price <= 0 {
		// price 非正：返回 400。
		writeError(w, http.StatusBadRequest, "price must be greater than 0")
		return
	}

	if product.Stock < 0 {
		// stock 为负：返回 400。
		writeError(w, http.StatusBadRequest, "stock cannot be negative")
		return
	}

	// 调用 model 层执行 UPDATE；rowsAffected==0 时会返回 "product not found"。
	updatedProduct, err := models.UpdateProduct(utils.DB, &product)
	if err != nil {
		if errors.Is(err, models.ErrProductNotFound) {
			// 404：要更新的资源不存在。
			writeError(w, http.StatusNotFound, "product not found")
		} else {
			// 500：其他内部错误。
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeSuccess(w, http.StatusOK, successResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    updatedProduct,
	})

}

// DeleteProduct 删除产品：如果不存在返回 404；成功返回 200。
func DeleteProduct(w http.ResponseWriter, r *http.Request, id int) {
	// 调用 model 层删除；内部通过 rowsAffected 判断是否真的删除到数据。
	err := models.DeleteProduct(utils.DB, id)
	if err != nil {
		if errors.Is(err, models.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "product not found")
		} else {
			// 500：删除过程的内部错误。
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeSuccess(w, http.StatusOK, successResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    nil,
	})
}

func SearchProducts(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	products, err := models.SearchProduct(utils.DB, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, successResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    products,
	})
}

func ProductBulk(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	defer r.Body.Close()

	var products []models.Product

	if err := json.NewDecoder(r.Body).Decode(&products); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(products) == 0 {
		writeError(w, http.StatusBadRequest, "products is empty")
		return
	}

	for i, product := range products {
		if product.Name == "" {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("products[%d].name is required", i))
			return
		}
		if product.Price <= 0 {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("products[%d].price must be greater than 0", i))
			return
		}
		if product.Stock < 0 {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("products[%d].stock cannot be negative", i))
			return
		}
	}

	productPtrs := make([]*models.Product, 0, len(products))
	for i := range products {
		productPtrs = append(productPtrs, &products[i])
	}

	created, err := models.ProductsBulk(utils.DB, productPtrs)

	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, successResponse{
		Code:    http.StatusCreated,
		Message: "success",
		Data:    created,
	})
}
