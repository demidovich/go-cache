package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	ctx := context.Background()
	c := NewCacheWithConfig(
		ctx,
		Config{
			Size:       10,
			GcInterval: 10 * time.Second,
		},
	)

	assert.Equal(t, c.size, 10)
	assert.Equal(t, c.gcInterval, 10*time.Second)
}
