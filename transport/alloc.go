package transport

import (
	"encoding/json"
	"io"

	"github.com/daemon-coder/idalloc/definition"
	"github.com/daemon-coder/idalloc/definition/dto"
	e "github.com/daemon-coder/idalloc/definition/errors"
	"github.com/daemon-coder/idalloc/endpoint"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	"github.com/kataras/iris/v12/context"
)

func Alloc(ctx *context.Context) definition.Result {
	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		log.GetLogger().Warn("ParamError", "error", err)
		e.Panic(e.NewParamError())
	}

	var reqDto dto.AllocReqDto
	err = json.Unmarshal(body, &reqDto)
	if err != nil {
		log.GetLogger().Warn("ParamError", "error", err)
		e.Panic(e.NewParamError())
	}

	respDto := endpoint.Alloc(reqDto)
	log.GetLogger().Debugw("Alloc", "request", reqDto, "response", respDto)
	return definition.NewResultOK(respDto)
}

// TODO grpc and others
