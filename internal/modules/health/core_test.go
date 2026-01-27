package health_test

import (
	"context"
	"testing"

	"github.com/internal-transfers-service/internal/modules/health"
	"github.com/stretchr/testify/suite"
)

// CoreTestSuite contains tests for health Core
type CoreTestSuite struct {
	suite.Suite
	ctx context.Context
}

func TestCoreSuite(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

func (s *CoreTestSuite) SetupTest() {
	s.ctx = context.Background()
}

// Test RunLivenessCheck - Success Cases

func (s *CoreTestSuite) TestLivenessCheckWhenHealthyReturnsServing() {
	core := health.NewCoreForTesting(nil, true)

	status, code := core.RunLivenessCheck(s.ctx)
	s.Equal("SERVING", status)
	s.Equal(200, code)
}

func (s *CoreTestSuite) TestLivenessCheckWhenUnhealthyReturnsNotServing() {
	core := health.NewCoreForTesting(nil, false)

	status, code := core.RunLivenessCheck(s.ctx)
	s.Equal("NOT_SERVING", status)
	s.Equal(503, code)
}

// Test RunReadinessCheck - Success Cases

func (s *CoreTestSuite) TestReadinessCheckWhenHealthyAndNilDbReturnsServing() {
	core := health.NewCoreForTesting(nil, true)

	status, code := core.RunReadinessCheck(s.ctx)
	s.Equal("SERVING", status)
	s.Equal(200, code)
}

func (s *CoreTestSuite) TestReadinessCheckWhenUnhealthyReturnsNotServing() {
	core := health.NewCoreForTesting(nil, false)

	status, code := core.RunReadinessCheck(s.ctx)
	s.Equal("NOT_SERVING", status)
	s.Equal(503, code)
}

// Test MarkUnhealthy

func (s *CoreTestSuite) TestMarkUnhealthySetsServiceAsUnhealthy() {
	core := health.NewCoreForTesting(nil, true)

	s.True(core.IsHealthy())
	core.MarkUnhealthy()
	s.False(core.IsHealthy())
}

// Test IsHealthy

func (s *CoreTestSuite) TestIsHealthyReturnsTrueWhenHealthy() {
	core := health.NewCoreForTesting(nil, true)
	s.True(core.IsHealthy())
}

func (s *CoreTestSuite) TestIsHealthyReturnsFalseWhenUnhealthy() {
	core := health.NewCoreForTesting(nil, false)
	s.False(core.IsHealthy())
}
