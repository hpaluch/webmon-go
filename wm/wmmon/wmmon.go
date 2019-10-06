// real monitoring code (fetches monitored url, computes latency etc...)
package wmmon

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/log"

	"github.com/hpaluch/webmon-go/wm/wmconsts"

)

type MonResult struct {
	Url        string
	When       time.Time
	Err        string // using string to avoid datastore serialization troubles
	Latency    time.Duration
	StatusCode int // -1 if unknown/error
	Length     int // -1 when unknown
}

// NOTE: we expect errors - we return them in structure...

func MonitorUrl(ctx context.Context, url string) MonResult {

	var res = MonResult{
		Url:        url,
		When:       time.Now(),
		StatusCode: -1,
		Length:     -1,
	}

	var client = urlfetch.Client(ctx)

	var tic = time.Now()
	// timeout code from: https://stackoverflow.com/a/25344458
	var timeout = time.Duration(wmconsts.FetchTimeoutSecs * time.Second)
	client.Timeout = timeout
	resp, err := client.Get(url)
	res.StatusCode = resp.StatusCode
	if err != nil {
		res.Err = err.Error()
		res.Latency = time.Since(tic)
		return res
	}

	const OkHttpStatus = 200
	// https://blog.alexellis.io/golang-json-api-client/
	body, err := ioutil.ReadAll(resp.Body)
	res.Latency = time.Since(tic)
	if err != nil {
		res.Err = err.Error()
		return res
	}

	res.Length = len(body)

	if resp.StatusCode != OkHttpStatus {
		res.Err = fmt.Sprintf("URL '%s' returned unexpected status %d <> %d: '%s' Body: %s", url, resp.StatusCode, OkHttpStatus, resp.Status, body)
		if len(res.Err) > wmconsts.DataStoreMaxStrLen {
			const suffix = " ..."
			res.Err = res.Err[0:(wmconsts.DataStoreMaxStrLen-len(suffix))] + suffix
		}
		return res
	}

	return res
}

func EntityKind(url string) string {
	var bytes = []byte(url)
	return fmt.Sprintf("%x", md5.Sum(bytes))
}

// monitor (fetch) specific url and stores it to datastore
// NOTE: error is returned for datastore errors only
func MonitorAndStoreUrl(ctx context.Context, url string) (MonResult, error) {
	var result = MonitorUrl(ctx, url)

	var entityKind = EntityKind(url)
	log.Infof(ctx, "EntityKind '%s'", entityKind)

	// put results to datastore
	var key = datastore.NewIncompleteKey(ctx, entityKind, nil)
	_, err := datastore.Put(ctx, key, &result)
	if err != nil {
		log.Errorf(ctx, "Error on Put('%s'): %v", entityKind, err)
		return result, err
	}

	return result, nil
}
