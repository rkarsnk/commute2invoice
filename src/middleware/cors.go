// Package middleware は、gin のミドルウェアを定義します。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rkarsnk/commute2invoice/src/models"
)

// CORSMiddleware は、クロスオリジンリクエストを許可します。
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// NoCacheMiddleware は、頻繁に更新される単一ページ画面を常に最新の状態で配信します。
func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

// ErrorHandlerMiddleware は、パニック時のエラーハンドリングを行います。
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error: "Internal Server Error",
					Code:  "INTERNAL_ERROR",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// NotFoundMiddleware は、存在しないエンドポイントへの対応。
func NotFoundMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Not Found",
			Code:  "NOT_FOUND",
		})
	}
}
