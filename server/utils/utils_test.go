package utils_test

import (
	"strconv"
	"testing"

	"github.com/chriskuehl/fluffy/server/utils"
)

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
