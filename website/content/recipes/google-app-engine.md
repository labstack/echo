---
title: Google App Engine
draft: true
menu:
  main:
    parent: recipes
---

Google App Engine (GAE) provides a range of hosting options from pure PaaS (App Engine Classic)
through Managed VMs to fully self-managed or container-driven Compute Engine instances. Echo
works great with all of these but requires a few changes to the usual examples to run on the
AppEngine Classic and Managed VM options. With a small amount of effort though it's possible
to produce a codebase that will run on these and also non-managed platforms automatically.

We'll walk through the changes needed to support each option.

### Standalone

Wait? What? I thought this was about AppEngine! Bear with me - the easiest way to show the changes
required is to start with a setup for standalone and work from there plus there's no reason we
wouldn't want to retain the ability to run our app anywhere, right?

We take advantage of the go [build constraints or tags](http://golang.org/pkg/go/build/) to change
how we create and run the Echo server for each platform while keeping the rest of the application
(e.g. handler wireup) the same across all of them.

First, we have the normal setup based on the examples but we split it into two files - `app.go` will
be common to all variations and holds the Echo instance variable. We initialise it from a function
and because it is a `var` this will happen _before_ any `init()` functions run - a feature that we'll
use to connect our handlers later.

`app.go`
```go
package main

// referecnce our echo instance and create it early
var e = createMux()
```  

A separate source file contains the function to create the Echo instance and add the static
file handlers and middleware. Note the build tag on the first line which says to use this when _not_
bulding with appengine or appenginevm tags (which thoese platforms automatically add for us). We also
have the `main()` function to start serving our app as normal. This should all be very familiar.

`app-standalone.go`
```go
// +build !appengine,!appenginevm

package main

import (
    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
)

func createMux() *echo.Echo {
    e := echo.New()

    e.Use(middleware.Recover())
    e.Use(middleware.Logger())
    e.Use(middleware.Gzip())

    e.Index("public/index.html")
    e.Static("/public", "public")

    return e
}

func main() {
    e.Run(":8080")
}
```

The handler-wireup that would normally also be a part of this Echo setup moves to separate files which
take advantage of the ability to have multiple `init()` functions which run _after_ the `e` Echo var is
initialised but _before_ the `main()` function is executed. These allow additional handlers to attach
themselves to the instance - I've found the `Group` feature naturally fits into this pattern with a file
per REST endpoint, often with a higher-level `api` group created that they attach to instead of the root
Echo instance directly (so things like CORS middleware can be added at this higher common-level).

`some-endpoint.go`
```go
package main

import "github.com/labstack/echo"

func init() {
	// our endpoint registers itself with the echo instance
	// and could add it's own middleware if it needed
	g := e.Group("/some-endpoint")
	g.Get("", listHandler)
	g.Get("/:id", getHandler)
	g.Patch("/:id", updateHandler)
	g.Post("/:id", addHandler)
	g.Put("/:id", replaceHandler)
	g.Delete("/:id", deleteHandler)
}

func listHandler(c *echo.Context) error {
	// actual handler implementations ...
```

If we run our app it should execute as it did before when everything was in one file although we have
at least gained the ability to organize our handlers a little more cleanly.

### AppEngine Classic and Managed VMs

So far we've seen how to split apart the Echo creation and setup but still have the same app that
still only runs standalone. Now we'll see hwo those changes allow us to add support for AppEngine
hosting.

Refer to the [AppEngine site](https://cloud.google.com/appengine/docs/go/) for full configuration
and deployment information.

#### app.yaml configuration file

Both of these are Platform as as Service options running on either sandboxed micro-containers
or managed Compute Engine instances. Both require an `app.yaml` file to describe the app to
the service. While the app _could_ still serve all it's static files itself, one of the benefits
of the platform is having Google's infrastructure handle that for us so it can be offloaded and
the app only has to deal with dynamic requests. The platform also handles logging and http gzip
compression so these can be removed from the codebase as well.

The yaml file also contains other options to control instance size and auto-scaling so for true
deployment freedom you would likely have separate `app-classic.yaml` and `app-vm.yaml` files and
this can help when making the transition from AppEngine Classic to Managed VMs.

`app.yaml`
```yaml
application: my-application-id  # defined when you create your app using google dev console
module: default                 # see https://cloud.google.com/appengine/docs/go/
version: alpha                  # you can run multiple versions of an app and A/B test
runtime: go                     # see https://cloud.google.com/appengine/docs/go/
api_version: go1                # used when appengine supports different go versions

default_expiration: "1d"        # for CDN serving of static files (use url versioning if long!)

handlers:
# all the static files that we normally serve ourselves are defined here and Google will handle
# serving them for us from it's own CDN / edge locations. For all the configuration options see:
# https://cloud.google.com/appengine/docs/go/config/appconfig#Go_app_yaml_Static_file_handlers
- url: /
  mime_type: text/html
  static_files: public/index.html
  upload: public/index.html

- url: /favicon.ico
  mime_type: image/x-icon
  static_files: public/favicon.ico
  upload: public/favicon.ico

- url: /scripts
  mime_type: text/javascript
  static_dir: public/scripts

# static files normally don't touch the server that the app runs on but server-side template files
# needs to be readable by the app. The application_readable option makes sure they are available as
# part of the app deployment onto the instance.
- url: /templates
  static_dir: /templates
  application_readable: true

# finally, we route all other requests to our application. The script name just means "the go app"
- url: /.*
  script: _go_app
```

#### Router configuration

We'll now use the [build constraints](http://golang.org/pkg/go/build/) again like we did when creating
our `app-standalone.go` instance but this time with the opposite tags to use this file _if_ the build has
the appengine or appenginevm tags (added automatically when deploying to these platforms).

This allows us to replace the `createMux()` function to create our Echo server _without_ any of the
static file handling and logging + gzip middleware which is no longer required. Also worth nothing is
that GAE classic provides a wrapper to handle serving the app so instead of a `main()` function where
we run the server, we instead wire up the router to the default `http.Handler` instead.

`app-engine.go`
```go
// +build appengine

package main

import (
    "net/http"
    "github.com/labstack/echo"
)

func createMux() *echo.Echo {
    e := echo.New()

    // note: we don't need to provide the middleware or static handlers, that's taken care of by the platform
    // app engine has it's own "main" wrapper - we just need to hook echo into the default handler
    http.Handle("/", e)

    return e
}
```

Managed VMs are slightly different. They are expected to respond to requests on port 8080 as well
as special health-check requests used by the service to detect if an instance is still running in
order to provide automated failover and instance replacement. The `google.golang.org/appengine`
package provides this for us so we have a slightly different version for Managed VMs:

`app-managed.go`
```go
// +build appenginevm

package main

import (
    "runtime"
    "net/http"
    "github.com/labstack/echo"
    "google.golang.org/appengine"
)

func createMux() *echo.Echo {
    // we're in a container on a Google Compute Engine instance so are not sandboxed anymore ...
    runtime.GOMAXPROCS(runtime.NumCPU())

    e := echo.New()

    // note: we don't need to provide the middleware or static handlers
    // for the appengine vm version - that's taken care of by the platform

    return e
}

func main() {
    // the appengine package provides a convenient method to handle the health-check requests
    // and also run the app on the correct port. We just need to add Echo to the default handler
    http.Handle("/", e)
    appengine.Main()
}
```

So now we have three different configurations. We can build and run our app as normal so it can
be executed locally, on a full Compute Engine instance or any other traditional hosting provider
(including EC2, Docker etc...). This build will ignore the code in appengine and appenginevm tagged
files and the `app.yaml` file is meaningless to anything other than the AppEngine platform.

We can also run locally using the [Google AppEngine SDK for GO](https://cloud.google.com/appengine/downloads)
either emulating [AppEngine Classic](https://cloud.google.com/appengine/docs/go/tools/devserver):

    goapp serve

Or [Managed VMs](https://cloud.google.com/appengine/docs/managed-vms/sdk#run-local):

    gcloud config set project [your project id]
    gcloud preview app run .

And of course we can deploy our app to both of these platforms for easy and inexpensive auto-scaling joy.

Depending on what your app actually does it's possible you may need to make other changes to allow
switching between AppEngine provided service such as Datastore and alternative storage implementations
such as MongoDB. A combination of go interfaces and build constraints can make this fairly straightforward
but is outside the scope of this recipe.  


## [Source Code](https://github.com/labstack/echo/blob/master/recipes/google-app-engine)
