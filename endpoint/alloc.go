package endpoint

import (
	"fmt"
	"strings"

	def "github.com/daemon-coder/idalloc/definition"
	"github.com/daemon-coder/idalloc/definition/dto"
	e "github.com/daemon-coder/idalloc/definition/errors"
	"github.com/daemon-coder/idalloc/service"
)

func Alloc(param dto.AllocReqDto) (result dto.AllocRespDto) {
	// param check
	param.ServiceName = strings.ToLower(strings.TrimSpace(param.ServiceName))
	if param.Count == 0 {
		param.Count = def.DEFAULT_USER_ALLOC_NUM
	}
	if len(param.ServiceName) == 0 || len(param.ServiceName) > 64 {
		e.Panic(e.NewParamError(e.WithMsg("service_name is invalid. length: 1~64")))
	} else if param.Count < 0 || param.Count > def.MAX_USER_BATCH_ALLOC_NUM {
		errMsg := fmt.Sprintf("count is invalid. min: %d max: %d input:%d", 1, def.MAX_USER_BATCH_ALLOC_NUM, param.Count)
		e.Panic(e.NewParamError(e.WithMsg(errMsg)))
	}

	result.Ids = service.DefaultAllocHandler.Alloc(param.ServiceName, param.Count)
	return
}
