package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ErrorTestSuite contains tests for apperror
type ErrorTestSuite struct {
	suite.Suite
}

func TestErrorSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}

// Test New constructor

func (s *ErrorTestSuite) TestNewCreatesErrorWithCorrectCode() {
	err := New(CodeBadRequest, errors.New("test error"))
	s.Equal(CodeBadRequest, err.Code())
}

func (s *ErrorTestSuite) TestNewCreatesErrorWithCorrectHTTPStatus() {
	testCases := []struct {
		name           string
		code           Code
		expectedStatus int
	}{
		{name: "badRequest", code: CodeBadRequest, expectedStatus: http.StatusBadRequest},
		{name: "notFound", code: CodeNotFound, expectedStatus: http.StatusNotFound},
		{name: "conflict", code: CodeConflict, expectedStatus: http.StatusConflict},
		{name: "insufficientFunds", code: CodeInsufficientFunds, expectedStatus: http.StatusUnprocessableEntity},
		{name: "internalError", code: CodeInternalError, expectedStatus: http.StatusInternalServerError},
		{name: "serviceUnavailable", code: CodeServiceUnavailable, expectedStatus: http.StatusServiceUnavailable},
		{name: "validationError", code: CodeValidationError, expectedStatus: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := New(tc.code, errors.New("test"))
			s.Equal(tc.expectedStatus, err.HTTPStatus())
		})
	}
}

func (s *ErrorTestSuite) TestNewCreatesErrorWithDefaultPublicMessage() {
	err := New(CodeBadRequest, errors.New("test"))
	s.NotEmpty(err.PublicMessage())
}

// Test NewWithMessage

func (s *ErrorTestSuite) TestNewWithMessageSetsCustomPublicMessage() {
	customMessage := "Custom error message"
	err := NewWithMessage(CodeBadRequest, errors.New("test"), customMessage)
	s.Equal(customMessage, err.PublicMessage())
}

// Test Error method

func (s *ErrorTestSuite) TestErrorReturnsFormattedString() {
	cause := errors.New("underlying error")
	err := New(CodeBadRequest, cause)
	errorString := err.Error()
	
	s.Contains(errorString, string(CodeBadRequest))
	s.Contains(errorString, "underlying error")
}

func (s *ErrorTestSuite) TestErrorWithNilCauseReturnsMessageOnly() {
	err := New(CodeBadRequest, nil)
	errorString := err.Error()
	
	s.Contains(errorString, string(CodeBadRequest))
	s.NotContains(errorString, "<nil>")
}

// Test WithField

func (s *ErrorTestSuite) TestWithFieldAddsFieldToError() {
	err := New(CodeBadRequest, errors.New("test"))
	err = err.WithField("account_id", int64(123)).(*Error)
	
	fields := err.Fields()
	s.Contains(fields, "account_id")
	s.Equal(int64(123), fields["account_id"])
}

func (s *ErrorTestSuite) TestWithFieldChainsMultipleFields() {
	err := New(CodeBadRequest, errors.New("test")).
		WithField("field1", "value1").
		WithField("field2", "value2").(*Error)
	
	fields := err.Fields()
	s.Equal("value1", fields["field1"])
	s.Equal("value2", fields["field2"])
}

// Test WithFields

func (s *ErrorTestSuite) TestWithFieldsAddsMultipleFieldsAtOnce() {
	err := New(CodeBadRequest, errors.New("test"))
	err = err.WithFields(map[string]interface{}{
		"field1": "value1",
		"field2": 123,
	}).(*Error)
	
	fields := err.Fields()
	s.Equal("value1", fields["field1"])
	s.Equal(123, fields["field2"])
}

// Test Unwrap

func (s *ErrorTestSuite) TestUnwrapReturnsCause() {
	cause := errors.New("underlying error")
	err := New(CodeBadRequest, cause)
	
	s.Equal(cause, err.Unwrap())
}

func (s *ErrorTestSuite) TestUnwrapWithNilCauseReturnsNil() {
	err := New(CodeBadRequest, nil)
	s.Nil(err.Unwrap())
}

// Test Code String

func TestCodeStringReturnsCorrectValue(t *testing.T) {
	testCases := []struct {
		code     Code
		expected string
	}{
		{CodeBadRequest, "BAD_REQUEST"},
		{CodeNotFound, "NOT_FOUND"},
		{CodeConflict, "CONFLICT"},
		{CodeInternalError, "INTERNAL_ERROR"},
		{CodeInsufficientFunds, "INSUFFICIENT_FUNDS"},
		{CodeServiceUnavailable, "SERVICE_UNAVAILABLE"},
		{CodeValidationError, "VALIDATION_ERROR"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.code.String())
		})
	}
}

// Test HTTPStatus mappings

func TestCodeHTTPStatusReturnsCorrectStatusCode(t *testing.T) {
	testCases := []struct {
		code     Code
		expected int
	}{
		{CodeBadRequest, http.StatusBadRequest},
		{CodeValidationError, http.StatusBadRequest},
		{CodeNotFound, http.StatusNotFound},
		{CodeConflict, http.StatusConflict},
		{CodeInsufficientFunds, http.StatusUnprocessableEntity},
		{CodeInternalError, http.StatusInternalServerError},
		{CodeServiceUnavailable, http.StatusServiceUnavailable},
	}

	for _, tc := range testCases {
		t.Run(tc.code.String(), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.code.HTTPStatus())
		})
	}
}

// Test Is and As functions

func TestIsWithMatchingErrorReturnsTrue(t *testing.T) {
	err1 := errors.New("test error")
	err2 := errors.New("test error")
	
	assert.True(t, Is(err1, err1))
	assert.False(t, Is(err1, err2))
}

func TestAsWithMatchingTypeSucceeds(t *testing.T) {
	appErr := New(CodeBadRequest, errors.New("test"))
	
	var target *Error
	result := As(appErr, &target)
	
	assert.True(t, result)
	assert.NotNil(t, target)
	assert.Equal(t, CodeBadRequest, target.Code())
}

// Test Fields initialization

func (s *ErrorTestSuite) TestFieldsInitializedAsEmptyMap() {
	err := New(CodeBadRequest, errors.New("test"))
	s.NotNil(err.Fields())
	s.Len(err.Fields(), 0)
}

func (s *ErrorTestSuite) TestWithFieldInitializesFieldsIfNil() {
	err := &Error{code: CodeBadRequest}
	err = err.WithField("key", "value").(*Error)
	
	s.NotNil(err.Fields())
	s.Contains(err.Fields(), "key")
}

func (s *ErrorTestSuite) TestWithFieldsInitializesFieldsIfNil() {
	err := &Error{code: CodeBadRequest}
	err = err.WithFields(map[string]interface{}{"key": "value"}).(*Error)
	
	s.NotNil(err.Fields())
	s.Contains(err.Fields(), "key")
}
