+++
title = "Google App Engine"
description = "Google App Engine example for Echo"
[menu.main]
  name = "Google App Engine"
  parent = "cookbook"
  weight = 12
+++

Google App Engine (GAE) provides a range of hosting options from pure PaaS (App Engine Classic)
through Managed VMs to fully self-managed or container-driven Compute Engine instances. Echo
works great with all of these but requires a few changes to the usual examples to run on the
AppEngine Classic and Managed VM options. With a small amount of effort though it's possible
to produce a codebase that will run on these and also non-managed platforms automatically.

We'll walk through the changes needed to support each option.

## Standalone

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

{{< embed "google-app-engine/app.go" >}}

A separate source file contains the function to create the Echo instance and add the static
file handlers and middleware. Note the build tag on the first line which says to use this when _not_
bulding with appengine or appenginevm tags (which thoese platforms automatically add for us). We also
have the `main()` function to start serving our app as normal. This should all be very familiar.

`app-standalone.go`

{{< embed "google-app-engine/app-standalone.go" >}}

The handler-wireup that would normally also be a part of this Echo setup moves to separate files which
take advantage of the ability to have multiple `init()` functions which run _after_ the `e` Echo var is
initialised but _before_ the `main()` function is executed. These allow additional handlers to attach
themselves to the instance - I've found the `Group` feature naturally fits into this pattern with a file
per REST endpoint, often with a higher-level `api` group created that they attach to instead of the root
Echo instance directly (so things like CORS middleware can be added at this higher common-level).

`users.go`

{{< embed "google-app-engine/users.go" >}}

If we run our app it should execute as it did before when everything was in one file although we have
at least gained the ability to organize our handlers a little more cleanly.

## AppEngine Classic and Managed VMs

So far we've seen how to split apart the Echo creation and setup but still have the same app that
still only runs standalone. Now we'll see hwo those changes allow us to add support for AppEngine
hosting.

Refer to the [AppEngine site](https://cloud.google.com/appengine/docs/go/) for full configuration
and deployment information.

### app.yaml configuration file

Both of these are Platform as as Service options running on either sandboxed micro-containers
or managed Compute Engine instances. Both require an `app.yaml` file to describe the app to
the service. While the app _could_ still serve all it's static files itself, one of the benefits
of the platform is having Google's infrastructure handle that for us so it can be offloaded and
the app only has to deal with dynamic requests. The platform also handles logging and http gzip
compression so these can be removed from the codebase as well.

The yaml file also contains other options to control instance size and auto-scaling so for true
deployment freedom you would likely have separate `app-classic.yaml` and `app-vm.yaml` files and
this can help when making the transition from AppEngine Classic to Managed VMs.

`app-engine.yaml`

{{< embed "google-app-engine/app-engine.yaml" >}}

### Router configuration

We'll now use the [build constraints](http://golang.org/pkg/go/build/) again like we did when creating
our `app-standalone.go` instance but this time with the opposite tags to use this file _if_ the build has
the appengine or appenginevm tags (added automatically when deploying to these platforms).

This allows us to replace the `createMux()` function to create our Echo server _without_ any of the
static file handling and logging + gzip middleware which is no longer required. Also worth nothing is
that GAE classic provides a wrapper to handle serving the app so instead of a `main()` function where
we run the server, we instead wire up the router to the default `http.Handler` instead.

`app-engine.go`

{{< embed "google-app-engine/app-engine.go" >}}

Managed VMs are slightly different. They are expected to respond to requests on port 8080 as well
as special health-check requests used by the service to detect if an instance is still running in
order to provide automated failover and instance replacement. The `google.golang.org/appengine`
package provides this for us so we have a slightly different version for Managed VMs:

`app-managed.go`

{{< embed "google-app-engine/app-managed.go" >}}

So now we have three different configurations. We can build and run our app as normal so it can
be executed locally, on a full Compute Engine instance or any other traditional hosting provider
(including EC2, Docker etc...). This build will ignore the code in appengine and appenginevm tagged
files and the `app.yaml` file is meaningless to anything other than the AppEngine platform.

We can also run locally using the [Google AppEngine SDK for Go](https://cloud.google.com/appengine/downloads)
either emulating [AppEngine Classic](https://cloud.google.com/appengine/docs/go/tools/devserver):

    goapp serve

Or [Managed VMs](https://cloud.google.com/appengine/docs/managed-vms/sdk#run-local):

    gcloud config set project [your project id]
    gcloud preview app run .

And of course we can deploy our app to both of these platforms for easy and inexpensive auto-scaling joy.

Depending on what your app actually does it's possible you may need to make other changes to allow
switching between AppEngine provided service such as Datastore and alternative storage implementations
such as MongoDB. A combination of go interfaces and build constraints can make this fairly straightforward
but is outside the scope of this example.  

## [Source Code]({{< source "google-app-engine" >}})

## Maintainers

- [CaptainCodeman](https://github.com/CaptainCodeman)
