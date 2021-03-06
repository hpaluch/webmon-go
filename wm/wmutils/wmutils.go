// misc utilities for WebMon application
package wmutils

import (
	"net/http"
	"time"

	"google.golang.org/appengine/log"
	"golang.org/x/net/context"
)

func RoundDurationToMs(d time.Duration) time.Duration {
	return ((d + time.Millisecond/2) / time.Millisecond) * time.Millisecond
}

// data model for templates/zz_layout.html
type LayoutModel struct {
	NowUTC     time.Time
	RenderTime time.Duration
	Title      string
}

func CreateLayoutModel(tic time.Time, title string, ctx context.Context, r *http.Request) (LayoutModel, error) {

	return LayoutModel{
		NowUTC:     time.Now(),
		RenderTime: time.Since(tic),
		Title:      title,
	}, nil
}

func VerifyGetMethod(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {

	// how to trigger this error:
	// curl -X POST -v http://localhost:8080
	if r.Method != "GET" {
		log.Errorf(ctx, "Method '%s' not allowed for path '%s'",
			r.Method, r.URL.Path)
		http.Error(w, "Method not allowed",
			http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func NoCacheHeaders(w http.ResponseWriter) {
	// look at headers of www.seznam.cz :-)
	// WARNING! This also sets Cache-Control: no-cache
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
}
