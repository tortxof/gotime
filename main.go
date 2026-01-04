package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

func ZoneOffset(t time.Time) int {
	_, offset := t.Zone()
	return offset
}

// Find the next time at which a timezone transition occurs.
// If a transition is found within the next 2 hours,
// return the time at which the transition occurs, and the new timezone offset.
// If no transition is found return an error.
func findNextTimezoneTransition(start time.Time) (time.Time, error) {
	end := start.Add(time.Hour * 2)

	if ZoneOffset(start) == ZoneOffset(end) {
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

	if ZoneOffset(start) == ZoneOffset(mid) {
		return recursiveTzTransitionSearch(mid, end)
	}

	return recursiveTzTransitionSearch(start, mid)
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

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

	var transMillis, transOffset int

	offsetChangeTimestamp, err := findNextTimezoneTransition(now)
	if err == nil {
		transMillis = int(offsetChangeTimestamp.UnixMilli())
		transOffset = ZoneOffset(offsetChangeTimestamp)
	}

	data := [4]int{
		int(now.UnixMilli()),
		ZoneOffset(now),
		transMillis,
		transOffset,
	} // [<current millisecond timestamp>, <UTC offset in seconds based on timezone>, <millisecond timestamp at next offset change>, <new offset>]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/time", timeHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
