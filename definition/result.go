package definition

import (
	"time"

	e "github.com/daemon-coder/idalloc/definition/errors"
	threadLocal "github.com/daemon-coder/idalloc/infrastructure/threadlocal_infra"
)

type Result struct {
	Code		int			`json:"code"`
	Msg 		string		`json:"msg"`
	Data		interface{} `json:"data"`
	TraceId		string		`json:"traceId"`
	Ts			int64		`json:"ts"`
}

func NewResultOK(data interface{}) Result {
	return Result{
		Code:		e.OK,
		Msg:		e.OKMsg,
		Data:		data,
		TraceId:	threadLocal.GetTraceId(),
		Ts:			time.Now().UnixMilli(),
	}
}

func NewResultFromError(err e.BaseError) Result {
	return Result{
		Code:		err.ErrorCode(),
		Msg:		err.Msg,
		Data:		err.GetData(),
		TraceId:	threadLocal.GetTraceId(),
		Ts:			time.Now().UnixMilli(),
	}
}
