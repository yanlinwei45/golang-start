package main

// import 块：显式列出本文件依赖的包；Go 会据此做编译依赖分析与裁剪。
import (
	// log：标准库日志输出，用于打印启动信息与致命错误。
	"log"
	// net/http：标准库 HTTP 服务端/客户端；这里用它来启动 HTTP Server 与路由分发。
	"net/http"

	// handlers：HTTP 路由注册与各接口处理函数（controller/handler 层）。
	"golang-starter/handlers"
	// utils：项目工具包；这里主要提供数据库初始化与关闭（全局 DB）。
	"golang-starter/utils"
)

// main：程序入口；负责初始化资源、注册路由并启动 HTTP 服务。
func main() {
	// 初始化数据库连接：打开 SQLite 文件并确保表存在。
	utils.InitDB()
	// defer：在 main 返回时执行；用于释放数据库资源（注意：log.Fatal 会 os.Exit，不会执行 defer）。
	defer utils.CloseDB()

	// 创建 HTTP 路由器：ServeMux 根据 URL path 匹配并调用对应 handler。
	mux := http.NewServeMux()

	// 注册路由：把各 API path（/api/health、/api/products...）绑定到 handler 函数。
	handlers.RegisterRoutes(mux)

	// 启动服务器：监听端口并开始处理请求；ListenAndServe 只有在启动失败或异常退出时才返回 error。
	port := ":8080"
	// 打印可访问的地址提示（方便开发时快速访问）。
	log.Printf("Server is running on http://localhost%s\n", port)
	// log.Fatal：打印错误并退出进程（内部调用 os.Exit(1)，因此不会触发 defer）。
	log.Fatal(http.ListenAndServe(port, mux))
}
