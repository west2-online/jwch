package jwch

import (
	"regexp"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Safely extract data from html node by expression
func safeExtractionFirst(node *html.Node, expr string) string {
	res := htmlquery.FindOne(node, expr)

	if res == nil {
		return htmlquery.OutputHTML(node, false)
	}

	return htmlquery.OutputHTML(res, false)
}

// Safely extract data from html node by expression, return the index-th element(if index is out of range, return the last element)
func safeExtractionValue(node *html.Node, expr string, value string, index int) string {
	res := htmlquery.Find(node, expr)

	if res == nil {
		return ""
	}

	if len(res) <= index {
		return htmlquery.SelectAttr(res[len(res)-1], value)
	}

	return htmlquery.SelectAttr(res[index], value)
}

// Safely extract data by regex
func safeExtractRegex(regex, str string) string {
	res := regexp.MustCompile(regex).FindStringSubmatch(str)

	if len(res) < 2 {
		return ""
	}

	return res[1]
}

func safeExtractHTMLFirst(node *html.Node, expr string) string {
	res := htmlquery.FindOne(node, expr)

	if res == nil {
		return ""
	}

	return htmlquery.OutputHTML(res, false)

}
