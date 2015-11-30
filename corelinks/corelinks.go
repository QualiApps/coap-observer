package core

import (
	"strings"
)

type (
	attribute struct {
		key   string
		value interface{}
	}

	attributes []*attribute

	resource struct {
		link   string
		params attributes
	}
)

func NewResource() *resource {
	return new(resource)
}

func newAttribute(key string, value interface{}) *attribute {
	return &attribute{key, value}
}

func (r *resource) AddLink(link string) {
	r.link = link
}

func (r *resource) AddAttribute(key string, value interface{}) {
	r.params = append(r.params, newAttribute(key, value))
}

func (r *resource) GetLink() string {
	return r.link
}

func (r *resource) GetParams() attributes {
	return r.params
}

func (r *resource) GetParamByKey(key string) *attribute {
	params := r.GetParams()
	for _, p := range params {
		if p.key == key {
			return p
		}
	}
	return nil
}

func parseAttributes(res *resource, attribs []string) {
	for _, attr := range attribs {
		pair := strings.Split(attr, "=")
		if len(pair) == 2 {
			key := pair[0]
			value := strings.Replace(pair[1], "\"", "", -1)
			res.AddAttribute(key, value)
		}
	}
}

func Parse(links string) ([]*resource, bool) {
	var (
		resources []*resource
		ok        = false
	)

	items := strings.Split(links, ",")
	for _, part := range items {
		parts := strings.Split(part, ";")
		// Add resource url
		resourceURL := parts[0][1 : len(parts[0])-1]
		res := NewResource()
		res.AddLink(resourceURL)
		if len(parts) > 1 {
			attribs := parts[1:]
			parseAttributes(res, attribs)
		}

		resources = append(resources, res)
	}
	if len(resources) != 0 {
		ok = true
	}

	return resources, ok
}
