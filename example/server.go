package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justphil/httprouter-controller"
)

type MyController struct {
	controller.Base
}

func (c *MyController) Index() error {
	name := c.Params.ByName("name")
	_, err := fmt.Fprintf(c.ResponseWriter, "URL param 'name' has been set to: %s", name)
	return err
}

func main() {
	router := httprouter.New()
	router.GET("/hello/:name", controller.Action((*MyController).Index))
	http.ListenAndServe(":3000", router)
}
