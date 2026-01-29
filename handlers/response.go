package handlers

import (
	"encoding/json"
	"net/http"
)

// errorResponse 是统一错误响应结构。
// - code：HTTP 状态码（同时也是业务层错误码的最小实现）
// - message：可读的错误信息
type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{
		Code:    status,
		Message: message,
	})
}
