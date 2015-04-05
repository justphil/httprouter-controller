// Package controller is a lightweight and composable controller implementation
// for github.com/julienschmidt/httprouter
//
// Sometimes plain httprouter handlers are not enough, and you want to have logic
// that is resource/concept specific, and data that is request specific.
//
// This is where controllers come into play. Controllers are structs that
// implement a specific interface related to lifecycle management. A Controller
// can also contain an arbitrary amount of methods that can be used as handlers
// to incoming requests. This package makes it easy to automatically construct
// a new Controller instance and invoke a specified method on that controller
// for every request.
//
// This is an example controller:
//
//    type MyController struct {
//      controller.Base
//    }
//
//    func (c *MyController) Index() error {
//		name := c.Params.ByName("name")
//      c.ResponseWriter.Write([]byte("URL Param name set to: " + name))
//      return nil
//    }
//
// To handle HTTP requests with this controller, use the controller.Action
// function:
//    router := httprouter.New()
//    router.GET("/hello/:name", controller.Action((*MyController).Index))
//
package controller

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

// Controller is an interface for defining a web controller that can be
// automatically constructed via the controller.Action function. This interface
// contains lifecycle methods that are vital during the controllers lifetime.
// A controller instance is constructed every time the httprouter.Handle result from
// controller.Action is invoked (this is usually every http request)
type Controller interface {
	// Init initializes the controller. If it returns an error, then the Error
	// method on the controller will be invoked.
	Init(http.ResponseWriter, *http.Request, httprouter.Params) error
	// Destroy is called after the Controllers action has been called or after an
	// error has occured. This is a useful method for cleaning up anything that
	// was initialized.
	Destroy()
	// Error is the error handling mechanism for the controller. It is called if
	// Init or controller action return an error. It can also be invoked manually
	// for consistent error handling across a controller.
	Error(code int, error string)
}

// Base is a base implementation for a Controller. It contains the Request and
// ResponseWriter objects for controller actions to easily consume. Additionally
// it contains the httprouter.Params allowing you to access the named parameters
// provided via the httprouter.Handle function.
// Base is meant to be embedded in your own controller struct.
type Base struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Params         httprouter.Params
}

// Init initializes the base controller with a ResponseWriter, Request and httprouter.Params.
// Embedders of this struct should remember to call Init if the embedder is
// implementing the Init function themselves.
func (b *Base) Init(rw http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	b.Request, b.ResponseWriter, b.Params = r, rw, params
	return nil
}

// Destroy performs cleanup for the base controller
func (b *Base) Destroy() {
}

// Error will send an HTTP error to the given ResponseWriter from Init
func (b *Base) Error(code int, error string) {
	http.Error(b.ResponseWriter, error, code)
}

// Action takes a method expression and translates it into a callable
// httprouter.Handle which, when called:
//
// 		1. Constructs a controller instance
// 		2. Initializes the controller via the Init function
// 		3. Invokes the Action method referenced by the method expression
// 		4. Calls destroy on the controller
//
// This flow allows for similar logic to be cleanly reused while data is no
// longer shared between requests. This is because a new Controller instance
// will be constructed every time the returned httprouter.Handle's ServeHTTP method
// is invoked.
//
// An example of a valid method expression is:
//
// 		controller.Action((*MyController).Index)
//
// Where MyController is an implementor of the Controller interface and Index
// is a method on MyController that takes no arguments and returns an err
func Action(action interface{}) httprouter.Handle {
	val := reflect.ValueOf(action)
	t, err := controllerType(val)
	if err != nil {
		panic(err)
	}

	return func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		v := reflect.New(t)
		c := v.Interface().(Controller)
		err = c.Init(rw, r, params)
		defer c.Destroy()
		if err != nil {
			c.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ret := val.Call([]reflect.Value{v})[0].Interface()
		if ret != nil {
			c.Error(http.StatusInternalServerError, ret.(error).Error())
			return
		}
	}
}

func controllerType(action reflect.Value) (reflect.Type, error) {
	t := action.Type()

	if t.Kind() != reflect.Func {
		return t, errors.New("Action is not a function")
	}

	if t.NumIn() != 1 {
		return t, errors.New("Wrong Number of Arguments in action")
	}

	if t.NumOut() != 1 {
		return t, errors.New("Wrong Number of return values in action")
	}

	out := t.Out(0)
	if !out.Implements(interfaceOf((*error)(nil))) {
		return t, errors.New("Action return type invalid")
	}

	t = t.In(0)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if !reflect.PtrTo(t).Implements(interfaceOf((*Controller)(nil))) {
		return t, errors.New("Controller does not implement ctrl.Controller interface")
	}

	return t, nil
}

func interfaceOf(value interface{}) reflect.Type {
	t := reflect.TypeOf(value)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}
