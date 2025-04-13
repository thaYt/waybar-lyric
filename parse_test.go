package main

import (
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		ts      string
		want    time.Duration
		wantErr bool
	}{
		{
			ts:      "20:11.50",
			want:    20*time.Minute + 11*time.Second + 500*time.Millisecond,
			wantErr: false,
		},
		{
			ts:      "20:50",
			want:    20*time.Minute + 50*time.Second,
			wantErr: false,
		},
		{
			ts:      "11:20:50",
			want:    11*time.Hour + 20*time.Minute + 50*time.Second,
			wantErr: false,
		},
		// Single value (seconds only)
		{
			ts:      "45",
			want:    45 * time.Second,
			wantErr: false,
		},
		// Decimal seconds
		{
			ts:      "30.5",
			want:    30*time.Second + 500*time.Millisecond,
			wantErr: false,
		},
		// Hours with decimal minutes
		{
			ts:      "1:30.5",
			want:    1*time.Minute + 30*time.Second + 500*time.Millisecond,
			wantErr: false,
		},
		// Hours with decimal minutes
		{
			ts:      "2:45.75",
			want:    2*time.Minute + 45*time.Second + 750*time.Millisecond,
			wantErr: false,
		},
		// Full format with decimal seconds
		{
			ts:      "1:23:45.5",
			want:    1*time.Hour + 23*time.Minute + 45*time.Second + 500*time.Millisecond,
			wantErr: false,
		},
		// Spaces in the timestamp
		{
			ts:      " 10:30 ",
			want:    10*time.Minute + 30*time.Second,
			wantErr: false,
		},
		// Zero values
		{
			ts:      "0:0:0",
			want:    0,
			wantErr: false,
		},
		// Error cases
		{
			ts:      "",
			want:    0,
			wantErr: true,
		},
		{
			ts:      "abc",
			want:    0,
			wantErr: true,
		},
		{
			ts:      "1:2:3:4",
			want:    0,
			wantErr: true,
		},
		{
			ts:      "1:ab:3",
			want:    0,
			wantErr: true,
		},
		{
			ts:      "1:-30:20",
			want:    0,
			wantErr: true,
		},
		{
			ts:      ":30",
			want:    0,
			wantErr: true,
		},
		{
			ts:      "1::",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run("Parsing time "+tt.ts, func(t *testing.T) {
			got, gotErr := ParseTimestamp(tt.ts)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseTimestamp() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseTimestamp() succeeded unexpectedly")
			}
			t.Logf("parseTimestamp() = %v, want %v", got, tt.want)
		})
	}
}
