package apperror

import (
	"errors"
	"net/http"
	"testing"

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

// Test New - HTTP Status mappings (individual tests instead of table-driven)

func (s *ErrorTestSuite) TestNewWithBadRequestReturnsCorrectHTTPStatus() {
	err := New(CodeBadRequest, errors.New("test"))
	s.Equal(http.StatusBadRequest, err.HTTPStatus())
}

func (s *ErrorTestSuite) TestNewWithNotFoundReturnsCorrectHTTPStatus() {
	err := New(CodeNotFound, errors.New("test"))
	s.Equal(http.StatusNotFound, err.HTTPStatus())
}

func (s *ErrorTestSuite) TestNewWithConflictReturnsCorrectHTTPStatus() {
	err := New(CodeConflict, errors.New("test"))
	s.Equal(http.StatusConflict, err.HTTPStatus())
}

func (s *ErrorTestSuite) TestNewWithInsufficientFundsReturnsCorrectHTTPStatus() {
	err := New(CodeInsufficientFunds, errors.New("test"))
	s.Equal(http.StatusUnprocessableEntity, err.HTTPStatus())
}

func (s *ErrorTestSuite) TestNewWithInternalErrorReturnsCorrectHTTPStatus() {
	err := New(CodeInternalError, errors.New("test"))
	s.Equal(http.StatusInternalServerError, err.HTTPStatus())
}

func (s *ErrorTestSuite) TestNewWithServiceUnavailableReturnsCorrectHTTPStatus() {
	err := New(CodeServiceUnavailable, errors.New("test"))
	s.Equal(http.StatusServiceUnavailable, err.HTTPStatus())
}

func (s *ErrorTestSuite) TestNewWithValidationErrorReturnsCorrectHTTPStatus() {
	err := New(CodeValidationError, errors.New("test"))
	s.Equal(http.StatusBadRequest, err.HTTPStatus())
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

// Test Code String (individual tests instead of table-driven)

func (s *ErrorTestSuite) TestCodeBadRequestStringReturnsCorrectValue() {
	s.Equal("BAD_REQUEST", CodeBadRequest.String())
}

func (s *ErrorTestSuite) TestCodeNotFoundStringReturnsCorrectValue() {
	s.Equal("NOT_FOUND", CodeNotFound.String())
}

func (s *ErrorTestSuite) TestCodeConflictStringReturnsCorrectValue() {
	s.Equal("CONFLICT", CodeConflict.String())
}

func (s *ErrorTestSuite) TestCodeInternalErrorStringReturnsCorrectValue() {
	s.Equal("INTERNAL_ERROR", CodeInternalError.String())
}

func (s *ErrorTestSuite) TestCodeInsufficientFundsStringReturnsCorrectValue() {
	s.Equal("INSUFFICIENT_FUNDS", CodeInsufficientFunds.String())
}

func (s *ErrorTestSuite) TestCodeServiceUnavailableStringReturnsCorrectValue() {
	s.Equal("SERVICE_UNAVAILABLE", CodeServiceUnavailable.String())
}

func (s *ErrorTestSuite) TestCodeValidationErrorStringReturnsCorrectValue() {
	s.Equal("VALIDATION_ERROR", CodeValidationError.String())
}

// Test Code HTTPStatus (individual tests instead of table-driven)

func (s *ErrorTestSuite) TestCodeBadRequestHTTPStatusReturnsCorrectValue() {
	s.Equal(http.StatusBadRequest, CodeBadRequest.HTTPStatus())
}

func (s *ErrorTestSuite) TestCodeValidationErrorHTTPStatusReturnsCorrectValue() {
	s.Equal(http.StatusBadRequest, CodeValidationError.HTTPStatus())
}

func (s *ErrorTestSuite) TestCodeNotFoundHTTPStatusReturnsCorrectValue() {
	s.Equal(http.StatusNotFound, CodeNotFound.HTTPStatus())
}

func (s *ErrorTestSuite) TestCodeConflictHTTPStatusReturnsCorrectValue() {
	s.Equal(http.StatusConflict, CodeConflict.HTTPStatus())
}

func (s *ErrorTestSuite) TestCodeInsufficientFundsHTTPStatusReturnsCorrectValue() {
	s.Equal(http.StatusUnprocessableEntity, CodeInsufficientFunds.HTTPStatus())
}

func (s *ErrorTestSuite) TestCodeInternalErrorHTTPStatusReturnsCorrectValue() {
	s.Equal(http.StatusInternalServerError, CodeInternalError.HTTPStatus())
}

func (s *ErrorTestSuite) TestCodeServiceUnavailableHTTPStatusReturnsCorrectValue() {
	s.Equal(http.StatusServiceUnavailable, CodeServiceUnavailable.HTTPStatus())
}

// Test Is and As functions

func (s *ErrorTestSuite) TestIsWithMatchingErrorReturnsTrue() {
	err1 := errors.New("test error")

	s.True(Is(err1, err1))
}

func (s *ErrorTestSuite) TestIsWithDifferentErrorsReturnsFalse() {
	err1 := errors.New("test error")
	err2 := errors.New("test error")

	s.False(Is(err1, err2))
}

func (s *ErrorTestSuite) TestAsWithMatchingTypeSucceeds() {
	appErr := New(CodeBadRequest, errors.New("test"))

	var target *Error
	result := As(appErr, &target)

	s.True(result)
	s.NotNil(target)
	s.Equal(CodeBadRequest, target.Code())
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
