package threadlocal_infra

import (
	"strings"

	"github.com/google/uuid"
	"github.com/jtolds/gls"
)

var (
	Context = gls.NewContextManager()
	TraceIdKey = gls.GenSym()
)

func GetTraceId() string {
	requestIdObj, exist := Context.GetValue(TraceIdKey)
	if exist && requestIdObj != nil {
		return requestIdObj.(string)
	}
	return ""
}

func SetTraceIdWithCallBack(traceId string, cb func()) {
	if len(traceId) == 0 {
		SetRandomTraceIdWithCallBack(cb)
	}
	Context.SetValues(gls.Values{TraceIdKey: traceId}, cb)
}

func SetRandomTraceIdWithCallBack(cb func()) {
	traceId := ""
	uid, err := uuid.NewRandom()
	if err == nil {
		traceId = strings.ReplaceAll(uid.String(), "-", "")
	}
	Context.SetValues(gls.Values{TraceIdKey: traceId}, cb)
}
