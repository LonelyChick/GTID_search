package utils

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/outbrain/golib/log"
	"strings"
)

//命令行参数
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

//binlog列表
type BinaryLogs struct {
	BinlogName string `db:"Log_name"`
	BinlogSize int64 `db:"File_size"`
}
var BinaryLogsList []*BinaryLogs

//binlog events
type BinlogEvents struct {
	BinlogName string `db:"Log_name"`
	Position int64 `db:"Pos"`
	EventType string `db:"Event_type"`
	ServerId int64 `db:"Server_id"`
	EndPosition int64 `db:"End_log_pos"`
	EventInfo string `db:"Info"`
}
var BinlogEventsList []*BinlogEvents

//GTID_SUBTRACT
type BinlogSub struct {
	Substract string `db:"gtid_substract"`
}

//MySQL参数gtid_executed
type BinlogExec struct {
	VariablesName string `db:"Variable_name"`
	Values string `db:"Value"`
}

//MySQL参数gtid_purged
type BinlogPurg struct {
	VariablesName string `db:"Variable_name"`
	Values string `db:"Value"`
}

//GTID_SUBSET
type BinlogSet struct {
	SubSet int `db:"gtid_subset"`
}


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

	//获取binlog中的Previous_gtids
	for _, v := range BinaryLogsList {
		EventsQuery := fmt.Sprintf("SHOW BINLOG EVENTS IN '%s' LIMIT 1,1;", v.BinlogName)
		err = db.Select(&BinlogEventsList, EventsQuery)
		if err != nil{
			log.Errorf("show binlog events error: %s", err.Error())
			return err
		}
	}

	//获取当前binlog中的gtid
	ExecQuery := fmt.Sprintf("SHOW GLOBAL VARIABLES LIKE 'gtid_executed';")
	BinlogExecuted := new(BinlogExec)
	err = db.Get(BinlogExecuted, ExecQuery)
	if err != nil {
		log.Errorf("get gtid_executed err:", err.Error())
		return err
	}

	BinlogPurged := new(BinlogPurg)
	PurgQuery := fmt.Sprintf("SHOW GLOBAL VARIABLES LIKE 'gtid_purged';")
	err = db.Get(BinlogPurged, PurgQuery)
	if err != nil {
		log.Errorf("get gtid_purged err:", err.Error())
		return err
	}

	binLen := len(BinlogEventsList)
	BinlogGtidDict := make(map[string]string)
	firstQuery := fmt.Sprintf("SELECT GTID_SUBTRACT('%s', '%s') as gtid_substract;", BinlogExecuted.Values, BinlogPurged.Values)
	firstSub := BinlogSub{}
	err =  db.Get(&firstSub, firstQuery)
	if err != nil {
		log.Errorf("get gtid_purged/gtid_executed sub err:", err.Error())
		return err
	}
	firstFile := BinlogEventsList[binLen-1]
	firstFile.EventInfo = strings.Replace(firstFile.EventInfo, "\n", "", -1)
	//求gtid集合的差集
	firstGtidQuery := fmt.Sprintf("SELECT GTID_SUBTRACT('%s', '%s') as gtid_substract;", firstSub.Substract, firstFile.EventInfo)
	err = db.Get(firstSub, firstGtidQuery)
	BinlogGtidDict[firstFile.BinlogName] = firstSub.Substract

	//遍历最近一个binlog之前的binlog中的gtid
	for i := 1; i < binLen; i++ {
		afterFile := BinlogEventsList[binLen - i]
		beforeFile := BinlogEventsList[binLen - (i+1)]
		afterFile.EventInfo = strings.Replace(afterFile.EventInfo, "\n", "", -1)
		beforeFile.EventInfo = strings.Replace(beforeFile.EventInfo, "\n", "", -1)
		//BinlogEventsList = BinlogEventsList[:len(BinlogEventsList)-1]
		CompareQuery := fmt.Sprintf("SELECT GTID_SUBTRACT('%s', '%s') as gtid_substract;", afterFile.EventInfo, beforeFile.EventInfo)

		gtidSub := BinlogSub{}
		err = db.Get(&gtidSub, CompareQuery)
		if err != nil{
			return err
		}
		//fmt.Println(gtidSub.Substract)

		BinlogGtidDict[beforeFile.BinlogName] = gtidSub.Substract
	}

	for key, valus := range BinlogGtidDict {
		//查找gitd是否在binlog的gtid集合中
		Query := fmt.Sprintf("SELECT GTID_SUBSET('%s', '%s') as gtid_subset;", this.searchContext.GtidSearch, valus)
		searchGtid := BinlogSet{}
		err = db.Get(&searchGtid, Query)
		if err != nil {
			return err
		}
		switch searchGtid.SubSet {
		case 1:
			log.Info("Find GTID in Binary Log File: ", key)
			return nil
		}
	}
	log.Infof("GTID Not search...")
	return nil
}

