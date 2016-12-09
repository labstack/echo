+++
title = "Auto TLS Example"
description = "Automatic TLS certificates from Let's Encrypt example for Echo"
[menu.main]
  name = "Auto TLS"
  parent = "recipes"
  weight = 2
+++

This recipe shows how to obtain TLS certificates for a domain automatically from
Let's Encrypt. `Echo#StartAutoTLS` accepts an address which should listen on port `443`.

Browse to `https://<your_domain>`. If everything goes fine, you should see a welcome
message with TLS enabled on the website.

> 
- For added security you should specify host policy in auto TLS manager
- To redirect HTTP traffic to HTTPS, you can use [redirect middleware](/middleware/redirect#https-redirect)

## Server

`server.go`

{{< embed "auto-tls/server.go" >}}

## [Source Code]({{< source "auto-tls" >}})

## Maintainers

- [vishr](https://github.com/vishr)
