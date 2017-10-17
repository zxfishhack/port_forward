package port_forward

import (
	"github.com/kataras/iris"
)


func RegisterRouter(app *iris.Application) {
	actionMiddleware := func(ctx iris.Context) {
		ctx.Next()
	}

	mgr := NewManager()
	
	actions := app.Party("/console", actionMiddleware)
	{
		actions.Get("/", mgr.ListAll)
		actions.Get("/add", mgr.Add)
		actions.Get("/add-http", mgr.AddHttp)
		actions.Get("/del/{RuleID:int}", mgr.Del)
		//actions.Get("/{RuleID:int}/filters", mgr.ViewFilter)
		//actions.Get("/{RuleID:int}/filters/add", mgr.AddFilter)
		//actions.Get("/{RuleID:int}/filters/del/{FilterID:int}", mgr.DelFilter)
	}
}