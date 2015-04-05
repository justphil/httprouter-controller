# httprouter-controller
httprouter-controller is a lightweight and composable controller implementation
for the [httprouter](https://github.com/julienschmidt/httprouter) package. It
is basically a fork of [controller](https://github.com/codegangsta/controller) with
minor changes that satisfy httprouter's [Handle](http://godoc.org/github.com/julienschmidt/httprouter#Handle) function.

Sometimes plain request handlers are not enough, and you want to have logic
that is resource/concept specific, and data that is request specific.

This is where controllers come into play. Controllers are structs that
implement a specific interface related to lifecycle management. A Controller
can also contain an arbitrary amount of methods that can be used as handlers to
incoming requests. This package makes it easy to automatically construct a new
Controller instance and invoke a specified method on that controller for every
request.

## Example

``` go
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

```
