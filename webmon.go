package zolist

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"appengine"

	"github.com/hpaluch/webmon-go/wm/wmutils"
)

var CzechLocation *time.Location

func tplCzDateStr(timeArg time.Time) string {
	return timeArg.In(CzechLocation).Format("02.01.2006 15:04:05 MST")
}

func tplCzDateStrWithAgo(timeArg time.Time) string {
	var dateStr = tplCzDateStr(timeArg)
	// compute ago
	var czNow = time.Now().In(CzechLocation)
	var duration = czNow.Sub(timeArg)
	duration = wmutils.RoundDurationToMs(duration)
	var czAgo = time.Duration(duration).String()

	var str = fmt.Sprintf("%s (%s ago)", dateStr, czAgo)
	return str
}

var (
	tplFn = template.FuncMap{
		"ZoCzDateFormat":        tplCzDateStr,
		"ZoCzDateFormatWithAgo": tplCzDateStrWithAgo,
	}

	// from: https://github.com/golang/appengine/blob/master/demos/guestbook/guestbook.go
	tpl = template.Must(template.New("").Funcs(tplFn).ParseGlob("templates/*.html"))

	str_mon_urls   = os.Getenv("MON_URLS")
	// initialized in init()
	mon_urls []string
)

type WebData struct {
	Url string
}

type ListModel struct {
	LayoutModel wmutils.LayoutModel
	WebData []WebData
}

func handlerList(w http.ResponseWriter, r *http.Request) {
	var tic = time.Now()
	var ctx = appengine.NewContext(r)
	wmutils.NoCacheHeaders(w)
	// report 404 for other path than "/"
	// see https://github.com/GoogleCloudPlatform/golang-samples/blob/master/appengine_flexible/helloworld/helloworld.go
	const MY_PATH = "/"
	if r.URL.Path != MY_PATH {
		ctx.Errorf("Unexpected path '%s' <> '%s'", r.URL.Path, MY_PATH)
		http.NotFound(w, r)
		return
	}

	if !wmutils.VerifyGetMethod(ctx, w, r) {
		return
	}

	webData := make([]WebData,len(mon_urls))
	for i,url := range mon_urls {
		var wd = WebData{
			Url: url,
		}
		webData [i] = wd
	}


	layoutModel, err := wmutils.CreateLayoutModel(tic, fmt.Sprintf("WebMon - Web monitor in Go"), ctx, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	listModel := ListModel{
		LayoutModel: layoutModel,
		WebData: webData,
	}

	if err := tpl.ExecuteTemplate(w, "home.html", listModel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// main handler fo Go/GAE application
func init() {
	var err error
	CzechLocation, err = time.LoadLocation("Europe/Prague")
	if err != nil {
		panic(fmt.Sprintf("Fatal error - unable to load timezone: %v", err))
	}

	if str_mon_urls == "" {
		panic("Fatal error - missing/empty MON_URLS in app.yaml")
	}

	mon_urls = strings.Split(str_mon_urls, " ")
	if len(mon_urls) == 0 {
		panic("No id found in MON_URLS")
	}

	for _, v := range mon_urls {
		_, err := url.ParseRequestURI(v)
		if err != nil {
			panic(fmt.Sprintf("Unable to parse '%s' as Url: %v",
				v, err))
		}
	}

	http.HandleFunc("/", handlerList)
}