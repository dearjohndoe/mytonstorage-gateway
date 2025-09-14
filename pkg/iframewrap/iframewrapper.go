// Package iframewrap wrap and return a complete HTML page (parent page) that embeds the user's HTML
// inside a tightly sandboxed iframe.
package iframewrap

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type Options struct {
	AllowScripts bool
	AllowForms   bool
}

func WrapHTML(userHTML string, o Options) (string, error) {
	iframeAttrs := make([]string, 0, 3)
	if o.AllowScripts {
		iframeAttrs = append(iframeAttrs, "allow-scripts")
	}
	if o.AllowForms {
		iframeAttrs = append(iframeAttrs, "allow-forms")
	}

	sandbox := strings.Join(iframeAttrs, " ")

	esc := html.EscapeString(userHTML)
	srcAttr := fmt.Sprintf("srcdoc=\"%s\"", esc)

	parent := fmt.Sprintf(parentTemplate, sandbox, srcAttr)

	return parent, nil
}
