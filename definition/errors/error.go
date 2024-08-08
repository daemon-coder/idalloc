package errors

import (
	"encoding/json"
	"fmt"
)

// Error Scope
const (
	ScopeCommon = 0
	ScoreServer = 101
	ScoreClient = 102
)

// Error Types
const (
	OK    = 0
	OKMsg = "OK"

	ParamErrorType = 400
	ParamErrorInfo = "Param error"

	AuthErrorType = 401
	AuthErrorInfo = "Not authorized"

	ForbiddenErrorType = 403
	ForbiddenErrorInfo = "Forbidden"

	NotFoundErrorType = 404
	NotFoundErrorInfo = "Not found"

	RateLimitErrorType = 429
	RateLimitErrorInfo = "Rate limit"

	// BusinessError: The code logic is correct, but errors occur due to business logic limitations.
	// These errors typically do not require attention.
	BusinessErrorType = 500
	BusinessErrorInfo = "Business error"

	// ServerError: These are exceptions that do not require manual handling for each instance.
	// Typically, they only need attention when there are a large number of errors.
	// The code logic is correct, and these exceptions are caused by issues like network errors or environmental problems.
	ServerErrorType = 501
	ServerErrorInfo = "Server error"

	// CriticalError: These are the highest level of errors, requiring manual intervention for each occurrence.
	// They are typically due to code bugs or serious business errors that necessitate human attention.
	CriticalErrorType = 502
	CriticalErrorInfo = "Critical error"
)

type BaseError struct {
	Scope int         `json:"scope"`
	Type  int         `json:"type"`
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func (e BaseError) ErrorCode() int {
	return e.Scope*1e6 + e.Type*1e3 + e.Code
}

func (e BaseError) Error() string {
	jsonStr, _ := json.Marshal(e)
	return fmt.Sprintf("[%d]%s detail:%s", e.ErrorCode(), e.Msg, jsonStr)
}

func (e BaseError) GetData() interface{} {
	return e.Data
}

func (e BaseError) Panic() {
	panic(e)
}


type BaseErrorOpt func(*BaseError)

func WithScope(scope int) BaseErrorOpt {
	return func(e *BaseError) {
		e.Scope = scope
	}
}

func WithType(errType int) BaseErrorOpt {
	return func(e *BaseError) {
		e.Type = errType
	}
}

func WithCode(code int) BaseErrorOpt {
	return func(e *BaseError) {
		e.Code = code
	}
}

func WithMsg(msg string) BaseErrorOpt {
	return func(e *BaseError) {
		e.Msg = msg
	}
}

func WithData(data interface{}) BaseErrorOpt {
	return func(e *BaseError) {
		e.Data = data
	}
}
