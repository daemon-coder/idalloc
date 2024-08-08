package db_infra

import (
	"database/sql"

	e "github.com/daemon-coder/idalloc/definition/errors"
	log "github.com/daemon-coder/idalloc/infrastructure/log_infra"
)

// To be compatible with DM database, we did not use an ORM framework here; we will switch to GORM later.

type SqlUtil struct {
	Sql  string
	Args []interface{}
}

func (q SqlUtil) QueryOne(rowParser func(*sql.Row) error) {
	stmt := Prepare(q.Sql)
	defer stmt.Close()

	row := stmt.QueryRow(q.Args...)

	err := rowParser(row)
	if err != nil && err != sql.ErrNoRows {
		log.GetLogger().Warnw("SqlRowParseError", "sql", q.Sql, "args", q.Args, "err", err)
		e.NewCriticalError(
			e.WithMsg("SqlRowParseError"),
			e.WithData(map[string]interface{}{
				"sql": q.Sql, "args": q.Args, "err": err,
			}),
		).Panic()
	}
}

func (q SqlUtil) QueryList(rowParser func(*sql.Rows) error) {
	stmt := Prepare(q.Sql)
	defer stmt.Close()

	rows, err := stmt.Query(q.Args...)
	if err == sql.ErrNoRows {
		return
	} else if err != nil {
		log.GetLogger().Warnw("QueryDbError", "sql", q.Sql, "args", q.Args, "err", err)
		e.NewServerError(
			e.WithMsg("QueryDbError"),
			e.WithData(map[string]interface{}{
				"sql": q.Sql, "args": q.Args, "err": err,
			}),
		).Panic()
	}
	defer rows.Close()

	for rows.Next() {
		err := rowParser(rows)
		if err != nil && err != sql.ErrNoRows {
			log.GetLogger().Warnw("SqlRowParseError", "sql", q.Sql, "args", q.Args, "err", err)
			e.NewCriticalError(
				e.WithMsg("SqlRowParseError"),
				e.WithData(map[string]interface{}{
					"sql": q.Sql, "args": q.Args, "err": err,
				}),
			).Panic()
		}
	}
	if err = rows.Err(); err != nil {
		log.GetLogger().Warnw("SqlRowParseError", "sql", q.Sql, "args", q.Args, "err", err)
		e.NewCriticalError(
			e.WithMsg("SqlRowParseError"),
			e.WithData(map[string]interface{}{
				"sql": q.Sql, "args": q.Args, "err": err,
			}),
		).Panic()
	}
}

func (q SqlUtil) Exec() (rowsAffected, lastInsertId int64, err error) {
	stmt := Prepare(q.Sql)
	defer stmt.Close()

	sqlResult, err := stmt.Exec(q.Args...)
	if err != nil {
		log.GetLogger().Warnw("SqlExecError", "sql", q.Sql, "args", q.Args, "err", err)
	}
	rowsAffected, _ = sqlResult.RowsAffected()
	lastInsertId, _ = sqlResult.LastInsertId()
	return
}


func Prepare(sqlStr string) *sql.Stmt {
	stmt, err := DBClient.Prepare(sqlStr)
	if err != nil {
		log.GetLogger().Warnw("SqlError", "sql", sqlStr, "err", err)
		e.NewCriticalError(
			e.WithMsg("SqlError"),
			e.WithData(map[string]interface{}{
				"sql": sqlStr, "err": err,
			}),
		).Panic()
	}
	return stmt
}
