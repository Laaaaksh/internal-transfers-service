package idempotency_test

import (
	"context"
	"testing"
	"time"

	"github.com/internal-transfers-service/internal/modules/idempotency"
	"github.com/internal-transfers-service/internal/modules/idempotency/entities"
	"github.com/internal-transfers-service/internal/modules/idempotency/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// RepositoryTestSuite contains tests for the idempotency repository
type RepositoryTestSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	mockRepo *mock.MockIRepository
	ctx      context.Context
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockRepo = mock.NewMockIRepository(s.ctrl)
	s.ctx = context.Background()
}

func (s *RepositoryTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RepositoryTestSuite) TestGetReturnsRecordWhenKeyExists() {
	expectedRecord := &entities.IdempotencyRecord{
		Key:            "test-key-123",
		ResponseStatus: 201,
		ResponseBody:   []byte(`{"transaction_id":"abc-123"}`),
		CreatedAt:      time.Now(),
	}

	s.mockRepo.EXPECT().
		Get(s.ctx, "test-key-123").
		Return(expectedRecord, nil).
		Times(1)

	record, err := s.mockRepo.Get(s.ctx, "test-key-123")

	s.NoError(err)
	s.NotNil(record)
	s.Equal("test-key-123", record.Key)
	s.Equal(201, record.ResponseStatus)
}

func (s *RepositoryTestSuite) TestGetReturnsNilWhenKeyNotFound() {
	s.mockRepo.EXPECT().
		Get(s.ctx, "nonexistent-key").
		Return(nil, nil).
		Times(1)

	record, err := s.mockRepo.Get(s.ctx, "nonexistent-key")

	s.NoError(err)
	s.Nil(record)
}

func (s *RepositoryTestSuite) TestStoreSuccessfullyStoresRecord() {
	key := "new-key-456"
	status := 201
	body := []byte(`{"transaction_id":"xyz-789"}`)

	s.mockRepo.EXPECT().
		Store(s.ctx, key, status, body).
		Return(nil).
		Times(1)

	err := s.mockRepo.Store(s.ctx, key, status, body)

	s.NoError(err)
}

func (s *RepositoryTestSuite) TestDeleteExpiredDeletesOldRecords() {
	ttl := 24 * time.Hour
	expectedDeleted := int64(5)

	s.mockRepo.EXPECT().
		DeleteExpired(s.ctx, ttl).
		Return(expectedDeleted, nil).
		Times(1)

	deleted, err := s.mockRepo.DeleteExpired(s.ctx, ttl)

	s.NoError(err)
	s.Equal(expectedDeleted, deleted)
}

// ModuleTestSuite contains tests for the idempotency module initialization
type ModuleTestSuite struct {
	suite.Suite
	ctx context.Context
}

func TestModuleSuite(t *testing.T) {
	suite.Run(t, new(ModuleTestSuite))
}

func (s *ModuleTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ModuleTestSuite) TestModuleGetRepositoryReturnsRepository() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	mockRepo := mock.NewMockIRepository(ctrl)
	module := &idempotency.Module{
		Repo: mockRepo,
	}

	repo := module.GetRepository()

	s.Equal(mockRepo, repo)
}

func (s *ModuleTestSuite) TestStopCleanupWorkerDoesNotPanicWhenNotStarted() {
	module := &idempotency.Module{}

	// Should not panic
	s.NotPanics(func() {
		module.StopCleanupWorker()
	})
}
