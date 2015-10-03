package KinGo

import (
	"fmt"
	"net/http"
	"strconv"
	"path"
	"encoding/json"
	"KinGo/context"
	"encoding/xml"
	"bytes"
	"io/ioutil"
)

const (
	applicationJson = "application/json"
	applicationXml  = "application/xml"
	textXml         = "text/xml"
)

type ControllerComments struct {
	Method           string
	Router           string
	AllowHTTPMethods []string
	Params           []map[string]string
}

// Controller defines some basic http request handler operations, such as
// http context, template and view, session and xsrf.
type Controller  struct {
	Ctx            *context.Context
	Data           map[interface{}]interface{}
	controllerName string
	actionName     string
	TplNames       string
	Layout         string
	LayoutSections map[string]string // the key is the section name and the value is the template name
	TplExt         string
	_xsrf_token    string
	gotofunc       string
	XSRFExpire     int
	AppController  interface{}
	EnableRender   bool
	EnableXSRF     bool
	methodMapping  map[string]func() //method:routertree
}

// ControllerInterface is an interface to uniform all controller handler.
type ControllerInterface interface {
	Init( ctx *context.Context , controllerName, actionName string, app interface{})
	Prepare()
	Get()
	Post()
	Delete()
	Put()
	Head()
	Patch()
	Options()
	Finish()
}

// Init generates default values of controller operations.
func (c *Controller) Init( ctx *context.Context ,  controllerName, actionName string, app interface{} ) {
	c.Layout = ""
	c.TplNames = ""
	c.controllerName = controllerName
	c.actionName = actionName
	c.Ctx = ctx
	c.TplExt = "tpl"
	c.AppController = app
	c.EnableRender = true
	c.EnableXSRF = true
	c.Data = ctx.Input.Data
	c.methodMapping = make(map[string]func())
}

// Prepare runs after Init before request function execution.
func (c *Controller) Prepare() {

}

// Finish runs after request function execution.
func (c *Controller) Finish() {

}

// Get adds a request function to handle GET request.
func (c *Controller) Get() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Post adds a request function to handle POST request.
func (c *Controller) Post() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Delete adds a request function to handle DELETE request.
func (c *Controller) Delete() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Put adds a request function to handle PUT request.
func (c *Controller) Put() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Head adds a request function to handle HEAD request.
func (c *Controller) Head() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Patch adds a request function to handle PATCH request.
func (c *Controller) Patch() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

// Options adds a request function to handle OPTIONS request.
func (c *Controller) Options() {
	http.Error(c.Ctx.ResponseWriter, "Method Not Allowed", 405)
}

func (c *Controller) Render(contentType string, data []byte) {
	c.Ctx.SetHeader("Content-Length", strconv.Itoa(len(data)))
	c.Ctx.ContentType(contentType)
	c.Ctx.ResponseWriter.Write(data)
}

func (c *Controller) RenderHtml(content string) {
	c.Render("html", []byte(content))
}

func (c *Controller) RenderText(content string) {
	c.Render("txt", []byte(content))
}

func (c *Controller) RenderJson(data interface{}) {
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Render("json", content)
}

func (c *Controller) RenderJQueryCallback(jsoncallback string, data interface{}) {
	var content []byte
	switch data.(type) {
	case string:
		content = []byte(data.(string))
	case []byte:
		content = data.([]byte)
	default:
		var err error
		content, err = json.Marshal(data)
		if err != nil {
			http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	bjson := []byte(jsoncallback)
	bjson = append(bjson, '(')
	bjson = append(bjson, content...)
	bjson = append(bjson, ')')
	c.Render("json", bjson)
}

func (c *Controller) RenderXml(data interface{}) {
	content, err := xml.Marshal(data)
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Render("xml", content)
}

func (c *Controller) RenderTemplate(contentType ...string) {
	if c.TplNames == "" {
		c.TplNames = c.controllerName + "/" + c.actionName + "." + c.TplExt
	}
	_, file := path.Split(c.TplNames)
	subdir := path.Dir(c.TplNames)
	ibytes := bytes.NewBufferString("")

	fmt.Println( subdir )
	t := kgo.TemplateRegistor.Templates[subdir]
	fmt.Println(  kgo.TemplateRegistor.Templates )
	if t == nil {
		http.Error(c.Ctx.ResponseWriter, "Internal Server Error (template not exist)", http.StatusInternalServerError)
		return
	}
	err := t.ExecuteTemplate(ibytes, file, c.Data)
	if err != nil {
//		log.Println("template Execute err:", err)
		http.Error(c.Ctx.ResponseWriter, "Internal Server Error (ExecuteTemplate faild)", http.StatusInternalServerError)
		return
	}
	icontent, _ := ioutil.ReadAll(ibytes)
	if len(contentType) > 0 {
		c.Render(contentType[0], icontent)
	} else {
		c.Render("html", icontent)
	}
}
