package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// DatabaseHelperTestSuite tests database helper functions
type DatabaseHelperTestSuite struct {
	suite.Suite
}

func TestDatabaseHelperSuite(t *testing.T) {
	suite.Run(t, new(DatabaseHelperTestSuite))
}

// TestCalculateNextBackoffDoublesValue verifies backoff doubles
func (s *DatabaseHelperTestSuite) TestCalculateNextBackoffDoublesValue() {
	current := 1 * time.Second
	max := 30 * time.Second

	next := calculateNextBackoff(current, max)

	s.Equal(2*time.Second, next)
}

// TestCalculateNextBackoffRespectsMax verifies backoff respects max limit
func (s *DatabaseHelperTestSuite) TestCalculateNextBackoffRespectsMax() {
	current := 20 * time.Second
	max := 30 * time.Second

	next := calculateNextBackoff(current, max)

	s.Equal(30*time.Second, next)
}

// TestCalculateNextBackoffReturnsMaxWhenDoubleExceedsMax verifies max is returned
func (s *DatabaseHelperTestSuite) TestCalculateNextBackoffReturnsMaxWhenDoubleExceedsMax() {
	current := 25 * time.Second
	max := 30 * time.Second

	next := calculateNextBackoff(current, max)

	s.Equal(30*time.Second, next)
}

// TestCalculateNextBackoffWithSmallValues verifies small values work
func (s *DatabaseHelperTestSuite) TestCalculateNextBackoffWithSmallValues() {
	current := 100 * time.Millisecond
	max := 5 * time.Second

	next := calculateNextBackoff(current, max)

	s.Equal(200*time.Millisecond, next)
}

// TestCalculateNextBackoffChainingMultipleTimes verifies exponential growth
func (s *DatabaseHelperTestSuite) TestCalculateNextBackoffChainingMultipleTimes() {
	max := 30 * time.Second
	current := 1 * time.Second

	// First double: 1s -> 2s
	current = calculateNextBackoff(current, max)
	s.Equal(2*time.Second, current)

	// Second double: 2s -> 4s
	current = calculateNextBackoff(current, max)
	s.Equal(4*time.Second, current)

	// Third double: 4s -> 8s
	current = calculateNextBackoff(current, max)
	s.Equal(8*time.Second, current)

	// Fourth double: 8s -> 16s
	current = calculateNextBackoff(current, max)
	s.Equal(16*time.Second, current)

	// Fifth double: 16s -> 30s (capped at max)
	current = calculateNextBackoff(current, max)
	s.Equal(30*time.Second, current)
}
