package KinGo

import (
	"fmt"
	"net/http"
	"path"
	"reflect"
	"KinGo/context"
	"runtime"
//	"errors
	"strconv"
	"strings"
)

const (
	IS_PRINT_LOG = false;
)

const (
	routerTypeYaf = iota
	routerTypeRESTFul
	routerTypeHandler
)

var (
	HTTPMETHOD = map[string]int{
		"GET":     1,
		"POST":    2,
		"PUT":     3,
		"DELETE":  4,
		"PATCH":   5,
		"OPTIONS": 6,
		"HEAD":    7,
		"TRACE":   8,
		"CONNECT": 9,
	}
)

type funcHandler func (  http.ResponseWriter ,  *http.Request );

type handlerContext func ( *context.Context );

type controllerInfo struct {
	pattern        string
	controllerType reflect.Type
	methods        map[string]string
	handler        http.Handler
	runfunction    funcHandler //func ( arg ...interface {}) ( ret_args ... interface {} )
	routerType     int
}

// ControllerRegistor containers registered router rules, controller handlers and filters.
type ControllerRegistor struct {
	routers      map[string]*ControllerTree `Regular expression matching, listen path1, through the path1 to set of matching, matching the callback function to run based on path2|正则匹配，监听path1，通过path1的数来匹配集合，根据path2来匹配运行的回调函数`
}

func newControllerRegistor( ) *ControllerRegistor {
	return &ControllerRegistor{
		routers : make( map[string]*ControllerTree ),
	};
}

/**
*	Add("app/list" , &controler , "*:listFeed")
*	Add("app/list" , &controler , "GET:CreateFeed")
*	......
*
 */
func (p *ControllerRegistor) Add(	pattern string, c ControllerInterface , mappingMethods ...string ) {
	var (
		reflectVal reflect.Value = reflect.ValueOf(c)
		controlerType reflect.Type = reflect.Indirect(reflectVal).Type();
		methods map[string]string  = make(map[string]string)
	)

	if len( mappingMethods ) > 0 {
		var MatchingRules []string = strings.Split( mappingMethods[0], ";" );
		for _, mRules := range MatchingRules {
			var mRulesArr  []string = strings.Split( mRules , ":")
			if len( mRulesArr ) != 2 {
				panic("method mapping format is invalid")
			}
			var httpMethods []string = strings.Split( mRulesArr[0], "," );

			for _, m := range httpMethods {
				if _ , ok := HTTPMETHOD[strings.ToUpper(m)]; m == "*" || ok {
					if val := reflectVal.MethodByName( mRulesArr[1] );  val.IsValid() {
						methods[strings.ToUpper(m)] = mRulesArr[1];
					} else {
						panic( mRulesArr[1] + " method doesn't exist in the controller " + controlerType.Name())
					}
				} else {
					panic( mRules  + " is an invalid method mapping. Method doesn't exist " + m)
				}
			}
		}
	}
	route := &controllerInfo{}
	route.pattern = pattern
	route.methods = methods
	route.routerType = routerTypeYaf
	route.controllerType = controlerType
	var forCall func( string ) = func ( httpMethod string ) {
		p.addToRouter( httpMethod , pattern, route );
	};
	if len( methods ) > 0 {
		for method , _ := range methods {
			if method == "*" {
				auxliForeach(forCall);
//				auxliForeach(func ( httpMethod string ) {
//					p.addToRouter( httpMethod , pattern, route );
//				});
			}else{
				p.addToRouter( method , pattern, route );
			}
		}
	}else{
		auxliForeach(forCall);
//		auxliForeach(func ( httpMethod string ) {
//			p.addToRouter( httpMethod , pattern, route );
//		});
	}

}

func auxliForeach( callBack func( string ) ){
	for httpMethod , _ := range HTTPMETHOD {
		callBack( httpMethod );
	}
}

func (p *ControllerRegistor) addToRouter( method, pattern string, r *controllerInfo) {
	if cTree  , ok := p.routers[method]; ok {
		cTree.AddRouter( pattern , r );
	}else{
		cTree = newControllerTree();
		cTree.AddRouter( pattern , r );
		p.routers[method] = cTree;
	}

}

// add user defined Handler
func (p *ControllerRegistor) Handler( pattern string, h http.Handler, options ...interface{}) {
	route := &controllerInfo{}
	route.pattern = pattern
	route.routerType = routerTypeHandler
	route.handler = h
	if len(options) > 0 {
		if _, ok := options[0].(bool); ok {
			pattern = path.Join(pattern, "?:all")
		}
	}
//	for m , _ := range HTTPMETHOD {
//		p.addToRouter(m, pattern, route)
//	}
	auxliForeach(func ( httpMethod string ) {
		p.addToRouter( httpMethod , pattern, route );
	});
}

func (p *ControllerRegistor) Get( pattern string, f funcHandler) {
	p.AddMethod("get", pattern, f)
}

// add post method
// usage:
//    Post("/api", func(ctx *Context){
//          ctx.Output.Body("hello world")
//    })
func (p *ControllerRegistor) Post(pattern string, f funcHandler) {
	p.AddMethod("post", pattern, f)
}

// add put method
// usage:
//    Put("/api/:id", func(ctx *Context){
//          ctx.Output.Body("hello world")
//    })
func (p *ControllerRegistor) Put(pattern string, f funcHandler) {
	p.AddMethod("put", pattern, f)
}

// add delete method
// usage:
//    Delete("/api/:id", func(ctx *Context){
//          ctx.Output.Body("hello world")
//    })
func (p *ControllerRegistor) Delete(pattern string, f funcHandler) {
	p.AddMethod("delete", pattern, f)
}

// add head method
// usage:
//    Head("/api/:id", func(ctx *Context){
//          ctx.Output.Body("hello world")
//    })
func (p *ControllerRegistor) Head(pattern string, f funcHandler) {
	p.AddMethod("head", pattern, f );
}

// add patch method
// usage:
//    Patch("/api/:id", func(ctx *Context){
//          ctx.Output.Body("hello world")
//    })
func (p *ControllerRegistor) Patch(pattern string, f funcHandler) {
	p.AddMethod("patch", pattern, f)
}

// add options method
// usage:
//    Options("/api/:id", func(ctx *Context){
//          ctx.Output.Body("hello world")
//    })
func (p *ControllerRegistor) Options(pattern string, f funcHandler) {
	p.AddMethod("options", pattern, f)
}

// add all method
// usage:
//    Any("/api/:id", func(ctx *Context){
//          ctx.Output.Body("hello world")
//    })
func (p *ControllerRegistor) Any(pattern string, f funcHandler) {
	p.AddMethod("*", pattern, f)
}

// add http method router
// usage:
//    AddMethod("get","/api/:id", func(ctx *context.Context){
//          ctx.Output.Body("hello world")
//    })
func (p *ControllerRegistor) AddMethod(	 method, pattern string, f funcHandler) {
	if _, ok := HTTPMETHOD[strings.ToUpper(method)]; method != "*" && !ok {
		panic("not support http method: " + method)
	}
	route := &controllerInfo{}
	route.pattern = pattern
	route.routerType = routerTypeRESTFul
	route.runfunction = f
	methods := make(map[string]string)
	if method == "*" {
//		for httpMethod , _ := range HTTPMETHOD {
//			methods[httpMethod] = httpMethod
//		}
		auxliForeach(func ( httpMethod string ) {
			methods[httpMethod] = httpMethod
		});
	} else {
		methods[strings.ToUpper(method)] = strings.ToUpper(method)
	}
	route.methods = methods
	for k, _ := range methods {
		if k == "*" {
//			for m , _ := range HTTPMETHOD {
//				p.addToRouter(m, pattern, route)
//			}
			auxliForeach(func ( httpMethod string ) {
				p.addToRouter( httpMethod , pattern, route);
			});
		} else {
			p.addToRouter(k, pattern, route)
		}
	}
}


func ( kgo_route  *ControllerRegistor) ServeHTTP( w http.ResponseWriter, r *http.Request ) {
	defer kgo_route.recoverPanic( w , r );
	var (
		runRouter reflect.Type
		findRouter bool
		runMethod string
		routerInfo *controllerInfo
		routePath string = r.URL.Path;//path.Clean( r.URL.Path );
	)
	// init context
	ctx := &context.Context{
		ResponseWriter: w,
		Request:        r,
		Input:          context.NewInput(r),
		Output:         context.NewOutput(),
	}
	ctx.Output.Context = ctx

	if _, ok := HTTPMETHOD[r.Method]; !ok {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed);
	}else{
		if kgo.checkIsStatic(routePath,ctx) {
			return;
		}
		if !findRouter {
			if rMethod, ok := kgo_route.routers[r.Method]; ok {
				runObject, p := rMethod.Match( r.URL.Path );

				if cInfo , ok := runObject.(*controllerInfo); ok {
					routerInfo = cInfo
					findRouter = true
				}
				if splat, ok := p[":splat"]; ok {
					splatlist := strings.Split(splat, "/")
					for k, v := range splatlist {
						p[strconv.Itoa(k)] = v
					}
				}
				ctx.Input.Params = p
//				fmt.Println( rMethod )
//				fmt.Println(runObject)
			}
		}
		if findRouter {
			isRunable := false
			if routerInfo != nil {
				if routerInfo.routerType == routerTypeRESTFul {
					if _, ok := routerInfo.methods[r.Method]; ok {
						isRunable = true
						routerInfo.runfunction( w , r );
					} else {
						http.NotFound(w,r);
					}
					return ;
				}else if routerInfo.routerType == routerTypeHandler {
					isRunable = true
					routerInfo.handler.ServeHTTP( w , r)
				} else {
					runRouter = routerInfo.controllerType
					method := r.Method
					if r.Method == "POST" && ctx.Input.Query("_method") == "PUT" {
						method = "PUT"
					}
					if r.Method == "POST" && ctx.Input.Query("_method") == "DELETE" {
						method = "DELETE"
					}
					if m, ok := routerInfo.methods[method]; ok {
						runMethod = m
					} else if m, ok = routerInfo.methods["*"]; ok {
						runMethod = m
					} else {
						runMethod = method
					}
				}
			}
			if !isRunable {
				vc := reflect.New(runRouter)
				execController, ok := vc.Interface().(ControllerInterface)
				if !ok {
					panic("controller is not ControllerInterface")
				}
				execController.Init( ctx , runRouter.Name(), runMethod, vc.Interface())
				execController.Prepare()
				if len( runMethod ) > 0 {
					//exec main logic
					switch runMethod {
					case "GET":
						execController.Get()
					case "POST":
						execController.Post()
					case "DELETE":
						execController.Delete()
					case "PUT":
						execController.Put()
					case "HEAD":
						execController.Head()
					case "PATCH":
						execController.Patch()
					case "OPTIONS":
						execController.Options()
					default:
							in := make([]reflect.Value, 0)
							if method := vc.MethodByName(runMethod); method.IsValid() {
								method.Call(in)
							}else{
								http.NotFound( w , r );
							}
					}
				}
			}
		}else{
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed);
		}
	}
}


func ( kgo_route  *ControllerRegistor) recoverPanic( w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		if !IS_PRINT_LOG {
			panic( err );
		}else{
			for i := 1; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				fmt.Println(file, line)
			}
		}
	}
}

