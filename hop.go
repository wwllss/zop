package zop

import (
	"net/http"
	"strings"
)

type NavHandlerFunc func(c *Context)

type NavGroup struct {
	zop         *Zop
	prefix      string
	parent      *NavGroup
	middlewares []NavHandlerFunc
}

type Zop struct {
	*NavGroup
	groups []*NavGroup
	nav    *nav
}

func New() *Zop {
	zop := &Zop{nav: newRouter()}
	zop.NavGroup = &NavGroup{zop: zop}
	zop.groups = []*NavGroup{zop.NavGroup}
	return zop
}

func (group *NavGroup) Group(prefix string) *NavGroup {
	zop := group.zop
	for _, g := range zop.groups {
		if g.prefix == prefix {
			return g
		}
	}
	newGroup := &NavGroup{
		zop:    zop,
		prefix: prefix,
		parent: group,
	}
	zop.groups = append(zop.groups, newGroup)
	return newGroup
}

func (group *NavGroup) Use(middlewares ...NavHandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *NavGroup) AddRoute(method string, path string, handler NavHandlerFunc) {
	group.zop.nav.AddRoute(method, group.prefix+path, handler)
}

func (group *NavGroup) GET(path string, handler NavHandlerFunc) {
	group.AddRoute("GET", path, handler)
}

func (group *NavGroup) POST(path string, handler NavHandlerFunc) {
	group.AddRoute("POST", path, handler)
}

func (zop *Zop) Run(addr string) (err error) {
	return http.ListenAndServe(addr, zop)
}

func (zop *Zop) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	middlewares := make([]NavHandlerFunc, 0)
	for _, group := range zop.groups {
		if strings.HasPrefix(request.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := NewContext(writer, request)
	c.handlers = middlewares
	zop.nav.Handle(c)
}
