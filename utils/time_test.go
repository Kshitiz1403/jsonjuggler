package utils

import (
	"testing"
	"time"
)

func TestParseISODuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "simple duration with hours",
			input:   "PT2H",
			want:    2 * time.Hour,
			wantErr: false,
		},
		{
			name:    "complex duration with days, hours, minutes",
			input:   "P2DT3H30M",
			want:    (2 * 24 * time.Hour) + (3 * time.Hour) + (30 * time.Minute),
			wantErr: false,
		},
		{
			name:    "duration with minutes and seconds",
			input:   "PT15M30S",
			want:    (15 * time.Minute) + (30 * time.Second),
			wantErr: false,
		},
		{
			name:    "duration with weeks",
			input:   "P2W",
			want:    14 * 24 * time.Hour,
			wantErr: false,
		},
		{
			name:    "zero duration",
			input:   "PT0S",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid duration format",
			input:   "2h30m",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid ISO duration",
			input:   "PTS",
			want:    0,
			wantErr: true,
		},
		{
			name:    "millisecond duration",
			input:   "PT2.5S",
			want:    2500 * time.Millisecond,
			wantErr: false,
		},
		{
			name:    "millisecond duration",
			input:   "PT0.5S",
			want:    500 * time.Millisecond,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseISODuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseISODuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseISODuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
