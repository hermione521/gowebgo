package webpagetest

import (
	"github.com/bitly/go-simplejson"
	"../globals"
	"fmt"
	"io/ioutil"
	"encoding/csv"
	"net/http"
	"time"
	"os"
	"strings"
)

func RunSingleTest(url string) string {
	// first, get test status
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// parse json
	status, err := simplejson.NewJson(body)
	if err != nil {
		panic(err)
	}
	if globals.Config().DebugMode {
		fmt.Println(status)
	}
	if status.Get("statusCode").MustInt() != 200 {
		fmt.Println("Test error!")
		return ""
	}

	resultUrl := status.Get("data").Get("jsonUrl").MustString()
	return resultUrl
}

func GetSingleResult(resultUrl string) *simplejson.Json {
	for true {
		// second, get result
		resp, err := http.Get(resultUrl)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		// parse json
		finalResult, err := simplejson.NewJson(body)
		if err != nil {
			fmt.Println(err)
		}
		code := finalResult.Get("statusCode").MustInt()
		if code / 100 == 4 {
			fmt.Println("Test error!")
		} else if code == 200 {
			return finalResult
		}

		time.Sleep(10 * time.Second)
	}

	return simplejson.New()
}

func GetHistoryResult(url string) []ReportDataCluster {
	config := globals.Config()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	csvReader := csv.NewReader(resp.Body)
	csvRecords, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
	}
	if config.DebugMode {
		fmt.Println(csvRecords)
	}

	// manage csv records
	if (len(csvRecords) - 1) % 3 != 0 {
		fmt.Println("warning: the number of test history is not multiple of 3.")
	}
	numCluster := (len(csvRecords) - 1) / 3
	csvRecords = csvRecords[1:]

	ans := make([]ReportDataCluster, numCluster)

	// get all report data from csv records
	for i := numCluster - 1; i >= 0; i-- {
		ans[numCluster - i - 1].ClusterId = csvRecords[i * 3][4]
		for j := 0; j < 3; j++ {
			resultUrl := globals.FmtResultUrl(csvRecords[i * 3 + j][2])
			resp, err = http.Get(resultUrl)
			if err != nil {
				fmt.Println(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			testResult, err := simplejson.NewJson(body)
			if err != nil {
				fmt.Println(err)
			}

			var curReport ReportData
			curReport.Id = testResult.GetPath("data", "id").MustString()
			curReport.TestUrl = testResult.GetPath("data", "testUrl").MustString()
			curReport.From = testResult.GetPath("data", "from").MustString()
			curReport.Date = time.Unix(testResult.GetPath("data", "completed").MustInt64(), 0).Format("2006-01-02 15:04")
			curReport.FirstView = genViewSummary(testResult.GetPath("data", "runs", "1", "firstView"))
			curReport.RepeatView = genViewSummary(testResult.GetPath("data", "runs", "1", "repeatView"))

			browser := testResult.GetPath("data", "location").MustString()
			browser = browser[strings.Index(browser, ":") + 1:]
			switch browser {
				case "Chrome":
					ans[numCluster - i - 1].ChromeData = curReport
				case "Firefox":
					ans[numCluster - i - 1].FirefoxData = curReport
				case "IE":
					ans[numCluster - i - 1].IEData = curReport
			}

			filename := "./output/" + curReport.Id + ".html"
			if _, err = os.Stat(filename); os.IsNotExist(err) {
				// single report not exist
				curUrl := globals.FmtResultUrl(curReport.Id)
				curRes := GetSingleResult(curUrl)
				GenCharts(curRes)
			}
		}
	}

	return ans
}