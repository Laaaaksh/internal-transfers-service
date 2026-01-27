package health

import (
	"context"

	"github.com/internal-transfers-service/pkg/database"
)

// HtModule is the health module singleton
var HtModule IModule

// NewModule initializes the health module
var NewModule = func(ctx context.Context, db *database.Database) IModule {
	if HtModule == nil {
		core := NewCore(db)
		HtModule = &Module{
			Core: core,
		}
	}
	return HtModule
}

// IModule defines the interface for the health module
type IModule interface {
	GetCore() ICore
}

// Module implements IModule
type Module struct {
	Core ICore
}

// GetCore returns the core business logic
func (m *Module) GetCore() ICore {
	return m.Core
}
