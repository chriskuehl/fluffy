package assets_test

import (
	"testing"

	"github.com/chriskuehl/fluffy/server/assets"
	"github.com/chriskuehl/fluffy/testfunc"
)

func TestAssetURLDev(t *testing.T) {
	conf := testfunc.NewConfig()
	conf.DevMode = true

	got, err := assets.AssetURL(conf, "img/favicon.ico")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "http://localhost:8080/dev/static/img/favicon.ico"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAssetURLProd(t *testing.T) {
	conf := testfunc.NewConfig()
	conf.DevMode = false

	got, err := assets.AssetURL(conf, "img/favicon.ico")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "http://localhost:8080/dev/object/static/5b707398fe549635b8794ac8e73db6938dd7b6b7a28b339296bde1b0fdec764b/img/favicon.ico"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
