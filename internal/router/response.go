package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
type Response struct {
	Code   int         `json:"code"`
	ErrMsg string      `json:"errMsg,omitempty"`
	Body   interface{} `json:"body,omitempty"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Body: data})
}

// Created 返回创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 0, Body: data})
}

// Error 返回错误响应
func Error(c *gin.Context, status int, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	c.JSON(status, Response{Code: status, ErrMsg: errMsg})
}