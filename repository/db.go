package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/daemon-coder/idalloc/definition/entity"
	e "github.com/daemon-coder/idalloc/definition/errors"
	db "github.com/daemon-coder/idalloc/infrastructure/db_infra"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
	"github.com/daemon-coder/idalloc/util"
)

func GetAllocInfoFromDB(serviceNames ...string) (result []*entity.AllocInfo) {
	result = make([]*entity.AllocInfo, 0, len(serviceNames))
	query := db.SqlUtil{
		Sql: fmt.Sprintf(
			"select service_name, last_alloc_value, data_version from tbl_alloc_info where service_name in (%s)",
			strings.Join(util.SliceRepeat("?", len(serviceNames)), ", "),
		),
		Args: util.ToInterfaceSlice(serviceNames),
	}
	log.GetLogger().Debugw("JIANWEI_DEBUG", "sql", query)
	query.QueryList(func(row *sql.Rows) (err error) {
		var serviceNamePtr *string
		var lastAllocValuePtr, dataVersionPtr *int64
		err = row.Scan(&serviceNamePtr, &lastAllocValuePtr, &dataVersionPtr)
		if err == nil {
			result = append(result, &entity.AllocInfo{
				ServiceName: serviceNamePtr,
				LastAllocValue: lastAllocValuePtr,
				DataVersion: dataVersionPtr,
			})
		}
		return
	})
	return
}

func GetServiceAllocInfoFromDB(serviceName string) (result *entity.AllocInfo) {
	query := db.SqlUtil{
		Sql: "select last_alloc_value, data_version from tbl_alloc_info where service_name = ?",
		Args: []interface{}{serviceName},
	}
	query.QueryOne(func(row *sql.Row) (err error) {
		var lastAllocValuePtr, dataVersionPtr *int64
		err = row.Scan(&lastAllocValuePtr, &dataVersionPtr)
		if err == nil {
			result = &entity.AllocInfo{
				ServiceName: util.Ptr(serviceName),
				LastAllocValue: lastAllocValuePtr,
				DataVersion: dataVersionPtr,
			}
		}
		return
	})
	return
}

func GetAllFromDB() (result []*entity.AllocInfo) {
	result = make([]*entity.AllocInfo, 0)
	query := db.SqlUtil{
		Sql:  "select service_name, last_alloc_value, data_version from tbl_alloc_info",
	}
	query.QueryList(func(row *sql.Rows) (err error) {
		var serviceNamePtr *string
		var lastAllocValuePtr, dataVersionPtr *int64
		err = row.Scan(&serviceNamePtr, &lastAllocValuePtr, &dataVersionPtr)
		if err == nil {
			result = append(result, &entity.AllocInfo{
				ServiceName: serviceNamePtr,
				LastAllocValue: lastAllocValuePtr,
				DataVersion: dataVersionPtr,
			})
		}
		return
	})
	return
}

func InsertAllocInfoToDB(allocInfo *entity.AllocInfo) {
	query := db.SqlUtil{
		Sql:  "insert into tbl_alloc_info(service_name, last_alloc_value, data_version) values (?, ?, ?)",
		Args: []interface{}{allocInfo.ServiceName, allocInfo.LastAllocValue, allocInfo.DataVersion},
	}
	_, _, err := query.Exec()
	if err != nil {
		log.GetLogger().Warnw("InsertAllocInfoToDB", "sql", query.Sql, "args", query.Args, "err", err)
		e.Panic(err)
	}
}

func UpdateAllocInfoToDB(allocInfo *entity.AllocInfo) {
	query := db.SqlUtil{
		Sql:  "update tbl_alloc_info set last_alloc_value = ?, data_version = ? where service_name = ? and data_version < ?",
		Args: []interface{}{
			allocInfo.LastAllocValue,
			allocInfo.DataVersion,
			allocInfo.ServiceName,
			allocInfo.DataVersion,
		},
	}
	_, _, err := query.Exec()
	if err != nil {
		log.GetLogger().Warnw("InsertAllocInfoToDB", "sql", query.Sql, "args", query.Args, "err", err)
		e.Panic(err)
	}
}

func InsertOrUpdateAllocInfoToDB(allocInfo *entity.AllocInfo) {
	lockKey := "insert_or_update_db_" + *allocInfo.ServiceName
	expire := 5 * time.Second
	WithRedisLock(lockKey, expire, func() {
		result := GetServiceAllocInfoFromDB(*allocInfo.ServiceName)
		if result == nil {
			InsertAllocInfoToDB(allocInfo)
		} else {
			UpdateAllocInfoToDB(allocInfo)
		}
	})
}
