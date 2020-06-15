package utils

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/outbrain/golib/log"
)

type SearchContext struct {
	Host     string
	Port     int
	User     string
	Password string
	GtidSearch string
}

type Searchor struct {
	searchContext *SearchContext

}

type BinaryLogs struct {
	BinlogName string `db:"Log_name"`
	BinlogSize int32 `db:"File_size"`
}
var BinaryLogsList []*BinaryLogs

type BinlogEvents struct {
	BinlogName string `db:"Log_name"`
	Position int32 `db:"Pos"`
	EventType string `db:"Event_type"`
	ServerId int32 `db:"Server_id"`
	EndPosition int32 `db:"End_log_pos"`
	EventInfo string `db:"Info"`
}
var BinlogEventsList []*BinlogEvents

func NewSearchContext() *SearchContext {
	return &SearchContext{
	}
}

func NewSearchor(context *SearchContext) *Searchor {
	searchor := &Searchor{
		searchContext: context,
	}
	return searchor
}

func (this *Searchor) SearchGtid() (err error) {
	BinaryLogsList = make([]*BinaryLogs, 0)
	BinlogEventsList = make([]*BinlogEvents, 0)
	db, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema?charset=utf8&timeout=5s",
		this.searchContext.User,
		this.searchContext.Password,
		this.searchContext.Host,
		this.searchContext.Port))
	if err != nil {
		log.Errorf("Connect DB error: %s", err.Error())
		return err
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Error("DB ping error")
		return err
	}
	Query := fmt.Sprintf("SHOW BINARY LOGS;")
	err = db.Select(&BinaryLogsList, Query)
	if err != nil {
		log.Errorf("show binary logs error: %s", err.Error())
		return err
	}
	for _, v := range BinaryLogsList {
		EventsQuery := fmt.Sprintf("SHOW BINLOG EVENTS IN '%s' LIMIT 2;", v.BinlogName)
		err = db.Select(&BinlogEventsList, EventsQuery)
		if err != nil{
			log.Errorf("show binlog events error: %s", err.Error())
			return err
		}
	}
	return nil
}
