package utils

import "github.com/gin-gonic/gin"

func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
	c.Abort()
}