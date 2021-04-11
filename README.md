<div align="center">
  <h1>Zeropino</h1>

  [![Test and Build Go Project][github-actions-svg]][github-actions]

</div>

Zeropino package provides a custom JSON format as a default for [zerolog][zerolog-github] logger. This log format is inspired by [Mia-Platform logging guidelines][logging-guidelines] and [glogger][glogger] logger.

In addition, it draws similarities to the structure adopted by [Pino][pino-github] logger for Node JS. This allows to parse Zerolog logs by prettifiers, such as [pino-pretty][pino-pretty] library, simplifing their inspection, while preserving logger efficiency during services opererations.

Besides logger customization, Zeropino package offers middleware functions for the following web frameworks:

- [Fiber][fiber-github]
- [Gorilla Mux][gorilla-mux-github]

These should help integrate the custom logger within a service.

## Installation

    go get -u github.com/danibix95/zeropino

## Getting Started

### Basic Initialization and Usage

Create a `zerolog` logger with the added fields specified by Zeropino.

```go
package main

import "github.com/danibix95/zeropino"

func main() {
    logger, err := zeropino.Init(zeropino.InitOptions{})
    // handle err here

    logger.Warn().Msg("there is no real going back")
}

// Output: {"level":"40","pid":12739,"hostname":"bag-end","time":1618003000857,"msg":"there is no real going back"}
```
For additional details on how to use and customize the logger, please read [`zerolog` documentation](https://pkg.go.dev/github.com/rs/zerolog).

### Custom Fields
Below are reported the custom JSON properties added or modified with respect to the default logger provided by `zerolog`:
- `level [string]` represents log message level. It can get a value from 10 to 70, increasing of 10 steps at each level.
- `pid [int]` the process id that is running the go program
- `hostname [string]` the hostname which is running the go program
- `time int` the time when the log is created, as a Unix Timestamp in milliseconds
- `msg [string]` the actual message (as same as `zerolog`)

### Init Options
There are three main options to customize the logger:
- `Level [string]` select logger level - it can be one of these values, starting from the lowest to the highest:
  - `trace`
  - `debug`
  - `info`
  - `warn`
  - `error`
  - `fatal`
  - `panic`
  - `silent` (no log is produced using this level)
- `DisableTimeMs [bool]` select whether the Unix timestamp should be in seconds rather than default format of milliseconds
- `Writer [io.Writer]` define which writer should be used to produce the logs

## Fiber Middleware

Here is provided an example of how to use the Zeropino `RequestLogger` middleware for `fiber`:

```go
package main

import (
  "github.com/gofiber/fiber/v2"

  zp "github.com/danibix95/zeropino"
  zpfiber "github.com/danibix95/zeropino/middlewares/fiber"
)

func main() {
  app := fiber.New()

  logger, _ := zp.Init(zp.InitOptions{Level: "trace"})

  // add the zeropino request logger middleware
  app.Use(zpfiber.RequestLogger(logger))

  app.Get("/welcome", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"msg": "Hello, World!"})
  })

  if err := app.Listen(":3000"); err != nil {
    logger.Fatal().Err(err).Msg("terminating app")
  }

  app.Listen(":3000")
}
```

The `RequestLogger` middleware store a logger specific to each request in the fiber [Locals](https://docs.gofiber.io/api/ctx#locals) request storage. It is possible to retrieve it with `ReqLogger` method, so that it is not necessary to create a new logger withing each handler.  
Moreover, it is possible to change the stored logger using the methods and `WithLogger`. This enables the possibility to further customize it.

For example:
```go
// add another middleware that use injected logger an further customize it
app.Use(func (c *fiber) error {
  // get existing logger
  // Note: if no logger exists, return a new zerolog Logger with
  // the default zeropino configuration (level: info, writer: os.Stdout)
  reqLogger := zpfiber.ReqLogger(c)

  quoteLogger := reqLogger.With().Str("quote", "This is the way").Logger()

  // modify the logger stored in the fiber context
  zpfiber.WithLogger(c, &quoteLogger)

  return c.Next()
})

app.Get("/quote", func (c *fiber.Ctx) error {
  // get existing logger and log current fields
  zpfiber.ReqLogger(c).Info().Send()

  return c.JSON(fiber.Map{"msg": "a quote has been logged"})
})
```

## Gorilla Mux Middleware

Here is provided an example of how to use the Zeropino `RequestLogger` middleware for `gorilla mux`:

```go
package main

import (
  "net/http"

  "github.com/gorilla/mux"
  "github.com/danibix95/zeropino"
  zpmux "github.com/danibix95/zeropino/middlewares/gorillamux"
)

func main() {
  logger, _ := zeropino.Init(zeropino.InitOptions{Level: "trace"})

  router := mux.NewRouter()

  // add the zeropino request logger middleware
  router.Use(zpmux.RequestLogger(logger, []string{"/-/"}))

  router.HandleFunc("/welcome", func(w http.ResponseWriter, req *http.Request) {
    w.Write("Hello, World!")
  })

  server := &http.Server{
    Addr:    "0.0.0.0:3000",
    Handler: router,
  }

  if err := server.ListenAndServe(); err != nil {
    logger.Error().Err(err).Send()
  }
}
```

[github-actions]: https://github.com/danibix95/zerolog-mia/actions/workflows/go.yml
[github-actions-svg]: https://github.com/danibix95/zerolog-mia/actions/workflows/go.yml/badge.svg?branch=main

[zerolog-github]: https://github.com/rs/zerolog
[logging-guidelines]: https://docs.mia-platform.eu/docs/getting_started/monitoring-dashboard/dev_ops_guide/log#json-logging-format
[glogger]: https://github.com/mia-platform/glogger
[pino-github]: https://github.com/pinojs/pino
[pino-pretty]: https://github.com/pinojs/pino-pretty
[fiber-github]: https://github.com/gofiber/fiber
[gorilla-mux-github]: https://github.com/gorilla/mux
