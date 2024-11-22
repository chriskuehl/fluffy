package testfunc

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func KeyFromURL(url string) string {
	return url[strings.LastIndex(url, "/")+1:]
}

// normalizeSpacing normalizes the spacing in a string similar to how a web browser would,
// collapsing multiple spaces into a single space, ignoring newlines, and trimming leading/trailing
// spaces.
func normalizeSpacing(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

type CanonicalizedLinks string

func CanonicalizeLinks(links []*url.URL) CanonicalizedLinks {
	// Sort the links first to ensure a consistent order.
	urls := make([]string, len(links))
	for i, link := range links {
		urls[i] = link.String()
	}
	sort.Strings(urls)
	return CanonicalizedLinks(strings.Join(urls, " :: "))
}

type ParsedUploadDetails struct {
	MetadataURL string
	Files       map[string]*ParsedUploadDetailsFile
}

type ParsedUploadDetailsFile struct {
	Icon              string
	Size              string
	DirectLinkFileKey string
	PasteLinkHTMLKey  string
}

func ParseUploadDetails(html string) (*ParsedUploadDetails, error) {
	gq, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	ret := &ParsedUploadDetails{
		MetadataURL: gq.Find("meta[name='fluffy:metadata-url']").AttrOr("content", ""),
		Files:       make(map[string]*ParsedUploadDetailsFile),
	}
	gq.Find(".file-holder .file").Each(func(_ int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".filename").Text())
		ret.Files[name] = &ParsedUploadDetailsFile{
			Icon:              KeyFromURL(s.Find(".filename img").AttrOr("src", "")),
			Size:              strings.TrimSpace(s.Find(".filesize").Text()),
			DirectLinkFileKey: KeyFromURL(s.Find(".download").AttrOr("href", "")),
			PasteLinkHTMLKey:  KeyFromURL(s.Find(".view-paste").AttrOr("href", "")),
		}
	})
	return ret, nil
}

type ParsedPaste struct {
	MetadataURL      string
	RawURL           string
	DefaultStyleName string
	// e.g. "5 lines of Python"
	ToolbarInfoLine string
	CopyAndEditText string
	HasDiffButtons  bool
	Texts           int
}

func ParsePaste(html string) (*ParsedPaste, error) {
	gq, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	ret := &ParsedPaste{
		MetadataURL: gq.Find("meta[name='fluffy:metadata-url']").AttrOr("content", ""),
		RawURL:      gq.Find("#raw-text").AttrOr("href", ""),
		DefaultStyleName: strings.TrimPrefix(
			gq.Find("#highlightContainer").AttrOr("class", ""),
			"style-",
		),
		ToolbarInfoLine: normalizeSpacing(gq.Find(".paste-toolbar .info").Text()),
		CopyAndEditText: gq.Find("#copy-and-edit").AttrOr("value", ""),
		HasDiffButtons:  gq.Find("#diff-setting").Length() > 0,
		Texts:           gq.Find(".text-container > .text").Length(),
	}
	return ret, nil
}
