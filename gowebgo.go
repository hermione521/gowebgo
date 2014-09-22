package main

import (
	"./webpagetest"
	"./globals"
	"fmt"
	"flag"
	"time"
)

func main() {
	// parse command line arguments
	urlPtr := flag.String("url", "", "site url that you want to test")
	newTestPtr := flag.Bool("newTest", false, "set true to start a new test")
	historyPtr := flag.Bool("history", true, "set false to disable generating history report")
	debugPtr := flag.Bool("debugMode", false, "set true to show debug message")
	flag.Parse()
	globals.Config()

	// set configuration
	if *urlPtr != "" {
		globals.SetTestUrl(*urlPtr)
	}
	globals.SetDebugMode(*debugPtr)

	// generate new test results
	if *newTestPtr {
		browsers := [3]string{"Chrome", "Firefox", "IE"}
		wptLabel := time.Now().Local().Format("20060102150405")
		for i := 0; i < len(browsers); i++ {
			fmt.Println("New test for", browsers[i], "is for running...")
			
			url := globals.FmtQuery(browsers[i], wptLabel)
			if globals.Config().DebugMode {
				fmt.Println("Query url:", url)
			}
			resultUrl := webpagetest.RunSingleTest(url)
			testResult := webpagetest.GetSingleResult(resultUrl)
			webpagetest.GenCharts(testResult)

			fmt.Println("New test for", browsers[i], "is done.")
		}
	}
	
	// generate history report
	if *historyPtr {
		fmt.Println("Generating history report...")

		url := globals.FmtHistoryQuery()
		dataClusters := webpagetest.GetHistoryResult(url)
		if globals.Config().DebugMode {
			fmt.Println(dataClusters)
		}
		webpagetest.GenHistory(dataClusters)

		fmt.Println("Generating history report finished.")
	}
}
