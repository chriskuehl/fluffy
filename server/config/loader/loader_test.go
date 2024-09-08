package loader

import (
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/testfunc"
)

var configWithEverything = []byte(`
branding = "foo"
custom_footer_html = "<p>foo</p>"
abuse_contact_email = "abuse@foo.com"
max_upload_bytes = 123
max_multipart_memory_bytes = 456
home_url = "http://foo.com"
object_url_pattern = "http://i.foo.com/o/:path:"
html_url_pattern = "http://i.foo.com/h/:path:"
forbidden_file_extensions = ["foo", "bar"]
host = "192.168.1.100"
port = 5555

[filesystem_storage_backend]
object_root = "/tmp/objects"
html_root = "/tmp/html"
`)

func TestLoadConfigTOMLEmptyFile(t *testing.T) {
	configPath := t.TempDir() + "/config.toml"
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	conf := testfunc.NewConfig()
	if err := LoadConfigTOML(conf, configPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	errs := conf.Validate()
	if len(errs) != 0 {
		t.Fatalf("config validation failed: %v", errs)
	}
}

func TestLoadConfigTOMLWithEverything(t *testing.T) {
	configPath := t.TempDir() + "/config.toml"
	if err := os.WriteFile(configPath, configWithEverything, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	conf := testfunc.NewConfig()
	if err := LoadConfigTOML(conf, configPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	errs := conf.Validate()
	if len(errs) != 0 {
		t.Fatalf("config validation failed: %v", errs)
	}

	want := &config.Config{
		Branding:                "foo",
		CustomFooterHTML:        "<p>foo</p>",
		AbuseContactEmail:       "abuse@foo.com",
		MaxUploadBytes:          123,
		MaxMultipartMemoryBytes: 456,
		HomeURL:                 &url.URL{Scheme: "http", Host: "foo.com"},
		ObjectURLPattern:        &url.URL{Scheme: "http", Host: "i.foo.com", Path: "/o/:path:"},
		HTMLURLPattern:          &url.URL{Scheme: "http", Host: "i.foo.com", Path: "/h/:path:"},
		ForbiddenFileExtensions: map[string]struct{}{"foo": {}, "bar": {}},
		Host:                    "192.168.1.100",
		Port:                    5555,
		StorageBackend: &storage.FilesystemBackend{
			ObjectRoot: "/tmp/objects",
			HTMLRoot:   "/tmp/html",
		},
		Version: "(test)",
	}
	if diff := cmp.Diff(want, conf); diff != "" {
		t.Fatalf("config mismatch (-want +got):\n%s", diff)
	}
}

func TestRoundtripDumpLoadConfigTOML(t *testing.T) {
	configPath := t.TempDir() + "/config.toml"
	if err := os.WriteFile(configPath, configWithEverything, 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	conf := testfunc.NewConfig()
	if err := LoadConfigTOML(conf, configPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	errs := conf.Validate()
	if len(errs) != 0 {
		t.Fatalf("config validation failed: %v", errs)
	}

	newConfigPath := t.TempDir() + "/new_config.toml"
	configText, err := DumpConfigTOML(conf)
	if err != nil {
		t.Fatalf("failed to dump config: %v", err)
	}
	if err := os.WriteFile(newConfigPath, []byte(configText), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	newConf := testfunc.NewConfig()
	if err := LoadConfigTOML(newConf, newConfigPath); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	errs = newConf.Validate()
	if len(errs) != 0 {
		t.Fatalf("config validation failed: %v", errs)
	}

	if diff := cmp.Diff(conf, newConf); diff != "" {
		t.Fatalf("config mismatch (-want +got):\n%s", diff)
	}
}
