/*
 * Retry Service provides a mechanism to execute operations with retry logic.
 * It allows for exponential backoff and a maximum number of retries.
 */
package services

import (
	"context"
	"fmt"
	"math"
	"time"
    "reflect"

	"WMS/services/inventory-service/pkg/utils/logger"
)

type RetryService struct {
    maxRetries int
    baseDelay  time.Duration
}

func NewRetryService(maxRetries int, baseDelay time.Duration) *RetryService {
    return &RetryService{
        maxRetries: maxRetries,
        baseDelay:  baseDelay,
    }
}

func (s *RetryService) ExecuteWithRetry(ctx context.Context, operation interface{}, args ...interface{}) error {
    op := reflect.ValueOf(operation)
	if op.Kind() != reflect.Func {
        return fmt.Errorf("operation must be a function, got %T", operation)
    }

    params := make([]reflect.Value, len(args))
	for i, arg := range args {
		params[i] = reflect.ValueOf(arg)
	}

    var lastErr error
    
    for attempt := 0; attempt <= s.maxRetries; attempt++ {
        if attempt > 0 {
            delay := time.Duration(math.Pow(2, float64(attempt-1))) * s.baseDelay
            logger.Info(fmt.Sprintf("Retrying operation, attempt %d/%d, delay: %v", attempt, s.maxRetries, delay))
            time.Sleep(delay)
        }
        
        err := op.Call(params)[0]
        if err.IsNil() {
            if attempt > 0 {
                logger.Info(fmt.Sprintf("Operation succeeded after %d attempts", attempt))
            }
            return nil
        }
        
        lastErr = err.Interface().(error)
        logger.Error(fmt.Sprintf("Operation failed, attempt %d/%d", attempt, s.maxRetries), lastErr)
    }
    
    return fmt.Errorf("operation failed after %d attempts: %w", s.maxRetries, lastErr)
}