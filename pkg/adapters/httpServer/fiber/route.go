package fiber

import "github.com/armineyvazi/common.git/pkg/ports"

func (fh *FiberHttpServer) SetRouteGroups(groupName string, middlewares []func(ctx *ports.HttpContext) error, routes []ports.Route) {
	g := fh.app.Group("/" + groupName)

	for _, middleware := range middlewares {
		g.Use(middleware)
	}

	for _, route := range routes {
		g.Add(string(route.Method), route.Path, route.Handler)
	}
}
