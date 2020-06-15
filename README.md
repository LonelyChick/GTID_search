# GTID_search
通过GTID快速定位Binlog文件

```cassandraql
.
├── README.md
├── cmd
│   └── main.go
├── go.mod
├── go.sum
├── utils
│   └── base.go
└── vendor
    ├── github.com
    └── modules.txt

```

## 安装：
```cassandraql
git clone git@github.com:LonelyChick/GTID_search.git
cd GTID_search
go mod download
go mod verify
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gtid_search cmd/main.go
```

## 使用示例：
```cassandraql
./gtid_search -host="127.0.0.1" -port=3306 -user="wt_test" -password="xxx" -gtid-search="xxxxxx:1"
```

