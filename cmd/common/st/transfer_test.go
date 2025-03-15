package st

import (
	"testing"
)

func Test_validateCommonFlags(t *testing.T) {
	tests := []struct {
		name           string
		source, target string
		concurrency    int
		wantErr        bool
	}{
		{"source empty", "", "abc", 1, true},
		{"source all", "all", "abc", 1, true},
		{"target all", "abc", "all", 1, true},
		{"same storages", "abc", "abc", 1, true},
		{"concurrency < 1", "source", "target", 0, true},
		{"valid", "source", "target", 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transferSourceStorage = tt.source
			targetStorage = tt.target
			transferConcurrency = tt.concurrency
			if err := validateCommonFlags(); (err != nil) != tt.wantErr {
				t.Errorf("validateCommonFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
