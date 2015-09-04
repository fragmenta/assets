package assets

import (
	"fmt"
	"html/template"
)

// Convert a set of group names to one style link tag (production)
func (c *Collection) StyleLink(names ...string) template.HTML {
	var html template.HTML

	// Iterate through names, setting up links for each
	// we link to groups if we have them, else we fall back to normal links
	for _, name := range names {
		g := c.Group(name)
		if g.stylehash != "" {
			if c.serveCompiled {
				html = html + StyleLink(g.StyleName())
			} else {
				for _, f := range g.Styles() {
					html = html + StyleLink(f.name) + template.HTML("\n")
				}
			}
		} else {
			html = html + StyleLink(name)
		}

	}

	return html
}

// Convert a set of group names to one style link tag (production)
func (c *Collection) ScriptLink(names ...string) template.HTML {
	var html template.HTML

	// Iterate through names, setting up links for each
	// we link to groups if we have them, else we fall back to normal links
	for _, name := range names {
		g := c.Group(name)
		if g.stylehash != "" {
			if c.serveCompiled {
				html = html + ScriptLink(g.ScriptName())
			} else {
				for _, f := range g.Scripts() {
					html = html + ScriptLink(f.name) + template.HTML("\n")
				}
			}
		} else {
			html = html + ScriptLink(name)
		}

	}

	return html
}

func StyleLink(name string) template.HTML {
	return template.HTML(fmt.Sprintf("<link href=\"/assets/styles/%s\" media=\"all\" rel=\"stylesheet\" type=\"text/css\" />", template.URLQueryEscaper(name)))
}

// Script inserts a script tag
func ScriptLink(name string) template.HTML {
	return template.HTML(fmt.Sprintf("<script src=\"/assets/scripts/%s\" type=\"text/javascript\"></script>", template.URLQueryEscaper(name)))
}
