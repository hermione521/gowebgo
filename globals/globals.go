package globals

import(
	"github.com/google/go-querystring/query"
	"io/ioutil"
	"encoding/json"
	"html/template"
	"fmt"
)

type Configuration struct{
	DebugMode bool
	WebpagetestURL string
	WebpagetestSuffix string
	WebpagetestOption WPTOption
	WebpagetestCharts []WPTChartsName
}

type WPTOption struct {
	Url string `url:"url"`
	Label string `url:"label"`
	Location string `url:"location"`
	Private string `url:"private"`
	Video string `url:"video"`
	F string `url:"f"`
}

type WPTChartsName struct {
	Title string
	CanvasId string
	ChartType template.JS
}

var _config = Configuration{}

func Config() Configuration {
	if _config.WebpagetestURL != "" {
		return _config
	}
	file, err := ioutil.ReadFile("./settings/config.json")
	if err != nil {
		panic(err);
	}
	err = json.Unmarshal(file, &_config)
	if err != nil {
		panic(err)
	}
	return _config
}

func SetTestUrl(url string) {
	_config.WebpagetestOption.Url = url
}

func SetDebugMode(debugMode bool) {
	_config.DebugMode = debugMode
}

func FmtQuery(browser, label string) string {
	tempLoc := _config.WebpagetestOption.Location
	tempLabel := _config.WebpagetestOption.Label

	_config.WebpagetestOption.Location = fmt.Sprintf(tempLoc, browser)
	_config.WebpagetestOption.Label = tempLabel + label

	url := _config.WebpagetestURL + _config.WebpagetestSuffix + "?"
	v, _ := query.Values(_config.WebpagetestOption)
	url += v.Encode()

	_config.WebpagetestOption.Location = tempLoc
	_config.WebpagetestOption.Label = tempLabel
	return url
}

func FmtHistoryQuery() string {
	url := _config.WebpagetestURL + "/testlog.php?days=30&&all=on&f=csv&url=" + _config.WebpagetestOption.Url + "&label=" + _config.WebpagetestOption.Label
	// url = url.Encode()
	return url
}

func FmtResultUrl(ID string) string {
	url := _config.WebpagetestURL + "/jsonResult.php?test=" + ID
	return url
}