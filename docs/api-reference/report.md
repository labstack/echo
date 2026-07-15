# Sourcey API Reference for labstack/echo

## Overview

Generated comprehensive API reference documentation for [labstack/echo](https://github.com/labstack/echo), a high-performance, minimalist Go web framework. The docs are generated from the actual Go source code using Sourcey's godoc adapter.

## What was built

- **87 public APIs documented** across3 packages
- **Full-text search** across all API surfaces
- **Structured navigation** with package-level organization
- **GitHub Actions workflow** for automatic deployment to GitHub Pages

## Packages documented

1. **echo** - Core framework types and functions (Echo struct, Context interface, routing, etc.)
2. **echotest** - Testing utilities for echo applications
3. **middleware** - HTTP middleware functions (CORS, Logger, Recover, etc.)

## Maintainer-facing gaps

1. **Missing examples**: The godoc snapshot includes test examples, but more real-world usage examples would improve the docs
2. **Configuration reference**: The docs don't include configuration options for middleware (e.g., CORS options, rate limiter settings)
3. **Error handling**: The error handling patterns could be documented more comprehensively
4. **Migration guide**: No migration guide between echo versions
5. **Performance tips**: No documentation on performance optimization or best practices

## How to use

1. Visit [https://patrick6x6.github.io/echo/](https://patrick6x6.github.io/echo/)
2. Browse the API reference by package
3. Use the search functionality to find specific types or functions
4. Click on type names to see detailed documentation

## Deployment

The docs are deployed to GitHub Pages at [https://patrick6x6.github.io/echo/](https://patrick6x6.github.io/echo/).

A PR has been submitted to the echo project to add these docs to the project's official documentation: [https://github.com/labstack/echo/pull/3045](https://github.com/labstack/echo/pull/3045)

Once merged, the docs will be live at `labstack.github.io/echo`.

## Technical details

- **Sourcey version**: 3.6.5
- **Go version**: 1.26.5
- **Adapter**: godoc (snapshot mode)
- **Output format**: Static HTML with CSS/JS
- **Total size**: 2.4MB
