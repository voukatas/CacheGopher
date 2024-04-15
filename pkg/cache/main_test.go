package cache

import (
	"os"
	"testing"

	"github.com/voukatas/CacheGopher/pkg/logger"
)

func TestMain(m *testing.M) {
	// Setup
	code := m.Run()
	// Teardown
	os.Exit(code)
}

func NewTestCache() *Cache {

	tlogger := logger.SetupDebugLogger()
	return &Cache{
		store:  make(map[string]string, 0),
		logger: tlogger,
		size:   0,
	}
}
