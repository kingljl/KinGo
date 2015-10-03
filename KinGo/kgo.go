package KinGo

import (
	"fmt"
	"os"
	"net/http"
	"runtime"
	"KinGo/context"
	"strings"
	"path"
	"time"
)


const (
	VERSION =  "1.0.0"
)

var (
	//Another frame instances
	kgo *Kgo;
)


func init () {
	runtime.GOMAXPROCS( runtime.NumCPU() );
	frameInit();
}

func frameInit(){
	kgo = &Kgo{
		Handlers : newControllerRegistor(),
		StaticDirs : make( map[string]string ),
		AppPath : getPath(),
		TemplateRegistor : NewTemplateRegistor(),
	};
}

func GetInstance() *Kgo {
	if kgo != nil {
		return kgo;
	}
	frameInit();
	return kgo;
}

/*
*	根据不同的操作系统选择不同目录
*	文件以绝对路径存储
*	os.Getwd()
 */
func getPath() string {
	//发布系统在根目录的时候在调用
	path ,_ := os.Getwd();
	if len( path ) > 0  &&  runtime.GOOS == "linux" {
		var (
			paths []string = strings.Split( path , "/" );
			len int = len( paths );
			retPath string ;
		)
		if len > 0 {
			switch paths[len-1]{
			case "src":
				retPath =  path;
				break;
			case "mistress":
				retPath =  path + "/src";
				break;
			case "Deployment":
				retPath =  path + "/../src";
				break;
			}
			if _ , err := os.Stat(retPath + "/Conf/config.conf"); err == nil {
				return retPath;
			}
		}
		return "/home/q/system/mall/mistress/src";
	}else{
		return "D:/GoWorkspace/src";
	}
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

type Kgo struct {
	Handlers	*ControllerRegistor `The handle to the HTTP server`
	StaticDirs	map[string]string `The different application of different path`
	TemplateRegistor	*TemplateRegistor `The handle to the HTTP server.`
	AppPath string `Application of the default directory`
	
}

func ( kgo *Kgo ) SetStaticPath (  pathList map[string]string ) {
	if len( pathList ) > 0 {
		for dir , val := range pathList {
			kgo.StaticDirs[dir] = val;
		}
	}
}

func ( kgo *Kgo ) BuildTemplate( TemplatePath string ) {
	if TemplatePath != "" {
		kgo.TemplateRegistor.buildTemplate( TemplatePath );
	}
}

func (  kgo *Kgo  ) AddRouter ( pattern string , c ControllerInterface , mappingMethods ...string ) {
	kgo.Handlers.Add( pattern , c , mappingMethods... );
}

func (  kgo *Kgo  ) Get (  pattern string , CallBack funcHandler  ) {
	kgo.Handlers.Get( pattern , CallBack );
}

func (  kgo *Kgo ) Post ( pattern string , CallBack funcHandler  ) {
	kgo.Handlers.Post( pattern , CallBack );
}

func (  kgo *Kgo  ) Delete( pattern string , CallBack funcHandler )  {
	kgo.Handlers.Delete( pattern , CallBack );
}

func (  kgo *Kgo  ) Put( pattern string , CallBack funcHandler ) {
	kgo.Handlers.Put( pattern , CallBack );
}

func (  kgo *Kgo  ) Head( pattern string , CallBack funcHandler ) {
	kgo.Handlers.Head( pattern , CallBack );
}

func (  kgo *Kgo  ) Options(  pattern string , CallBack funcHandler  ) {
	kgo.Handlers.Options( pattern , CallBack );
}


func (  kgo *Kgo  ) Any(  pattern string , CallBack funcHandler ) {
	kgo.Handlers.Any( pattern , CallBack );
}

func (  kgo *Kgo  ) Handler( pattern string , h http.Handler, options ...interface{} ) {
	kgo.Handlers.Handler( pattern , h , options... );
}

func (  kgo *Kgo  ) checkIsStatic( routePath string , ctx *context.Context ) ( res bool ) {
	for prefix, staticDir := range kgo.StaticDirs {
		if len(prefix) > 0 {
			if routePath == "/favicon.ico" {
				res = true;
				file := path.Join(staticDir, routePath)
				if FileExists(file) {
					http.ServeFile(ctx.ResponseWriter, ctx.Request, file)
					return
				}
			}else if strings.HasPrefix(routePath, prefix) && routePath[len(prefix)-1] == '/' {
				file := path.Join(staticDir, routePath[len(prefix):]);
				res = true;
				finfo, err := os.Stat(file)
				if err != nil {
					ctx.NotFound("404 page not found ");
					return
				}else if finfo.IsDir() {
					ctx.Error(403,"403 Forbidden")
					return;
				}
				http.ServeFile(ctx.ResponseWriter, ctx.Request, file)
				return;
			}
		}
	}
	return;
}


func (  kgo *Kgo  ) Run( isHold bool , UDF func() ) {
	if UDF != nil {
		go UDF();
	}else{
		var (
			sAddr string = ":8080";
			sRt time.Duration = time.Millisecond * 500;
			sWt time.Duration = time.Millisecond * 500;
		)
		server := http.Server{
			Addr:       sAddr,
			Handler : kgo.Handlers,
			ReadTimeout : sRt,
			WriteTimeout : sWt,
		}
		go func(){
			error := server.ListenAndServe();
			if error != nil {
				fmt.Println(error);
				os.Exit(0);
			}
		}();

		if isHold {
			select {}
		}
	}
}

