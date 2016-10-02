grace [![Build Status](https://secure.travis-ci.org/facebookgo/grace.png)](https://travis-ci.org/facebookgo/grace)
=====

Package grace provides a library that makes it easy to build socket
based servers that can be gracefully terminated & restarted (that is,
without dropping any connections).

It provides a convenient API for HTTP servers including support for TLS,
especially if you need to listen on multiple ports (for example a secondary
internal only admin server).  Additionally it is implemented using the same API
as systemd providing [socket
activation](http://0pointer.de/blog/projects/socket-activation.html)
compatibility to also provide lazy activation of the server.


Usage
-----

Demo HTTP Server with graceful termination and restart:
https://github.com/facebookgo/grace/blob/master/gracedemo/demo.go

1. Install the demo application

        go get github.com/facebookgo/grace/gracedemo

1. Start it in the first terminal

        gracedemo

   This will output something like:

        2013/03/25 19:07:33 Serving [::]:48567, [::]:48568, [::]:48569 with pid 14642.

1. In a second terminal start a slow HTTP request

        curl 'http://localhost:48567/sleep/?duration=20s'

1. In a third terminal trigger a graceful server restart (using the pid from your output):

        kill -USR2 14642

1. Trigger another shorter request that finishes before the earlier request:

        curl 'http://localhost:48567/sleep/?duration=0s'


If done quickly enough, this shows the second quick request will be served by
the new process (as indicated by the PID) while the slow first request will be
served by the first server. It shows how the active connection was gracefully
served before the server was shutdown. It is also showing that at one point
both the new as well as the old server was running at the same time.


Documentation
-------------

`http.Server` graceful termination and restart:
https://godoc.org/github.com/facebookgo/grace/gracehttp

`net.Listener` graceful termination and restart:
https://godoc.org/github.com/facebookgo/grace/gracenet
