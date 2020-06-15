package main

import (
	"GTID_search/utils"
	"flag"
	"fmt"
	"os"

	"github.com/outbrain/golib/log"
)

var AppVersion string

func main() {
	searchContext := utils.NewSearchContext()
	flag.StringVar(&searchContext.Host, "host", "127.0.0.1", "MySQL hostname")
	flag.StringVar(&searchContext.User, "user", "", "MySQL User")
	flag.StringVar(&searchContext.Password, "password", "", "MySQL password")
	flag.IntVar(&searchContext.Port, "port", 3306, "MySQL port")
	flag.StringVar(&searchContext.GtidSearch, "gtid-search", "", "GTID to find")
	help := flag.Bool("help", false, "Display usage")
	version := flag.Bool("version", false, "Print version & exit")
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	if *help {
		fmt.Fprintf(os.Stdout, "Uasge of GTID_search:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *version {
		appVersion := AppVersion
		if appVersion == "" {
			appVersion = "unversioned"
		}
		fmt.Println(appVersion)
		return
	}

	//if searchContext.GtidSearch == "" {
	//	log.Fatalf("--gtid-search must be provided and must not be empty")
	//	return
	//}

	log.SetLevel(log.INFO)
	log.Infof("starting GTID_search %+v", AppVersion)
	searchor := utils.NewSearchor(searchContext)
	err := searchor.SearchGtid()
	if err != nil {
		return
	}

}
