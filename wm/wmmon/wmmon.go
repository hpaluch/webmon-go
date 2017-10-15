// real monitoring code (fetches monitored url, computes latency etc...)
package wmmon


import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"appengine"
	"appengine/urlfetch"
)

type MonResult struct {
	Url string
	When time.Time
	Err error
	Latency time.Duration
	StatusCode int
	Length int
}

// NOTE: we expect errors - we return them in structure...

func MonitorUrl(ctx appengine.Context, url string) MonResult{

	var res = MonResult{
		Url:	url,
		When:	time.Now(),
	}

	var client = urlfetch.Client(ctx)

	var tic = time.Now()
	resp, err := client.Get(url)
	res.StatusCode = resp.StatusCode
	if err != nil {
		res.Err = err
		res.Latency = time.Since(tic)
		return res
	}

	const OkHttpStatus = 200
	// https://blog.alexellis.io/golang-json-api-client/
	body, err := ioutil.ReadAll(resp.Body)
	res.Latency = time.Since(tic)
	if err != nil {
		res.Err = err
		return res
	}

	res.Length = len(body)

	if resp.StatusCode != OkHttpStatus {
		res.Err =  errors.New(fmt.Sprintf("URL '%s' returned unexpected status %d <> %d, body: %s", url, resp.Status, OkHttpStatus, body))
		return res
	}

	return res
}
