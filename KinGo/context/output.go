package context

type YafOutput struct {
	Context    *Context
	Status     int
	EnableGzip bool
}

// NewOutput returns new BeegoOutput.
// it contains nothing now.
func NewOutput() *YafOutput {
	return &YafOutput{}
}
// Header sets response header item string via given key.
func (output *YafOutput) Header(key, val string) {
	output.Context.ResponseWriter.Header().Set(key, val)
}

// Header sets response header item string via given key.
func (output *YafOutput) AddHeader(key, val string) {
	output.Context.ResponseWriter.Header().Add(key, val);
}

// SetStatus sets response status code.
// It writes response header directly.
func (output *YafOutput) SetStatus(status int) {
	output.Context.ResponseWriter.WriteHeader(status)
	output.Status = status
}

// IsOk returns boolean of this request runs well.
// HTTP 200 means ok.
func (output *YafOutput) IsOk(status int) bool {
	return output.Status == 200
}
