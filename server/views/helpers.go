package views

import (
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/chriskuehl/fluffy/server/assets"
	"github.com/chriskuehl/fluffy/server/config"
)

func iconExtensionsJS(conf *config.Config) (template.JS, error) {
	extensionToURL := make(map[string]string)
	for ext, _ := range conf.Assets.MIMEExtensions {
		url, err := assets.AssetURL(conf, "img/mime/small/"+ext+".png")
		if err != nil {
			return "", fmt.Errorf("failed to get asset URL for %q: %w", ext, err)
		}
		extensionToURL[ext] = url
	}
	json, err := json.Marshal(extensionToURL)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mime extensions to JSON: %w", err)
	}
	return template.JS(json), nil
}
