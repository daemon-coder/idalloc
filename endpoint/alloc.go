package endpoint

import (
	"fmt"
	"strings"

	"github.com/daemon-coder/idalloc/definition/dto"
	e "github.com/daemon-coder/idalloc/definition/errors"
	"github.com/daemon-coder/idalloc/service"
)

const (
	MAX_ALLOC_COUNT = 100
	DEFAULT_ALLOC_COUNT = 1
)

func Alloc(param dto.AllocReqDto) (result dto.AllocRespDto) {
	// param check
	param.ServiceName = strings.ToLower(strings.TrimSpace(param.ServiceName))
	if param.Count == 0 {
		param.Count = DEFAULT_ALLOC_COUNT
	}
	if len(param.ServiceName) == 0 || len(param.ServiceName) > 64 {
		e.Panic(e.NewParamError(e.WithMsg("service_name is invalid. length: 1~64")))
	} else if param.Count < 0 || param.Count > MAX_ALLOC_COUNT {
		e.Panic(e.NewParamError(e.WithMsg(fmt.Sprintf("count is invalid. min: %d max: %d input:%d", 1, MAX_ALLOC_COUNT, param.Count))))
	}

	result.Ids = service.DefaultAllocHandler.Alloc(param.ServiceName, param.Count)
	return
}
