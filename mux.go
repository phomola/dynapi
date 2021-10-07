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

// HandlerContext is a context associated with a request.
type HandlerContext struct {
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

// FinishSetup frees up some resources that are no longer needed after setting the service up.
func (m *Mux) FinishSetup() {
	m.routes = nil
}

// Handler returns an HTTP handler for the multiplexer.
func (m *Mux) Handler() *http.ServeMux { return m.mux }

// HandleAll registers all the provided handler functions.
func (m *Mux) HandleAll(routePrefix string, ctx interface{}, fs ...interface{}) error {
	for _, f := range fs {
		if err := m.Handle(routePrefix, ctx, f); err != nil {
			return err
		}
	}
	return nil
}

// HandleService registers an object's functions as handlers.
func (m *Mux) HandleService(routePrefix string, s interface{}) error {
	t := reflect.TypeOf(s)
	for i := 0; i < t.NumMethod(); i++ {
		mt := t.Method(i)
		if err := m.Handle(routePrefix, s, mt.Func.Interface()); err != nil {
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
func (m *Mux) Handle(routePrefix string, ctx interface{}, f interface{}) error {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return fmt.Errorf("mux handler must be a function, got '%v' (%T)", f, f)
	}
	t := v.Type()
	if t.NumIn() != 4 {
		return fmt.Errorf("handler functions must take four parameters")
	}
	if t.NumOut() != 2 {
		return fmt.Errorf("handler functions must return two values")
	}
	if t.In(0).Kind() != reflect.Ptr {
		return fmt.Errorf("the first argument of handler functions must be a pointer")
	}
	if t.In(1) != reflect.TypeOf((*HandlerContext)(nil)) {
		return fmt.Errorf("the second argument of handler functions must be a pointer to a handler context")
	}
	if t.In(2).Kind() != reflect.Ptr {
		return fmt.Errorf("the third argument of handler functions must be a pointer")
	}
	if t.In(3).Kind() != reflect.Ptr {
		return fmt.Errorf("the fourth argument of handler functions must be a pointer")
	}
	if t.Out(0).Kind() != reflect.Ptr {
		return fmt.Errorf("the first value returned by handler functions must be a pointer")
	}
	if t.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("the second value returned by handler functions must be an error")
	}
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
	paramsType, argType := t.In(2).Elem(), t.In(3).Elem()
	pattern := route
	serviceParams := make([]*serviceParam, paramsType.NumField())
	for i := 0; i < paramsType.NumField(); i++ {
		f := paramsType.Field(i)
		if f.Type.Kind() != reflect.String && f.Type.Kind() != reflect.Int && f.Type.Kind() != reflect.Float32 && f.Type.Kind() != reflect.Float64 {
			return fmt.Errorf("a value of type '%s' can't be a handler parameter", f.Type.Kind())
		}
		n := strings.ToLower(f.Name)
		serviceParams[i] = &serviceParam{n, f.Offset, f.Type.Kind()}
		pattern += "{" + n + "}/"
	}
	hasParams, hasInput := paramsType != noneType, argType != noneType
	ctxVal := reflect.ValueOf(ctx)
	log.Printf("registering handler: %s %s (%s)", method, route, pattern)
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
					case reflect.Float32:
						x, err := strconv.ParseFloat(v, 32)
						if err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
						*(*float32)(unsafe.Pointer(params.Pointer() + p.offset)) = float32(x)
					case reflect.Float64:
						x, err := strconv.ParseFloat(v, 64)
						if err != nil {
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
						*(*float64)(unsafe.Pointer(params.Pointer() + p.offset)) = x
					default:
						http.Error(w, fmt.Sprintf("a value of type '%s' can't be a handler parameter", p.kind), http.StatusInternalServerError)
						return
					}
				}
			}
		}
		in := noneValue
		if hasInput {
			in = reflect.New(argType)
			dec := json.NewDecoder(req.Body)
			if err := dec.Decode(in.Interface()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		handlerCtx := &HandlerContext{req}
		r := v.Call([]reflect.Value{ctxVal, reflect.ValueOf(handlerCtx), params, in})
		out, err2 := r[0].Interface(), r[1].Interface()
		if err2 != nil {
			http.Error(w, err2.(error).Error(), http.StatusInternalServerError)
			return
		}
		enc := json.NewEncoder(w)
		if err := enc.Encode(out); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	return nil
}
