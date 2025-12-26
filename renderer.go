// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2015 LabStack LLC and Echo contributors

package echo

import "io"

// Renderer is the interface that wraps the Render function.
type Renderer interface {
	Render(c *Context, w io.Writer, templateName string, data any) error
}

// TemplateRenderer is helper to ease creating renderers for `html/template` and `text/template` packages.
// Example usage:
//
//		e.Renderer = &echo.TemplateRenderer{
//			Template: template.Must(template.ParseGlob("templates/*.html")),
//		}
//
//	  e.Renderer = &echo.TemplateRenderer{
//			Template: template.Must(template.New("hello").Parse("Hello, {{.}}!")),
//		}
type TemplateRenderer struct {
	Template interface {
		ExecuteTemplate(wr io.Writer, name string, data any) error
	}
}

// Render renders the template with given data.
func (t *TemplateRenderer) Render(c *Context, w io.Writer, name string, data any) error {
	return t.Template.ExecuteTemplate(w, name, data)
}
