package main

import (
	"KinGo"
	"net/http"
	"fmt"

)

var (
	Kgo *KinGo.Kgo = KinGo.GetInstance();
)


type MainController struct {
	KinGo.Controller
}

func (m *MainController) Get(){
	fmt.Println(m.Ctx.Input.Params)
	m.Ctx.ResponseWriter.Write([]byte("路由Get执行成功"));
	fmt.Println("路由Get执行成功")
}

func (m *MainController) Post(){
	fmt.Println(m.Ctx.Input.Params)
	m.Ctx.ResponseWriter.Write([]byte("路由Post执行成功"));
	fmt.Println("路由Post执行成功")
}

func (m *MainController) CreateFeed(){
	fmt.Println(m.Ctx.Input.Param(":id"))
	m.Ctx.ResponseWriter.Write([]byte("CreateFeed 执行成功"));
	fmt.Println("CreateFeed 执行成功")
}


func (m *MainController) ReanderHtm(){
	m.TplNames = "text/text.tpl";
	m.Data["ddd"] = "ddddddddd";
	fmt.Println( m.Data )
	m.RenderTemplate();
}

func main(){
	Kgo.SetStaticPath(map[string]string{"/static/":"src/templates/static"});
	Kgo.BuildTemplate( "src/templates/views" );
	Kgo.Get("/" , func (  w http.ResponseWriter,  r *http.Request) {
		fmt.Println("路由\\/执行成功")
		w.Write([]byte("路由\\/执行成功"));
	});
	Kgo.Get("/api.html" , func (  w http.ResponseWriter,  r *http.Request) {
		fmt.Println("路由api执行成功")
		w.Write([]byte("路由api执行成功"));
	});
	Kgo.AddRouter( "/api/app/:id" , &MainController{} , "get:CreateFeed");
	Kgo.AddRouter( "/collection/view/filter/:id" , &MainController{});
	Kgo.AddRouter( "/collection/:id/:name/:string" , &MainController{});
	Kgo.AddRouter( "/renderhtml" , &MainController{},"get:ReanderHtm");
	Kgo.Run( true , nil );
}
