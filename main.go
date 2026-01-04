package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start))
	}
}

func zoneOffset(t time.Time) int {
	_, offset := t.Zone()
	return offset
}

// Find the next time at which a timezone transition occurs.
// If a transition is found within the next 4 weeks,
// return the time at which the transition occurs, and the new timezone offset.
// If no transition is found return an error.
func findNextTimezoneTransition(start time.Time) (time.Time, error) {
	end := start.Add(time.Hour * 24 * 7 * 4)

	if zoneOffset(start) == zoneOffset(end) {
		return time.Time{}, errors.New("no transition")
	}

	transTime := recursiveTzTransitionSearch(
		start.Truncate(time.Second),
		end.Truncate(time.Second),
	)

	return transTime, nil
}

func recursiveTzTransitionSearch(start time.Time, end time.Time) time.Time {
	timeDiff := end.Sub(start)

	if timeDiff == time.Second {
		return end
	}

	mid := start.Add((timeDiff / 2).Truncate(time.Second))

	if zoneOffset(start) == zoneOffset(mid) {
		return recursiveTzTransitionSearch(mid, end)
	}

	return recursiveTzTransitionSearch(start, mid)
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	var now time.Time
	if currentTimeOverride != nil {
		now = *currentTimeOverride
	} else {
		now = time.Now()
	}

	timezone := r.Header.Get("X-Timezone")
	if timezone == "" {
		timezone = "UTC"
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		http.Error(w, "invalid timezone", http.StatusBadRequest)
		return
	}

	now = now.In(location)

	nowMillis := now.UnixMilli()
	nowOffset := int64(zoneOffset(now))

	var transMillis, transOffset *int64

	offsetChangeTimestamp, err := findNextTimezoneTransition(now)
	if err == nil {
		millis := offsetChangeTimestamp.UnixMilli()
		offset := int64(zoneOffset(offsetChangeTimestamp))
		transMillis = &millis
		transOffset = &offset
	}

	data := [4]*int64{
		&nowMillis,
		&nowOffset,
		transMillis,
		transOffset,
	}
	buf, err := json.Marshal(data)
	if err != nil {
		log.Printf("failed to encode response: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

var currentTimeOverride *time.Time

func init() {
	if override := os.Getenv("OVERRIDE_CURRENT_TIME"); override != "" {
		if ts, err := strconv.ParseInt(override, 10, 64); err == nil {
			currentTime := time.Unix(ts, 0)
			currentTimeOverride = &currentTime
		}
	}
}

func main() {
	http.HandleFunc("/time", loggingMiddleware(timeHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
