package health

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLivenessCheckWhenHealthyReturnsServing tests liveness when healthy
func TestLivenessCheckWhenHealthyReturnsServing(t *testing.T) {
	core := &Core{
		db:      nil, // nil db for liveness test
		healthy: 1,
	}
	ctx := context.Background()

	status, code := core.RunLivenessCheck(ctx)
	assert.Equal(t, "SERVING", status)
	assert.Equal(t, 200, code)
}

// TestLivenessCheckWhenUnhealthyReturnsNotServing tests liveness when unhealthy
func TestLivenessCheckWhenUnhealthyReturnsNotServing(t *testing.T) {
	core := &Core{
		db:      nil,
		healthy: 0,
	}
	ctx := context.Background()

	status, code := core.RunLivenessCheck(ctx)
	assert.Equal(t, "NOT_SERVING", status)
	assert.Equal(t, 503, code)
}

// TestReadinessCheckWhenUnhealthyReturnsNotServing tests readiness when service is marked unhealthy
func TestReadinessCheckWhenUnhealthyReturnsNotServing(t *testing.T) {
	core := &Core{
		db:      nil,
		healthy: 0,
	}
	ctx := context.Background()

	status, code := core.RunReadinessCheck(ctx)
	assert.Equal(t, "NOT_SERVING", status)
	assert.Equal(t, 503, code)
}

// TestMarkUnhealthySetsServiceAsUnhealthy tests the MarkUnhealthy method
func TestMarkUnhealthySetsServiceAsUnhealthy(t *testing.T) {
	core := &Core{
		db:      nil,
		healthy: 1,
	}

	assert.True(t, core.IsHealthy())
	core.MarkUnhealthy()
	assert.False(t, core.IsHealthy())
}

// TestIsHealthyReturnsCorrectStatus tests the IsHealthy method
func TestIsHealthyReturnsCorrectStatus(t *testing.T) {
	healthyCore := &Core{
		db:      nil,
		healthy: 1,
	}
	assert.True(t, healthyCore.IsHealthy())

	unhealthyCore := &Core{
		db:      nil,
		healthy: 0,
	}
	assert.False(t, unhealthyCore.IsHealthy())
}

// TestReadinessCheckWithNilDbReturnsServing tests readiness when db is nil but service is healthy
func TestReadinessCheckWithNilDbReturnsServing(t *testing.T) {
	core := &Core{
		db:      nil,
		healthy: 1,
	}
	ctx := context.Background()

	status, code := core.RunReadinessCheck(ctx)
	assert.Equal(t, "SERVING", status)
	assert.Equal(t, 200, code)
}
