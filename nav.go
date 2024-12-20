package zop

import (
	"net/http"
	"strings"
)

type nav struct {
	roots    map[string]*node
	handlers map[string]NavHandlerFunc
}

func newRouter() *nav {
	return &nav{
		roots:    make(map[string]*node),
		handlers: make(map[string]NavHandlerFunc),
	}
}

func (nav *nav) parsePattern(pattern string) []string {
	ss := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, s := range ss {
		if s != "" {
			parts = append(parts, s)
			if strings.HasPrefix(s, "*") {
				break
			}
		}
	}
	return parts
}

func (nav *nav) AddRoute(method string, path string, handler NavHandlerFunc) {
	key := method + "-" + path
	parts := nav.parsePattern(path)
	root, ok := nav.roots[method]
	if !ok {
		root = &node{}
		nav.roots[method] = root
	}
	root.Insert(path, parts, 0)
	nav.handlers[key] = handler
}

func (nav *nav) getRoute(method string, path string) (*node, map[string]string) {
	root, ok := nav.roots[method]
	if !ok {
		return nil, nil
	}
	searchParts := nav.parsePattern(path)
	node := root.Search(searchParts, 0)
	if node == nil {
		return nil, nil
	}
	parts := nav.parsePattern(node.pattern)
	params := make(map[string]string)
	for index, part := range parts {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = searchParts[index]
		}
		if strings.HasPrefix(part, "*") {
			params[part[1:]] = strings.Join(searchParts[index:], "/")
		}
	}
	return node, params
}

func (nav *nav) Handle(c *Context) {
	node, params := nav.getRoute(c.Method, c.Path)
	if node != nil {
		c.Params = params
		key := c.Method + "-" + node.pattern
		c.handlers = append(c.handlers, nav.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusInternalServerError, "500 BAD GATEWAY : %s\n", c.Path)
		})
	}
	c.Next()
}
