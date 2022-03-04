package render

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/justinas/nosurf"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// Render the render type
type Render struct {
	RenderingEngine string
	RootPath        string
	Secure          bool
	Port            string
	ServerName      string
	JetViews        *jet.Set
	Session         *scs.SessionManager
}

// TemplateData data will be sent to templates
type TemplateData struct {
	IsAuthenticated bool
	IntMap          map[string]int
	StringMap       map[string]string
	FloatMap        map[string]float32
	Data            map[string]interface{}
	CSRFToken       string
	Port            string
	ServerName      string
	Secure          bool
	Error           string
	Flash           string
}

// TemplateDefaultData adds default data to TemplateData type
func (re *Render) TemplateDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Secure = re.Secure
	td.ServerName = re.ServerName
	td.Port = re.Port

	/////////////////
	// Add csrf token
	/////////////////
	td.CSRFToken = nosurf.Token(r)

	///////////////////////
	// Add isAuthenticated
	///////////////////////
	// check if userId key in the session
	if re.Session.Exists(r.Context(), "userID") {
		td.IsAuthenticated = true
	}
	////////////////////////////////
	// Add error and flash messages
	////////////////////////////////
	td.Error = re.Session.PopString(r.Context(), "error")
	td.Flash = re.Session.PopString(r.Context(), "flash")

	// return all template data
	return td
}

// Page general render function to render jet or go templates
func (re *Render) Page(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	switch strings.ToLower(re.RenderingEngine) {
	case "go":
		return re.GoPage(w, r, view, data)
	case "jet":
		return re.JetPage(w, r, view, variables, data)
	}
	return nil
}

// GoPage render gohtml templates
func (re *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, data interface{}) error {
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/views/%s.page.gohtml", re.RootPath, view))
	if err != nil {
		return err
	}
	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}
	err = tmpl.Execute(w, &td)
	if err != nil {
		return err
	}
	return nil
}

// JetPage render jet templates
func (re *Render) JetPage(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	var vars jet.VarMap
	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData)
	}
	// template default data
	td = re.TemplateDefaultData(td, r)
	t, err := re.JetViews.GetTemplate(fmt.Sprintf("%s.jet", view))
	if err != nil {
		return err
	}
	if err := t.Execute(w, vars, td); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
