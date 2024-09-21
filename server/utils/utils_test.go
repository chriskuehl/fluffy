package utils_test

import (
	"strconv"
	"testing"

	"github.com/chriskuehl/fluffy/server/utils"
)

func TestHumanFileExtension(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{
			name: "no extension",
			in:   "file",
			want: "",
		},
		{
			name: "regular extension",
			in:   "file.txt",
			want: ".txt",
		},
		{
			name: "wrapped extension only",
			in:   "file.gz",
			want: ".gz",
		},
		{
			name: "wrapped extension after regular extension",
			in:   "file.tar.gz",
			want: ".tar.gz",
		},
		{
			name: "multiple wrapped extensions",
			in:   "file.tar.gz.bz2",
			want: ".tar.gz.bz2",
		},
		{
			name: "multiple wrapped extensions with a regular extension",
			in:   "file.txt.tar.gz.bz2",
			want: ".tar.gz.bz2",
		},
		{
			// Kind of nonsense, just making sure it doesn't remove more than it should.
			name: "wrapped extensions before regular extension",
			in:   "file.tar.gz.txt",
			want: ".txt",
		},
		{
			name: ". only",
			in:   ".",
			want: "",
		},
		{
			name: "multiple wrapped extensions with empty extensions",
			in:   "file.txt.tar.gz....bz2",
			want: ".tar.gz.bz2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.HumanFileExtension(tt.in); got != tt.want {
				t.Errorf("%q: got %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := map[int64]string{
		0: "things",
		1: "thing",
		2: "things",
	}
	for count, want := range tests {
		t.Run(strconv.FormatInt(count, 10), func(t *testing.T) {
			if got := utils.Pluralize("thing", count); got != want {
				t.Errorf("got Pluralize(%q, %d) = %v, want %v", "thing", count, got, want)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := map[int64]string{
		0:                                "0 bytes",
		1:                                "1 byte",
		1023:                             "1023 bytes",
		1024:                             "1.0 KiB",
		1024 * 1024:                      "1.0 MiB",
		1024*1024 - 1:                    "1023.9 KiB",
		1024*1024*1024 - 1:               "1023.9 MiB",
		1024 * 1024 * 1024:               "1.0 GiB",
		3*1024*1024*1024 + 717*1024*1024: "3.7 GiB",
	}
	for bytes, want := range tests {
		t.Run(strconv.FormatInt(bytes, 10), func(t *testing.T) {
			if got := utils.FormatBytes(bytes); got != want {
				t.Errorf("got FormatBytes(%d) = %v, want %v", bytes, got, want)
			}
		})
	}
}
