package transaction

import (
	"context"

	"github.com/internal-transfers-service/internal/modules/account"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module singleton instance
var TxModule IModule

// NewModule initializes the transaction module
var NewModule = func(ctx context.Context, pool *pgxpool.Pool, accountRepo account.IRepository) IModule {
	if TxModule == nil {
		repo := NewRepository(pool)
		core := NewCore(ctx, repo, accountRepo)
		handler := NewHTTPHandler(core)

		TxModule = &Module{
			Core:    core,
			Handler: handler,
			Repo:    repo,
		}
	}
	return TxModule
}

// IModule defines the interface for the transaction module
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
