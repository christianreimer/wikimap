package parser

import (
	"encoding/xml"
	"io"
	"regexp"
	"strings"
)

type Article struct {
	ID    uint32
	Title string
	Links []string
}

var linkRegex = regexp.MustCompile(`\[\[([^\]|#]+)(?:[|#][^\]]*)?\]\]`)

func Parse(r io.Reader) ([]Article, error) {
	decoder := xml.NewDecoder(r)
	var articles []Article

	var inPage, inRevision bool
	var current struct {
		title string
		ns    string
		id    uint32
		text  string
	}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "page":
				inPage = true
				current.title = ""
				current.ns = ""
				current.id = 0
				current.text = ""
			case "revision":
				inRevision = true
			case "title", "ns", "id", "text":
				if !inPage {
					continue
				}
				var content string
				if err := decoder.DecodeElement(&content, &t); err != nil {
					return nil, err
				}
				switch t.Name.Local {
				case "title":
					current.title = content
				case "ns":
					current.ns = content
				case "id":
					if !inRevision {
						var id uint32
						for _, c := range content {
							id = id*10 + uint32(c-'0')
						}
						current.id = id
					}
				case "text":
					current.text = content
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "revision":
				inRevision = false
			case "page":
				inPage = false
				if current.ns == "0" {
					links := extractLinks(current.text)
					articles = append(articles, Article{
						ID:    current.id,
						Title: current.title,
						Links: links,
					})
				}
			}
		}
	}
	return articles, nil
}

func extractLinks(text string) []string {
	matches := linkRegex.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var links []string
	for _, m := range matches {
		target := strings.TrimSpace(m[1])
		// Capitalize first letter to normalize
		if len(target) > 0 {
			target = strings.ToUpper(target[:1]) + target[1:]
		}
		if !seen[target] {
			seen[target] = true
			links = append(links, target)
		}
	}
	return links
}
