package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/taskqueue"

	"google.golang.org/appengine/log"

	"github.com/hpaluch/webmon-go/wm/wmmon"
	"github.com/hpaluch/webmon-go/wm/wmutils"
)

var CzechLocation *time.Location

func tplCzDateStr(timeArg time.Time) string {
	return timeArg.In(CzechLocation).Format("02.01.2006 15:04:05 MST")
}

func tplDurationMs(d time.Duration) string {
	return wmutils.RoundDurationToMs(d).String()
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
		"CzDateFormat":        tplCzDateStr,
		"CzDateFormatWithAgo": tplCzDateStrWithAgo,
		"DurationMs":          tplDurationMs,
	}

	// from: https://github.com/golang/appengine/blob/master/demos/guestbook/guestbook.go
	tpl = template.Must(template.New("").Funcs(tplFn).ParseGlob("templates/*.html"))

	str_mon_urls = os.Getenv("MON_URLS")
	// initialized in init()
	mon_urls []string
)

type WebData struct {
	Url     string
	Results []wmmon.MonResult
}

type ListModel struct {
	LayoutModel wmutils.LayoutModel
	WebData     []WebData
}

func handlerList(w http.ResponseWriter, r *http.Request) {
	var tic = time.Now()
	var ctx = appengine.NewContext(r)
	wmutils.NoCacheHeaders(w)
	// report 404 for other path than "/"
	// see https://github.com/GoogleCloudPlatform/golang-samples/blob/master/appengine_flexible/helloworld/helloworld.go
	const MY_PATH = "/"
	if r.URL.Path != MY_PATH {
		log.Errorf(ctx, "Unexpected path '%s' <> '%s'", r.URL.Path, MY_PATH)
		http.NotFound(w, r)
		return
	}

	if !wmutils.VerifyGetMethod(ctx, w, r) {
		return
	}

	webData := make([]WebData, len(mon_urls))
	for i, url := range mon_urls {

		var entityKind = wmmon.EntityKind(url)
		q := datastore.NewQuery(entityKind).Order("-When").Limit(100)
		var results []wmmon.MonResult
		var _, err = q.GetAll(ctx, &results)
		if err != nil {
			log.Errorf(ctx, "Error fetchhing entities for '%s': %v",
				url, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}

		var wd = WebData{
			Url:     url,
			Results: results,
		}
		webData[i] = wd
	}

	layoutModel, err := wmutils.CreateLayoutModel(tic, fmt.Sprintf("WebMon - Web monitor in Go"), ctx, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	listModel := ListModel{
		LayoutModel: layoutModel,
		WebData:     webData,
	}

	if err := tpl.ExecuteTemplate(w, "home.html", listModel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handlerCron(w http.ResponseWriter, r *http.Request) {

	var tic = time.Now()
	var ctx = appengine.NewContext(r)
	wmutils.NoCacheHeaders(w)

	const MY_PATH = "/cron"
	if r.URL.Path != MY_PATH {
		log.Errorf(ctx, "Unexpected path '%s' <> '%s'", r.URL.Path, MY_PATH)
		http.NotFound(w, r)
		return
	}

	if !wmutils.VerifyGetMethod(ctx, w, r) {
		return
	}

	// for Production verify that it is called by cron
	if !appengine.IsDevAppServer() {
		var cronHeader = r.Header.Get("X-Appengine-Cron")
		if cronHeader == "" {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Can be invoked from Cron only")
			return
		}
	}

	var txt = ""
	// our Cron job - Enqueue tasks to Worker - each Url = one Task
	for i, urlx := range mon_urls {
		var task = taskqueue.NewPOSTTask("/worker", url.Values{
			"index": {strconv.Itoa(i)},
		})
		if _, err := taskqueue.Add(ctx, task, ""); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		txt += fmt.Sprintf("Enqueued task for url %s\r\n", urlx)
	}
	txt += fmt.Sprintf("Cron finished in %v\r\n", time.Since(tic))
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	fmt.Fprintf(w, "%s", txt)

}

func handlerWorker(w http.ResponseWriter, r *http.Request) {
	var tic = time.Now()
	var ctx = appengine.NewContext(r)
	wmutils.NoCacheHeaders(w)

	const MY_PATH = "/worker"
	if r.URL.Path != MY_PATH {
		log.Errorf(ctx, "Unexpected path '%s' <> '%s'", r.URL.Path, MY_PATH)
		http.NotFound(w, r)
		return
	}

	// url index
	var iStr = r.FormValue("index")
	i, err := strconv.Atoi(iStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if i < 0 || i >= len(mon_urls) {
		http.Error(w, "url index out of range of MON_URLS",
			http.StatusInternalServerError)
		return
	}

	// and finally run our worker job :-)
	result, err := wmmon.MonitorAndStoreUrl(ctx, mon_urls[i])
	if err != nil {
		log.Errorf(ctx, "Error running worker for url %s: %v", mon_urls[i], err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Infof(ctx, "Finished worker on %v in %s", result, time.Since(tic))
	var txt = fmt.Sprintf("Succes on worker %v\r\n", result)
	txt += fmt.Sprintf("Job finished in %v\r\n", time.Since(tic))
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	fmt.Fprintf(w, "%s", txt)

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

	http.HandleFunc("/cron", handlerCron)
	http.HandleFunc("/worker", handlerWorker)
	http.HandleFunc("/", handlerList)
}

func main() {
	appengine.Main()
}
