package entity

type AllocInfo struct {
	ServiceName		*string	`json:"serviceName"`
	LastAllocValue	*int64	`json:"lastAllocValue"`
	DataVersion		*int64	`json:"dataVersion"`
	// TODO add fields for loop control
}
