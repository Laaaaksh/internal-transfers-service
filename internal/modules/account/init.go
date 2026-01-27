package account

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Module singleton instance
var AccModule IModule

// NewModule initializes the account module
var NewModule = func(ctx context.Context, pool *pgxpool.Pool) IModule {
	if AccModule == nil {
		repo := NewRepository(pool)
		core := NewCore(ctx, repo)
		handler := NewHTTPHandler(core)

		AccModule = &Module{
			Core:    core,
			Handler: handler,
			Repo:    repo,
		}
	}
	return AccModule
}

// IModule defines the interface for the account module
type IModule interface {
	GetCore() ICore
	GetHandler() *HTTPHandler
	GetRepository() IRepository
}

// Module implements IModule
type Module struct {
	Core    ICore
	Handler *HTTPHandler
	Repo    IRepository
}

// GetCore returns the core business logic
func (m *Module) GetCore() ICore {
	return m.Core
}

// GetHandler returns the HTTP handler
func (m *Module) GetHandler() *HTTPHandler {
	return m.Handler
}

// GetRepository returns the repository
func (m *Module) GetRepository() IRepository {
	return m.Repo
}
