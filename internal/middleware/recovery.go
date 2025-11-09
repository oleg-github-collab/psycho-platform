package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				stack := debug.Stack()
				log.Printf("PANIC: %v\n%s", err, stack)

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": fmt.Sprintf("%v", err),
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
