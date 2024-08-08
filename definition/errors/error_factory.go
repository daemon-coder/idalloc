package errors

func NewParamError(opts ...BaseErrorOpt) BaseError {
	result := New(ParamErrorType, WithMsg(ParamErrorInfo))
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func NewAuthError(opts ...BaseErrorOpt) BaseError {
	result := New(AuthErrorType, WithMsg(AuthErrorInfo))
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func NewForbiddenError(opts ...BaseErrorOpt) BaseError {
	result := New(ForbiddenErrorType, WithMsg(ForbiddenErrorInfo))
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func NewNotFoundError(opts ...BaseErrorOpt) BaseError {
	result := New(NotFoundErrorType, WithMsg(NotFoundErrorInfo))
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func NewRateLimitError(opts ...BaseErrorOpt) BaseError {
	result := New(RateLimitErrorType, WithMsg(RateLimitErrorInfo))
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func NewBusinessError(opts ...BaseErrorOpt) BaseError {
	result := New(BusinessErrorType, WithMsg(BusinessErrorInfo))
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func NewServerError(opts ...BaseErrorOpt) BaseError {
	result := New(ServerErrorType, WithMsg(ServerErrorInfo))
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func NewCriticalError(opts ...BaseErrorOpt) BaseError {
	result := New(CriticalErrorType, WithMsg(CriticalErrorInfo))
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func New(errType int, opts ...BaseErrorOpt) BaseError {
	result := BaseError{Type: errType}
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return result
}

func FromStdError(err error, opts ...BaseErrorOpt) (result BaseError) {
	if err == nil {
		return New(OK)
	}

	if v, ok := err.(BaseError); ok {
		result = v
	} else {
		result = NewServerError(WithMsg(err.Error()), WithData(err))
	}
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(&result)
		}
	}
	return
}
