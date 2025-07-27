/*
    * Retry Service provides a mechanism to execute operations with retry logic.
    * It allows for exponential backoff and a maximum number of retries.
*/
package services

import (
    "fmt"
    "math"
    "time"
    
    "warehouse/pkg/logger"
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

func (s *RetryService) ExecuteWithRetry(operation func() error, maxRetries int, baseDelay time.Duration) error {
    var lastErr error
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
            logger.Info(fmt.Sprintf("Retrying operation, attempt %d/%d, delay: %v", attempt, maxRetries, delay))
            time.Sleep(delay)
        }
        
        err := operation()
        if err == nil {
            if attempt > 0 {
                logger.Info(fmt.Sprintf("Operation succeeded after %d attempts", attempt))
            }
            return nil
        }
        
        lastErr = err
        logger.Error(fmt.Sprintf("Operation failed, attempt %d/%d", attempt, maxRetries), err)
    }
    
    return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, lastErr)
}

func (s *RetryService) Execute(operation func() error) error {
    return s.ExecuteWithRetry(operation, s.maxRetries, s.baseDelay)
}