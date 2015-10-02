---
title: WebSocket
menu:
  main:
    parent: recipes
---

## Server

`server.go`

```go
package main

import (
	"fmt"

	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
)

func main() {
	e := echo.New()

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	e.Static("/", "public")
	e.WebSocket("/ws", func(c *echo.Context) (err error) {
		ws := c.Socket()
		msg := ""

		for {
			if err = websocket.Message.Send(ws, "Hello, Client!"); err != nil {
				return
			}
			if err = websocket.Message.Receive(ws, &msg); err != nil {
				return
			}
			fmt.Println(msg)
		}
		return
	})

	e.Run(":1323")
}
```

## Client

`index.html`

```html
<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <title>WebSocket</title>
</head>
<body>
    <p id="output"></p>

    <script>
        var loc = window.location;
        var uri = 'ws:';

        if (loc.protocol === 'https:') {
            uri = 'wss:';
        }
        uri += '//' + loc.host;
        uri += loc.pathname + 'ws';

        ws = new WebSocket(uri)

        ws.onopen = function() {
            console.log('Connected')
        }

        ws.onmessage = function(evt) {
            var out = document.getElementById('output');
            out.innerHTML += evt.data + '<br>';
        }

        setInterval(function() {
            ws.send('Hello, Server!');
        }, 1000);
    </script>
</body>
</html>
```

## Output

`Client`

```sh
Hello, Client!
Hello, Client!
Hello, Client!
Hello, Client!
Hello, Client!
```

`Server`

```sh
Hello, Server!
Hello, Server!
Hello, Server!
Hello, Server!
Hello, Server!
```

## [Source Code](https://github.com/labstack/echo/blob/master/recipes/websocket)
