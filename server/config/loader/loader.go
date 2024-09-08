package loader

import (
	"fmt"
	"html/template"
	"net/url"

	"github.com/BurntSushi/toml"
	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/storage"
)

type filesystemStorageBackend struct {
	ObjectRoot string `toml:"object_root"`
	HTMLRoot   string `toml:"html_root"`
}

type configFile struct {
	Branding                string   `toml:"branding"`
	CustomFooterHTML        string   `toml:"custom_footer_html"`
	AbuseContactEmail       string   `toml:"abuse_contact_email"`
	MaxUploadBytes          int64    `toml:"max_upload_bytes"`
	MaxMultipartMemoryBytes int64    `toml:"max_multipart_memory_bytes"`
	HomeURL                 string   `toml:"home_url"`
	ObjectURLPattern        string   `toml:"object_url_pattern"`
	HTMLURLPattern          string   `toml:"html_url_pattern"`
	ForbiddenFileExtensions []string `toml:"forbidden_file_extensions"`
	Host                    string   `toml:"host"`
	Port                    uint     `toml:"port"`

	FilesystemStorageBackend *filesystemStorageBackend `toml:"filesystem_storage_backend"`
}

func LoadConfigTOML(conf *config.Config, path string) error {
	var cfg configFile
	md, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return fmt.Errorf("decoding config: %w", err)
	}
	if len(md.Undecoded()) > 0 {
		return fmt.Errorf("unknown keys in config: %v", md.Undecoded())
	}
	if cfg.Branding != "" {
		conf.Branding = cfg.Branding
	}
	if cfg.CustomFooterHTML != "" {
		conf.CustomFooterHTML = template.HTML(cfg.CustomFooterHTML)
	}
	if cfg.AbuseContactEmail != "" {
		conf.AbuseContactEmail = cfg.AbuseContactEmail
	}
	if cfg.MaxUploadBytes != 0 {
		conf.MaxUploadBytes = cfg.MaxUploadBytes
	}
	if cfg.MaxMultipartMemoryBytes != 0 {
		conf.MaxMultipartMemoryBytes = cfg.MaxMultipartMemoryBytes
	}
	if cfg.HomeURL != "" {
		u, err := url.ParseRequestURI(cfg.HomeURL)
		if err != nil {
			return fmt.Errorf("parsing HomeURL: %w", err)
		}
		conf.HomeURL = u
	}
	if cfg.ObjectURLPattern != "" {
		u, err := url.ParseRequestURI(cfg.ObjectURLPattern)
		if err != nil {
			return fmt.Errorf("parsing ObjectURLPattern: %w", err)
		}
		conf.ObjectURLPattern = u
	}
	if cfg.HTMLURLPattern != "" {
		u, err := url.ParseRequestURI(cfg.HTMLURLPattern)
		if err != nil {
			return fmt.Errorf("parsing HTMLURLPattern: %w", err)
		}
		conf.HTMLURLPattern = u
	}
	for _, ext := range cfg.ForbiddenFileExtensions {
		conf.ForbiddenFileExtensions[ext] = struct{}{}
	}
	if cfg.Host != "" {
		conf.Host = cfg.Host
	}
	if cfg.Port != 0 {
		conf.Port = cfg.Port
	}
	if cfg.FilesystemStorageBackend != nil {
		conf.StorageBackend = &storage.FilesystemBackend{
			ObjectRoot: cfg.FilesystemStorageBackend.ObjectRoot,
			HTMLRoot:   cfg.FilesystemStorageBackend.HTMLRoot,
		}
	}
	return nil
}

func DumpConfigTOML(conf *config.Config) (string, error) {
	cfg := configFile{
		Branding:                conf.Branding,
		CustomFooterHTML:        string(conf.CustomFooterHTML),
		AbuseContactEmail:       conf.AbuseContactEmail,
		MaxUploadBytes:          conf.MaxUploadBytes,
		MaxMultipartMemoryBytes: conf.MaxMultipartMemoryBytes,
		HomeURL:                 conf.HomeURL.String(),
		ObjectURLPattern:        conf.ObjectURLPattern.String(),
		HTMLURLPattern:          conf.HTMLURLPattern.String(),
		ForbiddenFileExtensions: make([]string, 0, len(conf.ForbiddenFileExtensions)),
		Host:                    conf.Host,
		Port:                    conf.Port,
	}
	for ext := range conf.ForbiddenFileExtensions {
		cfg.ForbiddenFileExtensions = append(cfg.ForbiddenFileExtensions, ext)
	}
	if fs, ok := conf.StorageBackend.(*storage.FilesystemBackend); ok {
		cfg.FilesystemStorageBackend = &filesystemStorageBackend{
			ObjectRoot: fs.ObjectRoot,
			HTMLRoot:   fs.HTMLRoot,
		}
	}
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshaling config: %w", err)
	}
	return string(buf), nil
}
