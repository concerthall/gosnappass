// Package view contains view functions used to render the front end of the application.
//
// Views for this application take on a simple format. Each view template must contain:
//   - a template called "base", shared across all views.
//   - a template called "content", rendering the content specific to a single view context.
//   - a template called "js" which contains base javascript imports, shared across all templates
//   - a template called "contentjs" which allows single-view js to be added after the base js
//
// Template execution always occurs by executing "base", which subsequently renders the
// other components.
package view

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/concerthall/gosnappass/internal/config"
	"github.com/concerthall/gosnappass/internal/embedded"
)

var (
	appHomeLinkRef          string = "/"
	indexTemplate           *template.Template
	confirmationTemplate    *template.Template
	previewPasswordTemplate *template.Template
	expiredTemplate         *template.Template
	showPasswordTemplate    *template.Template
)

// LoadTemplates reaches into the filesystem and loads the appropriate base and
// content used for the provided views.
func LoadTemplates() error {
	var err error
	if v, isSet := os.LookupEnv(config.EnvURLPrefix); isSet {
		appHomeLinkRef = v
	}

	if indexTemplate, err = template.ParseFS(embedded.Templates, "templates/base.html", "templates/set_password.html"); err != nil {
		return err
	}

	if confirmationTemplate, err = template.ParseFS(embedded.Templates, "templates/base.html", "templates/confirm.html"); err != nil {
		return err
	}

	if previewPasswordTemplate, err = template.ParseFS(embedded.Templates, "templates/base.html", "templates/preview.html"); err != nil {
		return err
	}

	if expiredTemplate, err = template.ParseFS(embedded.Templates, "templates/base.html", "templates/expired.html"); err != nil {
		return err
	}

	if showPasswordTemplate, err = template.ParseFS(embedded.Templates, "templates/base.html", "templates/password.html"); err != nil {
		return err
	}

	return nil
}

// init executes template loading.
// TODO: This panics if there are errors. While this may work for now
// we may want to do this instead at application initialization.
func init() {
	if err := LoadTemplates(); err != nil {
		panic(err)
	}
}

// bufferedWriteTo will execute the template with data to a byte buffer, and
// write that to w if no error occurs in writing.
func bufferedWriteTo(w http.ResponseWriter, tmpl *template.Template, data any) error {
	b := []byte{}
	buf := bytes.NewBuffer(b)
	if err := tmpl.ExecuteTemplate(buf, "base", data); err != nil {
		return fmt.Errorf("unable to execute template")
	}

	fmt.Fprintln(w, buf)
	return nil
}

func Index(w http.ResponseWriter) error {
	// TODO: fix redundant AppHomeLinkRef usage across all views.
	return bufferedWriteTo(w, indexTemplate, map[string]string{"AppHomeLinkRef": appHomeLinkRef})
}

func Confirm(w http.ResponseWriter, link string) error {
	return bufferedWriteTo(w, confirmationTemplate, map[string]string{"AppHomeLinkRef": appHomeLinkRef, "PasswordLink": link})
}

func PreviewPassword(w http.ResponseWriter) error {
	return bufferedWriteTo(w, previewPasswordTemplate, map[string]string{"AppHomeLinkRef": appHomeLinkRef})
}

// CredentialExpiredOrNotFound is the view corresponding with serving HTTP 404 responses. If the
// view rendering fails a buffered write, this view falls back to a plain text resposne.
func CredentialExpiredOrNotFound(w http.ResponseWriter) {
	if err := bufferedWriteTo(w, expiredTemplate, map[string]string{"AppHomeLinkRef": appHomeLinkRef}); err != nil {
		fmt.Fprintln(w, "We didn't find it (404)")
	}
}

func ShowPassword(w http.ResponseWriter, password string) error {
	return bufferedWriteTo(w, showPasswordTemplate, map[string]string{"AppHomeLinkRef": appHomeLinkRef, "Password": password})
}
