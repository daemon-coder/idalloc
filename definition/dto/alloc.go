package dto

type AllocReqDto struct {
	ServiceName string `json:"serviceName"`
	Count       int64  `json:"count"`
}

type AllocRespDto struct {
	Ids []int64 `json:"ids"`
}
