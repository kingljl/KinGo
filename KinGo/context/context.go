package context

import (
	"net/http"
	"mime"
	"strings"
)

type Context struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Input          *YafInput
	Output         *YafOutput
}




func NewContext( w http.ResponseWriter , r *http.Request , inp *YafInput , out  *YafOutput ) *Context {
	return &Context{
		Request : r,
		ResponseWriter : w ,
		Input  : inp,
		Output  :  out,
	}
}


// Redirect does redirection to localurl with http header status code.
// It sends http response header directly.
func (ctx *Context) Redirect(status int, localurl string) {
	ctx.Output.Header("Location", localurl)
	ctx.Output.SetStatus(status)
}

// Abort stops this request.
// if middleware.ErrorMaps exists, panic body.
// if middleware.HTTPExceptionMaps exists, panic HTTPException struct with status and body string.
func (ctx *Context) Abort(status int, body string) {
	ctx.Output.SetStatus(status)
	// last panic user string
	panic(body)
}

func (ctx *Context) NotModified() {
	ctx.ResponseWriter.WriteHeader(304)
}

func (ctx *Context) NotFound(message string) {
	ctx.Error(404, message)
}
func (ctx *Context) WriteString(content []byte) {
	ctx.ResponseWriter.Write( content );
}

func (ctx *Context) Error(code int, message string) {
	ctx.ResponseWriter.WriteHeader(code)
	ctx.ResponseWriter.Write([]byte(message))
}

func (ctx *Context) AddHeader(hdr string, val string) {
	ctx.Output.AddHeader(hdr, val)
}

//func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
func (ctx *Context) SetHeader(hdr string, val string) {
	ctx.Output.Header(hdr, val);
}
func (ctx *Context) ContentType(typ string) {
	ext := typ
	if !strings.HasPrefix(typ, ".") {
		ext = "." + typ
	}
	ctype := mime.TypeByExtension(ext)
	if ctype != "" {
		ctx.ResponseWriter.Header().Set("Content-Type", ctype)
	} else {
		ctx.ResponseWriter.Header().Set("Content-Type", typ)
	}
}
