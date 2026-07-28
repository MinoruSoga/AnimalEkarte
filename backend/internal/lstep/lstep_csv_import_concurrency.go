package lstep

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

const maxConcurrentLstepCSVImports = 2
const maxConcurrentLstepCSVImportsPerClinic = 1

type lstepCSVImportConcurrencyLimiter struct {
	globalSlots    chan struct{}
	mu             sync.Mutex
	activeByClinic map[uint64]int
}

// newLstepCSVImportConcurrencyGate は高コストなmultipart解析より前に、
// プロセス内で同時実行するCSV import数を制限する。
func newLstepCSVImportConcurrencyGate(maxConcurrent int) gin.HandlerFunc {
	if maxConcurrent < 1 {
		panic("maxConcurrent must be positive")
	}
	limiter := &lstepCSVImportConcurrencyLimiter{
		globalSlots:    make(chan struct{}, maxConcurrent),
		activeByClinic: make(map[uint64]int),
	}
	return func(c *gin.Context) {
		clinicID, ok := httpapi.ExtractClinicID(c)
		if !ok {
			c.Abort()
			return
		}
		if !limiter.tryAcquire(clinicID) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many concurrent CSV imports",
			})
			return
		}
		defer limiter.release(clinicID)
		c.Next()
	}
}

func (l *lstepCSVImportConcurrencyLimiter) tryAcquire(clinicID uint64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.activeByClinic[clinicID] >= maxConcurrentLstepCSVImportsPerClinic {
		return false
	}
	select {
	case l.globalSlots <- struct{}{}:
		l.activeByClinic[clinicID]++
		return true
	default:
		return false
	}
}

func (l *lstepCSVImportConcurrencyLimiter) release(clinicID uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.activeByClinic[clinicID] <= 1 {
		delete(l.activeByClinic, clinicID)
	} else {
		l.activeByClinic[clinicID]--
	}
	<-l.globalSlots
}
