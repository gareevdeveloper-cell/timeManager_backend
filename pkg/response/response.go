package response

import (
	"github.com/gin-gonic/gin"
)

// Error отправляет JSON-ответ с ошибкой в стандартном формате API.
// Формат: { "error": { "code": "...", "message": "..." } }
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

// Data отправляет JSON-ответ с данными в стандартном формате API.
// Формат: { "data": ... }
func Data(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{"data": data})
}
