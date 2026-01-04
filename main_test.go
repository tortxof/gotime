package main

import (
	"testing"
	"time"
)

func TestFindNextTimezoneTransition(t *testing.T) {
	// Load a timezone that has DST transitions
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("failed to load timezone: %v", err)
	}

	transitionTimes := [2]time.Time{
		time.Date(2024, 3, 10, 7, 0, 0, 0, time.UTC),
		time.Date(2024, 11, 3, 6, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name          string
		start         time.Time
		wantErr       bool
		wantTransTime time.Time
	}{
		{
			name:    "no spring transition in UTC",
			start:   transitionTimes[0].Add(-time.Hour),
			wantErr: true,
		},
		{
			name:    "no fall transition in UTC",
			start:   transitionTimes[1].Add(-time.Hour),
			wantErr: true,
		},
		{
			name:          "one hour before spring transition",
			start:         transitionTimes[0].Add(-time.Hour).In(loc),
			wantErr:       false,
			wantTransTime: transitionTimes[0].In(loc),
		},
		{
			name:          "one hour before fall transition",
			start:         transitionTimes[1].Add(-time.Hour).In(loc),
			wantErr:       false,
			wantTransTime: transitionTimes[1].In(loc),
		},
		{
			name:          "one second before spring transition",
			start:         transitionTimes[0].Add(-time.Second).In(loc),
			wantErr:       false,
			wantTransTime: transitionTimes[0].In(loc),
		},
		{
			name:          "one second before fall transition",
			start:         transitionTimes[1].Add(-time.Second).In(loc),
			wantErr:       false,
			wantTransTime: transitionTimes[1].In(loc),
		},
		{
			name:    "no transition at spring transition time",
			start:   transitionTimes[0].In(loc),
			wantErr: true,
		},
		{
			name:    "no transition at fall transition time",
			start:   transitionTimes[1].In(loc),
			wantErr: true,
		},
		{
			name:    "no transition one hour after spring transition",
			start:   transitionTimes[0].Add(time.Hour),
			wantErr: true,
		},
		{
			name:    "no transition one hour after fall transition",
			start:   transitionTimes[1].Add(time.Hour),
			wantErr: true,
		},
		{
			name:    "no transition 5 weeks before spring transition",
			start:   transitionTimes[0].Add(-time.Hour * 24 * 7 * 5),
			wantErr: true,
		},
		{
			name:    "no transition 5 weeks before fall transition",
			start:   transitionTimes[1].Add(-time.Hour * 24 * 7 * 5),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findNextTimezoneTransition(tt.start)

			if tt.wantErr {
				if err == nil {
					t.Errorf("findNextTimezoneTransition() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("findNextTimezoneTransition() unexpected error: %v", err)
				return
			}

			if !got.Equal(tt.wantTransTime) {
				t.Errorf("findNextTimezoneTransition() = %v, want %v", got, tt.wantTransTime)
			}
		})
	}
}
