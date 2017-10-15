// real monitoring code (fetches monitored url, computes latency etc...)
package wmmon


import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

type MonResult struct {
	Url string
	When time.Time
	Err string // using string to avoid datastore serialization troubles
	Latency time.Duration
	StatusCode int // -1 if unknown/error
	Length int // -1 when unknown
}

// NOTE: we expect errors - we return them in structure...

func MonitorUrl(ctx appengine.Context, url string) MonResult{

	var res = MonResult{
		Url:	url,
		When:	time.Now(),
		StatusCode: -1,
		Length: -1,
	}

	var client = urlfetch.Client(ctx)

	var tic = time.Now()
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
		res.Err =  fmt.Sprintf("URL '%s' returned unexpected status %d <> %d, body: %s", url, resp.Status, OkHttpStatus, body)
		return res
	}

	return res
}

func EntityKind( url string ) string {
	var bytes = []byte(url)
	return fmt.Sprintf("%x",md5.Sum(bytes))
}

// monitor (fetch) specific url and stores it to datastore
// NOTE: error is returned for datastore errors only
func MonitorAndStoreUrl(ctx appengine.Context, url string) (MonResult,error) {
	var result = MonitorUrl(ctx,url)

	var entityKind = EntityKind(url)
	ctx.Infof("EntityKind '%s'",entityKind)

	// put results to datastore 
	var key = datastore.NewIncompleteKey(ctx, entityKind, nil)
	_, err := datastore.Put(ctx, key, &result)
	if err != nil {
		ctx.Errorf("Error on Put('%s'): %v",entityKind,err)
		return result,err
	}

	return result,nil
}