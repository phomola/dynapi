// Dynapi is a simple API multiplexer.
package dynapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

var (
	noneType  = reflect.TypeOf((*None)(nil)).Elem()
	noneValue = reflect.ValueOf(new(None))
)

type serviceParam struct {
	name   string
	offset uintptr
	kind   reflect.Kind
}

// None is an empty type.
// It can be used in handlers to indicate that they have no parameters or input object.
type None struct{}

// Context is a call context passed to the handler.
type Context struct {
	// Data can contain custom data provided when the handler was reigstered.
	Data interface{}
	// Request is the underlying HTTP request.
	Request *http.Request
}

// Mux is an API multiplexer.
type Mux struct {
	mux    *http.ServeMux
	routes map[string]map[string]func(http.ResponseWriter, *http.Request)
}

// New returns a new multiplexer.
func New() *Mux {
	return &Mux{mux: http.NewServeMux(), routes: make(map[string]map[string]func(http.ResponseWriter, *http.Request))}
}

// Handler returns an HTTP handler for the multiplexer.
func (m *Mux) Handler() *http.ServeMux { return m.mux }

// HandleAll registers all the provided handler functions.
func (m *Mux) HandleAll(routePrefix string, data interface{}, fs ...interface{}) error {
	for _, f := range fs {
		if err := m.Handle(routePrefix, data, f); err != nil {
			return err
		}
	}
	return nil
}

// Handle registers a handler function.
// The name of the function must be camelcased.
// The first segment of the function's name is the HTTP method (get, post, put, delete or patch).
// The remainder of the function's name specifies the route.
// For example, the route for getBooksAll is /books/all.
func (m *Mux) Handle(routePrefix string, data interface{}, f interface{}) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return fmt.Errorf("mux handler must be a function, got '%v' (%T)", f, f)
	}
	t := v.Type()
	n := runtime.FuncForPC(v.Pointer()).Name()
	comps := strings.Split(n, ".")
	n = comps[len(comps)-1]
	comps = splitCamelcasedString(n)
	for i, comp := range comps {
		comps[i] = strings.ToLower(comp)
	}
	var method string
	switch comps[0] {
	case "get":
		method = http.MethodGet
	case "post":
		method = http.MethodPost
	case "put":
		method = http.MethodPut
	case "patch":
		method = http.MethodPatch
	case "delete":
		method = http.MethodDelete
	default:
		return fmt.Errorf("unknown HTTP method '%s' in '%s'", comps[0], n)
	}
	route := routePrefix + "/" + strings.Join(comps[1:], "/") + "/"
	mm, ok := m.routes[route]
	if !ok {
		mm = make(map[string]func(http.ResponseWriter, *http.Request))
		m.routes[route] = mm
		m.mux.HandleFunc(route, func(w http.ResponseWriter, req *http.Request) {
			if f, ok := mm[req.Method]; ok {
				f(w, req)
			} else {
				http.Error(w, fmt.Sprintf("couldn't handle method for route (%s %s)", req.Method, route), http.StatusNotFound)
			}
		})
	}
	if _, ok := mm[method]; ok {
		return fmt.Errorf("handler for '%s' (%s) already registered (can't register '%s')", route, method, n)
	}
	paramsType, argType := t.In(1).Elem(), t.In(2).Elem()
	serviceParams := make([]*serviceParam, paramsType.NumField())
	for i := 0; i < paramsType.NumField(); i++ {
		f := paramsType.Field(i)
		serviceParams[i] = &serviceParam{strings.ToLower(f.Name), f.Offset, f.Type.Kind()}
	}
	hasParams, hasInput := paramsType != noneType, argType != noneType
	log.Printf("registering handler: %s %s", method, route)
	mm[method] = func(w http.ResponseWriter, req *http.Request) {
		params := noneValue
		if hasParams {
			params = reflect.New(paramsType)
			suffix := req.URL.Path[len(route):]
			if suffix != "" {
				paramsSlice := strings.Split(suffix, "/")
				for i, p := range serviceParams {
					if i >= len(paramsSlice) {
						break
					}
					v := paramsSlice[i]
					switch p.kind {
					case reflect.String:
						*(*string)(unsafe.Pointer(params.Pointer() + p.offset)) = v
					case reflect.Int:
						n, err := strconv.Atoi(v)
						if err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
						*(*int)(unsafe.Pointer(params.Pointer() + p.offset)) = n
					}
				}
			}
		}
		in := noneValue
		if hasInput {
			in = reflect.New(argType)
			dec := json.NewDecoder(req.Body)
			err := dec.Decode(in.Interface())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		ctx := &Context{Data: data, Request: req}
		r := v.Call([]reflect.Value{reflect.ValueOf(ctx), params, in})
		out, err2 := r[0].Interface(), r[1].Interface()
		if err2 != nil {
			http.Error(w, err2.(error).Error(), http.StatusInternalServerError)
			return
		}
		enc := json.NewEncoder(w)
		err := enc.Encode(out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return nil
}
