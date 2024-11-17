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

type ParsedUploadDetailsFile struct {
	Name              string
	Icon              string
	Size              string
	DirectLinkFileKey string
	PasteLinkHTMLKey  string
}

type ParsedUploadDetails struct {
	Files map[string]*ParsedUploadDetailsFile
}

func ParseUploadDetails(html string) (*ParsedUploadDetails, error) {
	gq, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	ret := &ParsedUploadDetails{
		Files: make(map[string]*ParsedUploadDetailsFile),
	}
	gq.Find(".file-holder .file").Each(func(_ int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".filename").Text())
		icon, _ := s.Find(".filename img").Attr("src")
		size := strings.TrimSpace(s.Find(".filesize").Text())
		directLink, _ := s.Find(".download").Attr("href")
		pasteLink, _ := s.Find(".view-paste").Attr("href")
		ret.Files[name] = &ParsedUploadDetailsFile{
			Name:              name,
			Icon:              KeyFromURL(icon),
			Size:              size,
			DirectLinkFileKey: KeyFromURL(directLink),
			PasteLinkHTMLKey:  KeyFromURL(pasteLink),
		}
	})
	return ret, nil
}
