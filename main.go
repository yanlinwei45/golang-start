package main

import (
	"log"
	"net/http"

	"golang-starter/handlers"
	"golang-starter/utils"
)

func main() {
	// 初始化数据库连接
	utils.InitDB()
	defer utils.CloseDB()

	// 创建 HTTP 路由器
	mux := http.NewServeMux()

	// 注册路由
	handlers.RegisterRoutes(mux)

	// 启动服务器
	port := ":8080"
	log.Printf("Server is running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}