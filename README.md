# Go 语言初学者项目

这是一个为 Go 语言初学者设计的完整项目示例，包含了 HTTP 服务器、API 接口、数据库操作和测试代码。

## 项目结构

```
golang-starter/
├── main.go              # 项目入口文件
├── handlers/
│   └── products.go      # API 接口处理函数
├── models/
│   └── products.go      # 数据模型和数据库操作
├── utils/
│   └── database.go      # 数据库连接和初始化
├── tests/
│   └── products_test.go # 测试代码
├── go.mod               # Go 模块文件
└── README.md            # 项目说明文档
```

## 功能特性

- 🏗️ 完整的 RESTful API
- 📦 SQLite 数据库操作
- 🧪 完整的单元测试
- 📚 清晰的代码结构
- 🔧 现代的 Go 语言最佳实践

## 快速开始

### 1. 安装依赖

```bash
cd golang-starter
go mod download
```

### 2. 运行项目

```bash
go run main.go
```

服务器将在 http://localhost:8080 启动。

### 3. 运行测试

```bash
go test ./tests -v
```

## API 文档

### 健康检查

```
GET /api/health
```

**响应**：
```json
{"status": "ok"}
```

### 获取所有产品

```
GET /api/products
```

**响应**：
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "Test Product",
      "price": 99.99,
      "stock": 10,
      "created_at": "2024-01-28T10:00:00Z",
      "updated_at": "2024-01-28T10:00:00Z"
    }
  ]
}
```

### 获取单个产品

```
GET /api/products/{id}
```

**响应**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "Test Product",
    "price": 99.99,
    "stock": 10,
    "created_at": "2024-01-28T10:00:00Z",
    "updated_at": "2024-01-28T10:00:00Z"
  }
}
```

### 创建产品

```
POST /api/products
```

**请求体**：
```json
{
  "name": "New Product",
  "price": 199.99,
  "stock": 5
}
```

**响应**：
```json
{
  "code": 201,
  "message": "product created",
  "data": {
    "id": 2,
    "name": "New Product",
    "price": 199.99,
    "stock": 5,
    "created_at": "2024-01-28T10:00:00Z",
    "updated_at": "2024-01-28T10:00:00Z"
  }
}
```

### 更新产品

```
PUT /api/products/{id}
```

**请求体**：
```json
{
  "name": "Updated Product",
  "price": 159.99,
  "stock": 8
}
```

**响应**：
```json
{
  "code": 200,
  "message": "product updated",
  "data": {
    "id": 1,
    "name": "Updated Product",
    "price": 159.99,
    "stock": 8,
    "created_at": "2024-01-28T10:00:00Z",
    "updated_at": "2024-01-28T10:00:00Z"
  }
}
```

### 删除产品

```
DELETE /api/products/{id}
```

**响应**：
```json
{
  "code": 200,
  "message": "product deleted"
}
```

## 技术栈

- **Go 1.21+** - 编程语言
- **net/http** - 官方 HTTP 服务器库
- **SQLite3** - 数据库
- **github.com/mattn/go-sqlite3** - SQLite 驱动
- **testing** - 官方测试库

## 学习建议

1. **从 `main.go` 开始**：理解项目的入口点和服务器启动流程
2. **查看 `utils/database.go`**：了解数据库连接和初始化
3. **学习 `models/products.go`**：掌握数据模型和数据库操作
4. **研究 `handlers/products.go`**：理解 API 接口设计
5. **运行 `tests/products_test.go`**：学习测试写法

## 下一步学习

- 学习使用 Gin 或 Echo 等 HTTP 框架
- 尝试使用 Postgres 或 MySQL 数据库
- 学习中间件开发（CORS、身份验证、日志）
- 了解 Go 的并发编程（Goroutine、Channel）
- 学习 Docker 容器化部署

## 参考资源

- [Go 官方文档](https://golang.org/doc/)
- [Go by Example](https://gobyexample.com/)
- [The Go Programming Language](https://www.gopl.io/)