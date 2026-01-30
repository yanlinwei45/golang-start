package main

// import 块：测试代码依赖的包。
import (
	// encoding/json：解析 handler 返回的 JSON 响应，提取字段用于断言。
	"encoding/json"
	// fmt：用于格式化 URL（例如 /api/products/%d）。
	"fmt"
	// io：读取响应体内容（用于失败时输出 body，方便排查）。
	"io"
	// net/http：HTTP 状态码常量（StatusOK/StatusCreated/...）。
	"net/http"
	// net/http/httptest：标准库测试工具，用于模拟 Request/ResponseWriter，不需要真实起服务监听端口。
	"net/http/httptest"
	// strings：把 JSON 字符串转成 io.Reader 作为请求体。
	"strings"
	// testing：Go 官方测试框架（t *testing.T）。
	"testing"

	// handlers：注册路由与处理函数（被测对象）。
	"golang-starter/handlers"
	// utils：初始化与关闭数据库（测试会真实访问 SQLite 文件 DB）。
	"golang-starter/utils"
)

func clearProductsTable(t *testing.T) {
	t.Helper()

	if _, err := utils.DB.Exec("DELETE FROM products"); err != nil {
		t.Fatalf("failed to clear products table: %v", err)
	}
}

// 初始化测试数据库
func setupTestDB() func() {
	// 初始化全局 DB（打开 SQLite 文件并建表）。
	utils.InitDB()
	// 清空 products 表，避免不同测试/不同运行之间互相污染导致用例不稳定。
	// 这里必须传入当前测试的 *testing.T 才能在失败时正确终止用例；
	// setupTestDB 没有拿到 t，因此用 panic 的方式暴露错误（比静默忽略更安全）。
	if _, err := utils.DB.Exec("DELETE FROM products"); err != nil {
		panic(err)
	}
	// 返回 teardown 函数，供每个测试 defer 调用，确保资源释放。
	return func() {
		// 关闭全局 DB 连接池句柄。
		utils.CloseDB()
	}
}

// TestHealthCheck 测试健康检查接口
func TestHealthCheck(t *testing.T) {
	// setup：初始化数据库（虽然 health 本身不依赖 DB，但这里保持测试套路一致）。
	teardown := setupTestDB()
	// defer：确保测试结束后关闭 DB。
	defer teardown()

	// 构造一个 GET 请求：不会真的走网络，只是一个 *http.Request 对象。
	req := httptest.NewRequest("GET", "/api/health", nil)
	// NewRecorder：一个假的 ResponseWriter，用来记录 handler 写出的 status/header/body。
	w := httptest.NewRecorder()
	// 直接调用 handler：因为 HealthCheck 不依赖路由分发。
	handlers.HealthCheck(w, req)

	// Result：把 recorder 里的内容转成 *http.Response，便于按真实响应读取。
	resp := w.Result()
	// 关闭响应体 reader（释放资源）。
	defer resp.Body.Close()

	// 断言状态码为 200 OK。
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// 统一成功响应结构：{"code":200,"message":"success","data":{"status":"ok"}}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", result["data"])
	}

	status, ok := data["status"].(string)
	if !ok {
		t.Fatalf("expected data.status string, got %T", data["status"])
	}

	if status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", status)
	}
}

// TestCreateProduct 测试创建产品接口
func TestCreateProduct(t *testing.T) {
	// 初始化测试 DB。
	teardown := setupTestDB()
	defer teardown()

	// product：作为请求体的 JSON 字符串（name/price/stock）。
	product := `{"name": "Test Product", "price": 99.99, "stock": 10}`
	// 构造 POST 请求，请求体来自 strings.NewReader（实现了 io.Reader）。
	req := httptest.NewRequest("POST", "/api/products", strings.NewReader(product))
	// 设置 Content-Type，虽然当前 handler 不强制检查，但这是规范写法。
	req.Header.Set("Content-Type", "application/json")
	// recorder：捕获响应。
	w := httptest.NewRecorder()

	// 使用路由处理函数
	// mux：模拟真实服务端路由器，确保请求走完整的路由分发逻辑。
	mux := http.NewServeMux()
	// 注册 API 路由到 mux。
	handlers.RegisterRoutes(mux)
	// ServeHTTP：把请求交给 mux 分发并写入 recorder。
	mux.ServeHTTP(w, req)

	// 获取响应对象。
	resp := w.Result()
	defer resp.Body.Close()

	// 断言创建成功状态码 201 Created；失败时打印 body 方便定位。
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusCreated, resp.StatusCode, body)
	}
}

// TestGetAllProducts 测试获取所有产品接口
func TestGetAllProducts(t *testing.T) {
	// 初始化测试 DB。
	teardown := setupTestDB()
	defer teardown()

	// 先创建一个产品
	// 目的：保证列表接口返回非空，避免“空列表也算对”的假阳性。
	createProduct(t)

	// 构造 GET /api/products 请求。
	req := httptest.NewRequest("GET", "/api/products", nil)
	// recorder：捕获响应。
	w := httptest.NewRecorder()

	// 使用路由处理函数
	// 走路由分发以覆盖 RegisterRoutes 的逻辑。
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	// 获取响应对象并关闭 body。
	resp := w.Result()
	defer resp.Body.Close()

	// 断言 200 OK；失败时打印 body。
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}
}

// TestGetProduct 测试获取单个产品接口
func TestGetProduct(t *testing.T) {
	// 初始化测试 DB。
	teardown := setupTestDB()
	defer teardown()

	// 先创建一个产品
	// productID：后续用于拼出 /api/products/{id} 路径。
	productID := createProduct(t)

	// 构造 GET /api/products/{id} 请求；fmt.Sprintf 格式化字符串。
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/products/%d", productID), nil)
	// recorder：捕获响应。
	w := httptest.NewRecorder()

	// 使用路由处理函数
	// 创建 mux 并注册路由。
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	// 分发请求。
	mux.ServeHTTP(w, req)

	// 获取响应对象。
	resp := w.Result()
	defer resp.Body.Close()

	// 断言 200 OK。
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}
}

// TestUpdateProduct 测试更新产品接口
func TestUpdateProduct(t *testing.T) {
	// 初始化测试 DB。
	teardown := setupTestDB()
	defer teardown()

	// 先创建一个产品
	// productID：要更新的目标产品 id。
	productID := createProduct(t)

	// updatedProduct：更新请求体 JSON（新的 name/price/stock）。
	updatedProduct := `{"name": "Updated Product", "price": 199.99, "stock": 5}`
	// 构造 PUT /api/products/{id} 请求，并带上 JSON body。
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/products/%d", productID), strings.NewReader(updatedProduct))
	// 设置 Content-Type 为 application/json。
	req.Header.Set("Content-Type", "application/json")

	// recorder：捕获响应。
	w := httptest.NewRecorder()
	// 使用路由处理函数
	// 创建 mux、注册路由、分发请求。
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	// 获取响应对象。
	resp := w.Result()
	defer resp.Body.Close()

	// 断言 200 OK；失败时打印 body。
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}
}

// TestDeleteProduct 测试删除产品接口
func TestDeleteProduct(t *testing.T) {
	// 初始化测试 DB。
	teardown := setupTestDB()
	defer teardown()

	// 先创建一个产品
	// productID：要删除的目标产品 id。
	productID := createProduct(t)

	// 构造 DELETE /api/products/{id} 请求。
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/products/%d", productID), nil)
	// recorder：捕获响应。
	w := httptest.NewRecorder()

	// 使用路由处理函数
	// 创建 mux 并注册路由。
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	// 分发删除请求。
	mux.ServeHTTP(w, req)

	// 获取响应对象。
	resp := w.Result()
	defer resp.Body.Close()

	// 断言删除成功返回 200 OK。
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}

	// 验证产品是否已删除
	// 再发一个 GET 请求，期望得到 404 Not Found。
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/products/%d", productID), nil)
	// 新的 recorder 用于记录第二次请求响应。
	w = httptest.NewRecorder()
	// 使用路由处理函数
	// 重新创建 mux 并注册路由（简单但略重复；学习项目可接受）。
	mux = http.NewServeMux()
	handlers.RegisterRoutes(mux)
	// 分发 GET 请求。
	mux.ServeHTTP(w, req)

	// w.Code：Recorder 的状态码字段；断言为 404。
	if w.Code != http.StatusNotFound {
		t.Errorf("expected product not found, got status %d", w.Code)
	}
}

func TestSearchProducts(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	// 创建两条产品：一条匹配，一条不匹配。
	createProductWithName(t, "Apple")
	createProductWithName(t, "Banana")

	req := httptest.NewRequest("GET", "/api/products/search?name=App", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data array, got %T", result["data"])
	}
	if len(data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(data))
	}

	first, ok := data[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected first element object, got %T", data[0])
	}
	if first["name"] != "Apple" {
		t.Fatalf("expected first.name Apple, got %v", first["name"])
	}
}

func TestSearchProductsMissingName(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	req := httptest.NewRequest("GET", "/api/products/search", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("expected json error response, decode failed: %v", err)
	}
}

func TestBulkCreateProductsSuccess(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	payload := `[{"name":"A","price":10.0,"stock":1},{"name":"B","price":20.0,"stock":0}]`
	req := httptest.NewRequest("POST", "/api/products/bulk", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusCreated, resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data array, got %T", result["data"])
	}
	if len(data) != 2 {
		t.Fatalf("expected 2 created products, got %d", len(data))
	}

	// 检查每条都有 id 且 name 正确。
	names := map[string]bool{}
	for i := range data {
		item, ok := data[i].(map[string]interface{})
		if !ok {
			t.Fatalf("expected element object, got %T", data[i])
		}
		if _, ok := item["id"].(float64); !ok {
			t.Fatalf("expected element.id number, got %T", item["id"])
		}
		if name, ok := item["name"].(string); ok {
			names[name] = true
		}
	}
	if !names["A"] || !names["B"] {
		t.Fatalf("expected created names A and B, got %v", names)
	}
}

func TestBulkCreateProductsRollbackOnInvalidItem(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	// 先插入一条基准数据，方便验证回滚不会改变总数（除了这条基准数据）。
	createProductWithName(t, "Baseline")

	// 第二个元素非法（price<=0），期望整体 400，且不会插入第一条合法数据。
	payload := `[{"name":"Good","price":10.0,"stock":1},{"name":"Bad","price":0,"stock":1}]`
	req := httptest.NewRequest("POST", "/api/products/bulk", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.StatusCode, body)
	}

	// 再拉全量列表，确保只存在 Baseline，不存在 Good。
	req = httptest.NewRequest("GET", "/api/products", nil)
	w = httptest.NewRecorder()
	mux = http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var listResult map[string]interface{}
	if err := json.NewDecoder(w.Result().Body).Decode(&listResult); err != nil {
		t.Fatalf("failed to decode list response json: %v", err)
	}
	listData, ok := listResult["data"].([]interface{})
	if !ok {
		t.Fatalf("expected list data array, got %T", listResult["data"])
	}
	if len(listData) != 1 {
		t.Fatalf("expected 1 product after rollback, got %d", len(listData))
	}
	first, ok := listData[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected first element object, got %T", listData[0])
	}
	if first["name"] != "Baseline" {
		t.Fatalf("expected remaining product Baseline, got %v", first["name"])
	}
}

func TestPatchProductPriceOnlySuccess(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	id := createProductWithName(t, "PatchTarget")

	payload := `{"price":123.45}`
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/products/%d", id), strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusOK, resp.StatusCode, body)
	}
}

func TestPatchProductEmptyBodyReturns400(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	id := createProductWithName(t, "PatchTarget")

	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/products/%d", id), strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusBadRequest, resp.StatusCode, body)
	}
}

func TestPatchProductNotFoundReturns404(t *testing.T) {
	teardown := setupTestDB()
	defer teardown()

	payload := `{"price":123.45}`
	req := httptest.NewRequest("PATCH", "/api/products/99999999", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusNotFound, resp.StatusCode, body)
	}
}

// 辅助函数：创建测试产品
func createProduct(t *testing.T) int {
	// t.Helper：标记为 helper，失败时更友好地定位到调用处而不是 helper 内部。
	t.Helper()

	// 构造创建产品的 JSON 请求体。
	product := `{"name": "Test Product", "price": 99.99, "stock": 10}`
	// 构造 POST 请求并携带 body。
	req := httptest.NewRequest("POST", "/api/products", strings.NewReader(product))
	// 设置 Content-Type。
	req.Header.Set("Content-Type", "application/json")
	// recorder：捕获响应。
	w := httptest.NewRecorder()

	// 创建 mux 并注册路由。
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	// 分发创建请求。
	mux.ServeHTTP(w, req)

	// result：用于接收响应 JSON（code/message/data）。
	var result map[string]interface{}
	// 解码响应体到 map；注意：数字默认解成 float64。
	json.NewDecoder(w.Result().Body).Decode(&result)

	// data：响应中的 data 字段，类型是对象，因此断言为 map[string]interface{}。
	data := result["data"].(map[string]interface{})
	// id：从 float64 转成 int（encoding/json 默认规则）。
	return int(data["id"].(float64))
}

func createProductWithName(t *testing.T, name string) int {
	t.Helper()

	product := fmt.Sprintf(`{"name": %q, "price": 99.99, "stock": 10}`, name)
	req := httptest.NewRequest("POST", "/api/products", strings.NewReader(product))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Result().Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode create response json: %v", err)
	}

	data := result["data"].(map[string]interface{})
	return int(data["id"].(float64))
}
