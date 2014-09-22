package webpagetest

import (
	"github.com/bitly/go-simplejson"
	"fmt"
	"encoding/json"
	"../globals"
	"html/template"
	"net/http"
	"io/ioutil"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

type ChartPart struct {
	Value int
	Color string
	// Highlight string `json:"highlight"`
	Label string
}

type ChartData struct {
	Title string
	CanvasId string
	Data []template.JS
	Colors []template.JS
	ChartType template.JS
}

type ViewSummary struct {
	Backend float64
	StartRender float64
	DomStart float64
	DomEnd float64
	DomDelta float64
	FullyLoaded float64
	SpeedIndex int
}

type ReportData struct {
	Id string
	ResultUrl string
	TestUrl string
	From string
	Date string
	FirstView ViewSummary
	RepeatView ViewSummary
	Charts []ChartData
	FirstWaterfall string
	RepeatWaterfall string
}

type ReportDataCluster struct {
	ClusterId string
	ChromeData ReportData
	FirefoxData ReportData
	IEData ReportData
}

type HistoryReport struct {
	TestUrl string
	WPTUrl string
	Data template.JS
}

func genViewSummary(view *simplejson.Json) ViewSummary {
	summary := ViewSummary{
		Backend: float64(view.Get("TTFB").MustInt()) / 1000,
		StartRender: float64(view.Get("render").MustInt()) / 1000,
		DomStart: float64(view.Get("domContentLoadedEventStart").MustInt()) / 1000,
		DomEnd: float64(view.Get("domContentLoadedEventEnd").MustInt()) / 1000,
		FullyLoaded: float64(view.Get("fullyLoaded").MustInt()) / 1000,
		SpeedIndex: view.Get("SpeedIndex").MustInt(),
	}
	summary.DomDelta = summary.DomEnd - summary.DomStart
	summary.DomDelta = float64(int(summary.DomDelta * 1000)) / 1000

	return summary
}

func interfaceToInt(interfaceArr []interface{}) []int {
	intArr := make([]int, len(interfaceArr))
	for i, v := range interfaceArr {
		// TODO: better way to convert json.Number to int
		str, _ := json.Marshal(v)
		var temp int
		err := json.Unmarshal(str, &temp)
		if err != nil {
			panic(err)
		}
		intArr[i] = temp
	}
	return intArr
}

/////////////////////////////////////////////////////////////////////////////////////////////

func averageBytesPerPageByContentType(data *simplejson.Json) ([]template.JS, []template.JS) {
	breakdown := data.GetPath("runs", "1", "firstView", "breakdown")

	var chartParts []template.JS
	var colorStrings []template.JS
	var curPart ChartPart
	labels := []string{"html", "js", "css", "image", "flash", "font", "other"}

	chartParts = append(chartParts, "['label', 'percentage']")
	for _, label := range labels {
		curPart.Value = breakdown.Get(label).Get("bytes").MustInt()

		// convert colors from []interface{} to []int
		colors := breakdown.Get(label).Get("color").MustArray()
		intColors := interfaceToInt(colors)
		curPart.Color = fmt.Sprintf("rgb(%d,%d,%d)", intColors[0], intColors[1], intColors[2])

		curPart.Label = label

		chartParts = append(chartParts, template.JS(fmt.Sprintf("['%s', %d]", curPart.Label, curPart.Value)))
		colorStrings = append(colorStrings, template.JS(curPart.Color))
	}

	return chartParts, colorStrings
}

func averageRequestsPerPageByContentType(data *simplejson.Json) ([]template.JS, []template.JS) {
	breakdown := data.GetPath("runs", "1", "firstView", "breakdown")

	var chartParts []template.JS
	var colorStrings []template.JS
	var curPart ChartPart
	labels := []string{"html", "js", "css", "image", "flash", "font", "other"}

	chartParts = append(chartParts, "['label', 'percentage']")
	for _, label := range labels {
		curPart.Value = breakdown.Get(label).Get("requests").MustInt()

		// convert colors from []interface{} to []int
		colors := breakdown.Get(label).Get("color").MustArray()
		intColors := interfaceToInt(colors)
		curPart.Color = fmt.Sprintf("rgb(%d,%d,%d)", intColors[0], intColors[1], intColors[2])

		curPart.Label = label

		chartParts = append(chartParts, template.JS(fmt.Sprintf("['%s', %d]", curPart.Label, curPart.Value)))
		colorStrings = append(colorStrings, template.JS(curPart.Color))
	}

	return chartParts, colorStrings
}

func imageRequestsByFormat(data *simplejson.Json) ([]template.JS, []template.JS) {
	requests := data.GetPath("runs", "1", "firstView", "requests")

	var chartParts []template.JS
	var colorStrings []template.JS
	chartMap := make(map[string] ChartPart)

	chartParts = append(chartParts, "['label', 'requests']")
	labels := []string{"png", "jpeg", "gif", "Other"}
	colors := []string{"rgb(178,234,148)", "rgb(254,197,132)", "rgb(196,154,232)", "rgb(130,181,252)"}
	for i, label := range labels {
		curPart := ChartPart{Label: label, Value: 0, Color: colors[i]}
		chartMap[label] = curPart
	}

	for i := 0; i < len(requests.MustArray()); i++ {
		request := requests.GetIndex(i)

		contentType := request.Get("contentType")
		if contentType == nil {
			continue
		}
		curLabel := contentType.MustString()
		if len(curLabel) < 5 || curLabel[:5] != "image" {
			continue
		}
		// find image type
		curLabel = curLabel[6:]
		
		if val, ok := chartMap[curLabel]; !ok {
			val = chartMap["Other"]
			val.Value++
			chartMap["Other"] = val
		} else {
			val.Value++
			chartMap[curLabel] = val
		}
	}

	for _, label := range labels {
		chartParts = append(chartParts, template.JS(fmt.Sprintf("['%s', %d]", chartMap[label].Label, chartMap[label].Value)))
		colorStrings = append(colorStrings, template.JS(chartMap[label].Color))
	}

	return chartParts, colorStrings
}

func httpRequests(data *simplejson.Json) ([]template.JS, []template.JS) {
	requests := data.GetPath("runs", "1", "firstView", "requests")

	var chartParts []template.JS
	var colorStrings []template.JS
	chartMap := make(map[string] ChartPart)

	chartParts = append(chartParts, "['label', 'requests']")
	labels := []string{"https", "http", "Other"}
	colors := []string{"rgb(178,234,148)", "rgb(254,197,132)", "rgb(130,181,252)"}
	for i, label := range labels {
		curPart := ChartPart{Label: label, Value: 0, Color: colors[i]}
		chartMap[label] = curPart
	}

	for i := 0; i < len(requests.MustArray()); i++ {
		request := requests.GetIndex(i)

		fullUrl := request.Get("full_url")
		if fullUrl == nil {
			continue
		}
		httpType := fullUrl.MustString()
		pos := strings.Index(httpType, "://")
		if pos == -1 {
			continue
		}

		// find image type
		httpType = httpType[:pos]
		
		if val, ok := chartMap[httpType]; !ok {
			val = chartMap["Other"]
			val.Value++
			chartMap["Other"] = val
		} else {
			val.Value++
			chartMap[httpType] = val
		}
	}

	for _, label := range labels {
		chartParts = append(chartParts, template.JS(fmt.Sprintf("['%s', %d]", chartMap[label].Label, chartMap[label].Value)))
		colorStrings = append(colorStrings, template.JS(chartMap[label].Color))
	}

	return chartParts, colorStrings
}

func cacheLifetime(data *simplejson.Json) ([]template.JS, []template.JS) {
	requests := data.GetPath("runs", "1", "firstView", "requests")

	chartParts := make([]ChartPart, 5)

	labels := []string{"t = 0", "0 < t <= 1", "1 < t <= 30", "30 < t <= 365", "t > 365"}
	colors := []string{"rgb(130,181,252)", "rgb(178,234,148)", "rgb(196,154,232)", "rgb(254,197,132)", "rgb(196, 196, 196)"}
	for i, label := range labels {
		chartParts[i] = ChartPart{Label: label, Value: 0, Color: colors[i]}
	}

	for i := 0; i < len(requests.MustArray()); i++ {
		request := requests.GetIndex(i)

		cacheTime := request.Get("cache_time")
		if cacheTime == nil {
			continue
		}
		time := cacheTime.MustInt()

		var index int
		switch {
		case time == 0:
			index = 0
		case time <= 24 * 60: // 1 day
			index = 1
		case time <= 24 * 60 * 30: // 1 month
			index = 2
		case time <= 24 * 60 * 365: // 1 year
			index = 3
		default:
			index = 4
		}

		chartParts[index].Value++
	}

	chartStrings := make([]template.JS, 6)
	chartStrings[0] = "['label', 'requests', { role: 'style' }]"
	for i, label := range labels {
		chartStrings[i + 1] = template.JS(fmt.Sprintf("['%s', %d, '%s']", label, chartParts[i].Value, chartParts[i].Color))
	}

	return chartStrings, []template.JS{}
}

func timeSummary(data *simplejson.Json) ([]template.JS, []template.JS) {
	chartParts := make([]template.JS, 6)
	chartParts[0] = "['time', 'first view', 'repeat view']"

	format := "['%s', %f, %f]"
	firstView := data.GetPath("runs", "1", "firstView")
	repeatView := data.GetPath("runs", "1", "repeatView")
	chartParts[1] = template.JS(fmt.Sprintf(format, "backend",
		float64(firstView.Get("TTFB").MustInt()) / 1000,
		float64(repeatView.Get("TTFB").MustInt()) / 1000))
	chartParts[2] = template.JS(fmt.Sprintf(format, "start render",
		float64(firstView.Get("render").MustInt()) / 1000,
		float64(repeatView.Get("render").MustInt()) / 1000))
	chartParts[3] = template.JS(fmt.Sprintf(format, "DOM start",
		float64(firstView.Get("domContentLoadedEventStart").MustInt()) / 1000,
		float64(repeatView.Get("domContentLoadedEventStart").MustInt()) / 1000))
	chartParts[4] = template.JS(fmt.Sprintf(format, "DOM end",
		float64(firstView.Get("domContentLoadedEventEnd").MustInt()) / 1000,
		float64(repeatView.Get("domContentLoadedEventEnd").MustInt()) / 1000))
	chartParts[5] = template.JS(fmt.Sprintf(format, "fully loaded",
		float64(firstView.Get("fullyLoaded").MustInt()) / 1000,
		float64(repeatView.Get("fullyLoaded").MustInt()) / 1000))

	return chartParts, []template.JS{}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

func GenCharts(testResult *simplejson.Json) {
	statusCode := testResult.Get("statusCode").MustInt()
	FVruns := testResult.GetPath("data", "successfulFVRuns").MustInt()
	RVruns := testResult.GetPath("data", "successfulRVRuns").MustInt()
	if statusCode != 200 || FVruns == 0 || RVruns == 0 {
		filename := "./output/" + testResult.GetPath("data", "id").MustString() + ".html"
		ioutil.WriteFile(filename,
			[]byte("<html><head><title>Broken Test</title><body>This test failed in webpagetest.</body></html>"),
			0777)
		return
	}

	funcs := map[string]func(*simplejson.Json) ([]template.JS, []template.JS) {
		"averageBytesPerPageByContentType": averageBytesPerPageByContentType,
		"averageRequestsPerPageByContentType": averageRequestsPerPageByContentType,
		"imageRequestsByFormat": imageRequestsByFormat,
		"httpRequests": httpRequests,
		"cacheLifetime": cacheLifetime,
		"timeSummary": timeSummary,
	}

	config := globals.Config()
	firstView := testResult.GetPath("data", "runs", "1", "firstView")
	repeatView := testResult.GetPath("data", "runs", "1", "repeatView")

	var report ReportData
	report.Id = testResult.GetPath("data", "id").MustString()
	report.ResultUrl = config.WebpagetestURL + "/result/" + string(report.Id)
	report.TestUrl = testResult.GetPath("data", "testUrl").MustString()
	report.From = testResult.GetPath("data", "from").MustString()
	report.Date = time.Unix(testResult.GetPath("data", "completed").MustInt64(), 0).Format("2006-01-02 15:04")

	report.FirstView = genViewSummary(firstView)
	report.RepeatView = genViewSummary(repeatView)

	report.FirstWaterfall = firstView.GetPath("images", "waterfall").MustString()
	report.RepeatWaterfall = repeatView.GetPath("images", "waterfall").MustString()
	report.Charts = make([]ChartData, len(config.WebpagetestCharts))

	// call chart functions
	for i, chart := range config.WebpagetestCharts {
		f := reflect.ValueOf(funcs[chart.CanvasId])
		param := make([]reflect.Value,1)
		param[0] = reflect.ValueOf(testResult.Get("data"))

		res := f.Call(param)
		report.Charts[i].Data = res[0].Interface().([]template.JS)
		report.Charts[i].Colors = res[1].Interface().([]template.JS)
		report.Charts[i].Title = chart.Title
		report.Charts[i].CanvasId = chart.CanvasId
		report.Charts[i].ChartType = chart.ChartType
	}
	if globals.Config().DebugMode {
		fmt.Println(report)
	}

	file, _ := ioutil.ReadFile("./template/detail.html")
	tmpl, err := template.New("detail").Parse(string(file))
	if err != nil {
		panic(err)
	}

	wfile, _ := os.OpenFile("./output/" + report.Id + ".html", os.O_RDWR | os.O_CREATE, 0777)
	err = tmpl.Execute(wfile, report)
	if err != nil {
		fmt.Println(err)
	}
}

func GenHistory(clusters []ReportDataCluster) {
	// generate compare image
	GenCompareImage(clusters[len(clusters) - 1])

	// generate html from template
	testUrl := clusters[0].ChromeData.TestUrl
	dataStr, _ := json.Marshal(clusters)

	report := HistoryReport{
		testUrl,
		globals.Config().WebpagetestURL,
		template.JS(dataStr),
	}

	file, _ := ioutil.ReadFile("./template/history.html")
	tmpl, err := template.New("history").Parse(string(file))
	if err != nil {
		panic(err)
	}

	wfile, _ := os.OpenFile("./output/history.html", os.O_RDWR | os.O_CREATE, 0777)
	err = tmpl.Execute(wfile, report)
	if err != nil {
		fmt.Println(err)
	}
}

func GenCompareImage(cluster ReportDataCluster) {
	config := globals.Config()

	url := config.WebpagetestURL + "/video/filmstrip.php?tests=" +
			cluster.ChromeData.Id + "," +
			cluster.FirefoxData.Id + "," +
			cluster.IEData.Id + "&thumbSize=100&ival=100&end=visual&text=ffffff&bg=000000"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	file, _ := os.OpenFile("./output/image/" + cluster.ClusterId + ".png", os.O_RDWR | os.O_CREATE, 0777)
	io.Copy(file, resp.Body)
}