package health_test

import (
	"context"
	"testing"

	"github.com/internal-transfers-service/internal/modules/health"
	"github.com/stretchr/testify/suite"
)

// InitTestSuite contains tests for health module initialization
type InitTestSuite struct {
	suite.Suite
	ctx context.Context
}

func TestInitSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

func (s *InitTestSuite) SetupTest() {
	s.ctx = context.Background()
}

// TestNewModuleCreatesModule verifies module creation
func (s *InitTestSuite) TestNewModuleCreatesModule() {
	module := health.NewModule(s.ctx, nil)

	s.NotNil(module)
}

// TestNewModuleReturnsSingleton verifies singleton pattern
func (s *InitTestSuite) TestNewModuleReturnsSingleton() {
	module1 := health.NewModule(s.ctx, nil)
	module2 := health.NewModule(s.ctx, nil)

	s.Equal(module1, module2)
}

// TestModuleGetCoreReturnsCore verifies GetCore returns valid core
func (s *InitTestSuite) TestModuleGetCoreReturnsCore() {
	module := health.NewModule(s.ctx, nil)

	core := module.GetCore()

	s.NotNil(core)
}

// TestModuleGetCoreReturnsHealthyCore verifies core starts healthy
func (s *InitTestSuite) TestModuleGetCoreReturnsHealthyCore() {
	module := health.NewModule(s.ctx, nil)

	core := module.GetCore()

	s.True(core.IsHealthy())
}
