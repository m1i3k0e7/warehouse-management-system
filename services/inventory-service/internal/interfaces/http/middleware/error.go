package middleware

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "warehouse/pkg/errors"
    "warehouse/pkg/logger"
)

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            switch e := err.(type) {
                case *errors.ValidationError:
                    c.JSON(http.StatusBadRequest, gin.H{"error": e.Message})
                case *errors.NotFoundError:
                    c.JSON(http.StatusNotFound, gin.H{"error": e.Message})
                case *errors.ConflictError:
                    c.JSON(http.StatusConflict, gin.H{"error": e.Message})
                case *errors.InternalError:
                    logger.Error("Internal server error", e)
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
                default:
                    logger.Error("Unhandled error", err)
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            }
        }
    }
}