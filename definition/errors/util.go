package errors


func Error2Panic(err error) {
	if err == nil {
		return
	}
	Panic(err)
}

func PanicIf(condition bool, err error) {
	if condition {
		Panic(err)
	}
}

func Panic(err error) {
	if v, ok := err.(BaseError); ok {
		v.Panic()
	}
	NewServerError(WithMsg(err.Error())).Panic()
}

func PanicRecover(errHandler func(err BaseError)) {
	errObj := recover()
	if errObj == nil || errHandler == nil {
		return
	}
	switch v := errObj.(type) {
	case BaseError:
		errHandler(v)
	case error:
		err := NewServerError(WithMsg(v.Error()))
		errHandler(err)
	default:
		err := NewServerError(WithData(v))
		errHandler(err)
	}
}
