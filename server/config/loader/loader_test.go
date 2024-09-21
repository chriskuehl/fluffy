package loader

import (
	"net/url"
	"os"
	"testing"
	"time"

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
file_url_pattern = "http://i.foo.com/o/:key:"
html_url_pattern = "http://i.foo.com/h/:key:"
forbidden_file_extensions = ["foo", "bar"]
host = "192.168.1.100"
port = 5555
global_timeout_ms = 5555

[filesystem_storage_backend]
file_root = "/tmp/file"
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

func diffConfig(c1, c2 config.Config) string {
	for _, c := range []*config.Config{&c1, &c2} {
		c.Assets = nil
		c.Templates = nil
	}
	return cmp.Diff(c1, c2)
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
		FileURLPattern:          &url.URL{Scheme: "http", Host: "i.foo.com", Path: "/o/:key:"},
		HTMLURLPattern:          &url.URL{Scheme: "http", Host: "i.foo.com", Path: "/h/:key:"},
		ForbiddenFileExtensions: map[string]struct{}{"foo": {}, "bar": {}},
		Host:                    "192.168.1.100",
		Port:                    5555,
		GlobalTimeout:           5555 * time.Millisecond,
		StorageBackend: &storage.FilesystemBackend{
			FileRoot: "/tmp/file",
			HTMLRoot: "/tmp/html",
		},
		Version: "(test)",
	}
	if diff := diffConfig(*want, *conf); diff != "" {
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

	if diff := diffConfig(*conf, *newConf); diff != "" {
		t.Fatalf("config mismatch (-want +got):\n%s", diff)
	}
}
