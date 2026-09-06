[![Latest release](https://img.shields.io/github/v/release/labstack/echo?style=flat-square&label=release&color=00afd1)](https://github.com/labstack/echo/releases)
[![Last commit](https://img.shields.io/github/last-commit/labstack/echo/master?style=flat-square)](https://github.com/labstack/echo/commits/master)
[![Sourcegraph](https://sourcegraph.com/github.com/labstack/echo/-/badge.svg?style=flat-square)](https://sourcegraph.com/github.com/labstack/echo?badge)
[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/labstack/echo/v5)
[![Go Report Card](https://goreportcard.com/badge/github.com/labstack/echo?style=flat-square)](https://goreportcard.com/report/github.com/labstack/echo)
[![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/labstack/echo/echo.yml?style=flat-square)](https://github.com/labstack/echo/actions)
[![Codecov](https://img.shields.io/codecov/c/github/labstack/echo.svg?style=flat-square)](https://codecov.io/gh/labstack/echo)
[![Forum](https://img.shields.io/badge/community-forum-00afd1.svg?style=flat-square)](https://github.com/labstack/echo/discussions)
[![Twitter](https://img.shields.io/badge/twitter-@labstack-55acee.svg?style=flat-square)](https://twitter.com/labstack)
[![License](https://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/labstack/echo/master/LICENSE)

## Echo

High performance, extensible, minimalist Go web framework.

Echo is built on Go's standard `net/http` — and interoperates with it via `echo.WrapHandler` / `echo.WrapMiddleware` — adding the parts the standard library leaves to you: a fast radix-tree router, request binding (with a pluggable validator), a deep middleware ecosystem, and centralized error handling. Actively maintained, with `v5` as the current release line (see badges above for the latest version and most recent commit).

* [Official website](https://echo.labstack.com)
* [Quick start](https://echo.labstack.com/docs/quick-start)
* [Middlewares](https://echo.labstack.com/docs/category/middleware)

Help and questions: [Github Discussions](https://github.com/labstack/echo/discussions)

### Feature Overview

- Optimized HTTP router which smartly prioritize routes
- Build robust and scalable RESTful APIs
- Group APIs
- Extensible middleware framework
- Define middleware at root, group or route level
- Data binding for JSON, XML and form payload
- Handy functions to send variety of HTTP responses
- Centralized HTTP error handling
- Template rendering with any template engine
- Define your format for the logger
- Highly customizable
- Automatic TLS via Let’s Encrypt
- HTTP/2 support

## Sponsors

<div>
  <a href="https://encore.dev" style="display: inline-flex; align-items: center; gap: 10px">
    <img src="https://user-images.githubusercontent.com/78424526/214602214-52e0483a-b5fc-4d4c-b03e-0b7b23e012df.svg" height="28px" alt="encore icon"></img>
  <b>Encore – the platform for building Go-based cloud backends</b>
    </a>
</div>
<br/>

Click [here](https://github.com/sponsors/labstack) for more information on sponsorship.

## [Guide](https://echo.labstack.com/guide)

### Supported Echo versions

- Latest major version of Echo is `v5` as of 2026-01-18.
  - See [API_CHANGES_V5.md](./API_CHANGES_V5.md) for public API changes between `v4` and `v5`, notes on upgrading.
- Echo `v4` is supported with **security*** updates and **bug** fixes until **2026-12-31**

See [ROADMAP.md](./ROADMAP.md) for where Echo is heading and the version support policy.

### Installation

```sh
// go get github.com/labstack/echo/{version}
go get github.com/labstack/echo/v5
```

Latest version of Echo supports last four Go major [releases](https://go.dev/doc/devel/release) and might work with
older versions.

### Example

```go
package main

import (
  "github.com/labstack/echo/v5"
  "github.com/labstack/echo/v5/middleware"
  "log/slog"
  "net/http"
)

func main() {
  // Echo instance
  e := echo.New()

  // Middleware
  e.Use(middleware.RequestLogger()) // use the RequestLogger middleware with slog logger
  e.Use(middleware.Recover())       // recover panics as errors for proper error handling

  // Routes
  e.GET("/", hello)

  // Start server
  if err := e.Start(":8080"); err != nil {
    slog.Error("failed to start server", "error", err)
  }
}

// Handler
func hello(c *echo.Context) error {
  return c.String(http.StatusOK, "Hello, World!")
}
```

# Official middleware repositories

Following list of middleware is maintained by Echo team.

| Repository                                                                               | Description                                                                                                                                                  |
|------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [github.com/labstack/echo-jwt](https://github.com/labstack/echo-jwt)                     | [JWT](https://github.com/golang-jwt/jwt) middleware                                                                                                          | 
| [github.com/labstack/echo-contrib](https://github.com/labstack/echo-contrib)             | [casbin](https://github.com/casbin/casbin), [gorilla/sessions](https://github.com/gorilla/sessions), [pprof](https://pkg.go.dev/net/http/pprof)) middlewares | 
| [github.com/labstack/echo-opentelemetry](https://github.com/labstack/echo-opentelemetry) | [OpenTelemetry](https://opentelemetry.io/) middleware for tracing and metrics                                                                                |
| [github.com/labstack/echo-prometheus](https://github.com/labstack/echo-prometheus)       | [Prometheus](https://github.com/prometheus/client_golang/) middleware for Echo                                                                               |

# Third-party middleware repositories

Be careful when adding 3rd party middleware. Echo teams does not have time or manpower to guarantee safety and quality
of middlewares in this list.

| Repository                                                                                           | Description                                                                                                                                                                                              |
|------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [oapi-codegen/oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)                            | Automatically generate RESTful API documentation with [OpenAPI](https://swagger.io/specification/) Client and Server Code Generator                                                                      |
| [github.com/swaggo/echo-swagger](https://github.com/swaggo/echo-swagger)                             | Automatically generate RESTful API documentation with [Swagger](https://swagger.io/) 2.0.                                                                                                                |
| [github.com/ziflex/lecho](https://github.com/ziflex/lecho)                                           | [Zerolog](https://github.com/rs/zerolog) logging library wrapper for Echo logger interface.                                                                                                              |
| [github.com/brpaz/echozap](https://github.com/brpaz/echozap)                                         | Uber´s [Zap](https://github.com/uber-go/zap) logging library wrapper for Echo logger interface.                                                                                                          |
| [github.com/samber/slog-echo](https://github.com/samber/slog-echo)                                   | Go [slog](https://pkg.go.dev/golang.org/x/exp/slog) logging library wrapper for Echo logger interface.                                                                                                   |
| [github.com/darkweak/souin/plugins/echo](https://github.com/darkweak/souin/tree/master/plugins/echo) | HTTP cache system based on [Souin](https://github.com/darkweak/souin) to automatically get your endpoints cached. It supports some distributed and non-distributed storage systems depending your needs. |
| [github.com/mikestefanello/pagoda](https://github.com/mikestefanello/pagoda)                         | Rapid, easy full-stack web development starter kit built with Echo.                                                                                                                                      |
| [github.com/go-woo/protoc-gen-echo](https://github.com/go-woo/protoc-gen-echo)                       | ProtoBuf generate Echo server side code                                                                                                                                                                  |

Please send a PR to add your own library here.

## Contribute

**Use issues for everything**

- For a small change, just send a PR.
- For bigger changes open an issue for discussion before sending a PR.
- PR should have:
  - Test case
  - Documentation
  - Example (If it makes sense)
- You can also contribute by:
  - Reporting issues
  - Suggesting new features or enhancements
  - Improve/fix documentation

## Credits

- [Vishal Rana](https://github.com/vishr) (Author)
- [Nitin Rana](https://github.com/nr17) (Consultant)
- [Roland Lammel](https://github.com/lammel) (Maintainer)
- [Martti T.](https://github.com/aldas) (Maintainer)
- [Pablo Andres Fuente](https://github.com/pafuent) (Maintainer)
- [Contributors](https://github.com/labstack/echo/graphs/contributors)

## License

[MIT](https://github.com/labstack/echo/blob/master/LICENSE)


## 🌐 Web Resources & Interactive Index
- [CATEGORY TURN BASED](https://quizverses.github.io/category-turn-based.html)
- [KAWAII REALM ADVENTURE](https://quizverses.pages.dev/kawaii-realm-adventure.html)
- [IDLE INVENTOR](https://thequizzone.pages.dev/idle-inventor.html)
- [SUDOKU VAULT](https://quizverses.github.io/sudoku-vault.html)
- [CATEGORY BUBBLE SHOOTER](https://quizverses-9d2f2.web.app/category-bubble-shooter.html)
- [CATEGORY TITANIUMNETWORK](https://quizverses-9d2f2.web.app/category-titaniumnetwork.html)
- [BADLANDS HERO](https://studyplaying.github.io/badlands-hero.html)
- [ANIMAL BLOCKS](https://studyplaying.github.io/animal-blocks.html)
- [ATHENA MATCH 2](https://quizverses.github.io/athena-match-2.html)
- [BUBBLE SHOOTER POP](https://studyplaying.github.io/bubble-shooter-pop.html)
- [DIRTY MONEY THE RICH GET RICH](https://quizverses.github.io/dirty-money-the-rich-get-rich.html)
- [MR BEAN JUMP](https://studyquests.github.io/mr-bean-jump.html)
- [DRESS PRINCESS](https://quizverses.pages.dev/dress-princess.html)
- [BUS COLOR JAM](https://studyquests.pages.dev/bus-color-jam.html)
- [CHICKEN WILD RUN](https://quizverses.pages.dev/chicken-wild-run.html)
- [ANOMALY CONTENT RECORD](https://quizverses.github.io/anomaly-content-record.html)
- [GOMU GOMAN](https://studyquests.pages.dev/gomu-goman.html)
- [LABUBU AND TREASURES FUN ADVENTURE](https://quizverses.github.io/labubu-and-treasures-fun-adventure.html)
- [MY PERFECT MINE](https://quizverses-9d2f2.web.app/my-perfect-mine.html)
- [MY TINY LAND](https://studyquests.github.io/my-tiny-land.html)
- [SHAPE TRANSFORMING SHIFTING RUN](https://studyplaying.github.io/shape-transforming-shifting-run.html)
- [COUNTRY LIFE MEADOWS](https://studyquests.pages.dev/country-life-meadows.html)
- [ZINDEX](https://quizverses-9d2f2.web.app/zindex.html)
- [COSMIC AVIATOR](https://studyplaying.github.io/cosmic-aviator.html)
- [CATEGORY STRATEGY 2](https://studyplaying.github.io/category-strategy-2.html)
- [SPACE SURVIVOR](https://studyplaying.github.io/space-survivor.html)
- [CATEGORY CUTE62](https://studyquests.pages.dev/category-cute62.html)
- [GARAGE MASTER NUTS AND BOLTS](https://studyquests.pages.dev/garage-master-nuts-and-bolts.html)
- [KAWAII FRIENDS TILES MATCHER](https://studyquests.pages.dev/kawaii-friends-tiles-matcher.html)
- [RACE CLICKER](https://studyquesthub.web.app/race-clicker.html)
- [VALENTINE HIDDEN HEART](https://quizverses.pages.dev/valentine-hidden-heart.html)
- [TOWER WARS ARENA](https://quizverses-9d2f2.web.app/tower-wars-arena.html)
- [CAR PARKING SIMULATOR](https://quizverses.pages.dev/car-parking-simulator.html)
- [CATEGORY COOKING](https://studyquests.pages.dev/category-cooking.html)
- [FRUIT JAM MERGE PUZZLE GAME](https://quizverses.pages.dev/fruit-jam-merge-puzzle-game.html)
- [HIDE MOODENG HIPPO](https://studyquesthub.web.app/hide-moodeng-hippo.html)
- [CRYPTOWORD](https://studyplaying.github.io/cryptoword.html)
- [BATTLE RACING STARS](https://quizverses.pages.dev/battle-racing-stars.html)
- [AIDAN IN DANGER](https://studyquests.pages.dev/aidan-in-danger.html)
- [BLOOM SORT 2 BEE PUZZLE](https://studyquesthub.web.app/bloom-sort-2-bee-puzzle.html)
- [PING PONG AIR](https://studyplaying.github.io/ping-pong-air.html)
- [CATEGORY TOP DOWN251](https://quizverses.github.io/category-top-down251.html)
- [CATEGORY TOWER DEFENSE 3](https://quizverses.github.io/category-tower-defense-3.html)
- [VIBRANT HEARTS GLAMOUR VS PUNK](https://studyquesthub.web.app/vibrant-hearts-glamour-vs-punk.html)
- [OBBY DRAW TO ESCAPE](https://quizverses.pages.dev/obby-draw-to-escape.html)
- [CATEGORY SURVIVAL365](https://quizverses.github.io/category-survival365.html)
- [FASHION CHALLENGE CATWALK RUN](https://quizverses.pages.dev/fashion-challenge-catwalk-run.html)
- [GUESS THE DRAWING](https://quizverses-9d2f2.web.app/guess-the-drawing.html)
- [PRINCESS ROYAL WEDDING](https://quizverses.github.io/princess-royal-wedding.html)
- [NONOGRAM DAILY](https://studyquesthub.web.app/nonogram-daily.html)
- [OREPLICATION](https://studyplaying.github.io/oreplication.html)
- [GROW A GARDEN ONLINE OFFLINE](https://studyplaying.github.io/grow-a-garden-online-offline.html)
- [CATEGORY EDUCATIONAL](https://studyquests.pages.dev/category-educational.html)
- [MAHJONG CONNECT MAJONG CLASS](https://studyquests.pages.dev/mahjong-connect-majong-class.html)
- [GOOBER DASH](https://studyquests.pages.dev/goober-dash.html)
- [INDEX8](https://studyquests.pages.dev/index8.html)
- [BIG BAD APE](https://quizverses.pages.dev/big-bad-ape.html)
- [SUSHI PUZZLE](https://quizverses.pages.dev/sushi-puzzle.html)
- [BROOMCRAFT MYSTIC EVASION](https://studyplaying.github.io/broomcraft-mystic-evasion.html)
- [IDLE BARBER SHOP](https://quizverses.pages.dev/idle-barber-shop.html)
- [CATEGORY CASUAL 6](https://studyquests.pages.dev/category-casual-6.html)
- [HEROIC KNIGHT](https://quizverses.pages.dev/heroic-knight.html)
- [MOVE EMOJI](https://quizverses.pages.dev/move-emoji.html)
- [SERIOUS HEAD 2](https://quizverses.github.io/serious-head-2.html)
- [FASHION VALKYRIES SAGA OF STYLE](https://quizverses.pages.dev/fashion-valkyries-saga-of-style.html)
- [BOOM STICK BAZOOKA](https://studyquesthub.web.app/boom-stick-bazooka.html)
- [BUBBLE SHOOTER REMASTERED](https://quizverses-9d2f2.web.app/bubble-shooter-remastered.html)
- [BATTLE ARENA](https://quizverses.pages.dev/battle-arena.html)
- [DRIVE RACE CRASH](https://quizverses.github.io/drive-race-crash.html)
- [TUNG TUNG SAHUR OBBY CHALLENGE](https://studyquesthub.web.app/tung-tung-sahur-obby-challenge.html)
- [CATEGORY SECURLY](https://studyplaying.github.io/category-securly.html)
- [FIERCE BATTLE BREAKOUT](https://studyquesthub.web.app/fierce-battle-breakout.html)
- [FESTIVAL VIBES MAKEUP](https://studyplaying.github.io/festival-vibes-makeup.html)
- [MR BOUNCE](https://quizverses.pages.dev/mr-bounce.html)
- [RENT OUT LANDLORD TYCOON](https://quizverses-9d2f2.web.app/rent-out-landlord-tycoon.html)
- [DRAW TO FLY](https://quizverses.github.io/draw-to-fly.html)
- [GRANDMAS LAST STAND](https://studyplaying.github.io/grandmas-last-stand.html)
- [BOLTS AND NUTS SORTING](https://studyplaying.github.io/bolts-and-nuts-sorting.html)
- [CATEGORY HORROR](https://studyplaying.github.io/category-horror.html)
- [CATEGORY DRESS UP97](https://studyquests.pages.dev/category-dress-up97.html)
- [DEVIL DASH](https://quizverses.github.io/devil-dash.html)
- [NEON BLAST](https://studyplaying.github.io/neon-blast.html)
- [TOWER OF HELL OBBY BLOX](https://quizverses.pages.dev/tower-of-hell-obby-blox.html)
- [FOOTBALL DUEL](https://quizverses.github.io/football-duel.html)
- [SMASH THE BOTTLE](https://quizverses.github.io/smash-the-bottle.html)
- [DELIVERY NOW](https://quizverses.github.io/delivery-now.html)
- [GROW A GARDEN 3D](https://quizverses.github.io/grow-a-garden-3d.html)
- [MEGA RAMP CAR STUNTS](https://quizverses.github.io/mega-ramp-car-stunts.html)
- [SPIDER ROPE HERO CITY FIGHT](https://studyplaying.github.io/spider-rope-hero-city-fight.html)
- [BOMB EVOLUTION](https://studyplaying.github.io/bomb-evolution.html)
- [BUS DRIVER SIMULATOR 3D](https://quizverses.pages.dev/bus-driver-simulator-3d.html)
- [WAVE ROAD 3D](https://quizverses-9d2f2.web.app/wave-road-3d.html)
- [JELLY BELLY MAKE THE ELEPHANT](https://quizverses.github.io/jelly-belly-make-the-elephant.html)
- [CATEGORY GUN238](https://studyquests.pages.dev/category-gun238.html)
- [REAL RACING 3D](https://studyplaying.github.io/real-racing-3d.html)
- [BLACK PINK BLACK FRIDAY FEVER](https://quizverses.pages.dev/black-pink-black-friday-fever.html)
- [MAX MIXED COCKTAILS](https://studyquests.pages.dev/max-mixed-cocktails.html)
- [CATEGORY RPG](https://studyplaying.github.io/category-rpg.html)
- [NUBIK COURIER AN OPEN WORLD](https://quizverses.pages.dev/nubik-courier-an-open-world.html)
- [ERASE THE EXTRA ELEMENT](https://studyplaying.github.io/erase-the-extra-element.html)
- [PULL THE PINS](https://quizverses.pages.dev/pull-the-pins.html)
- [ROYAL REBELLION PUNK MAGIC](https://quizverses-9d2f2.web.app/royal-rebellion-punk-magic.html)
- [JET FIGHTER AIRPLANE RACING](https://studyplaying.github.io/jet-fighter-airplane-racing.html)
- [PIECE OF CAKE MERGE AND BAKE](https://studyquesthub.web.app/piece-of-cake-merge-and-bake.html)
- [HOOP WORLD 3D](https://studyquests.pages.dev/hoop-world-3d.html)
- [SURVIVAL IN AREA 51](https://studyquests.pages.dev/survival-in-area-51.html)
- [WALL HOP](https://quizverses.github.io/wall-hop.html)
- [CHROMA TREK](https://studyquests.pages.dev/chroma-trek.html)
- [AYLA WORLD PRINCESS LIFE](https://studyplaying.github.io/ayla-world-princess-life.html)
- [GRILL IT ALL](https://studyquests.pages.dev/grill-it-all.html)
- [BUILD A ROLLERCOASTER SIMULATOR](https://studyquests.pages.dev/build-a-rollercoaster-simulator.html)
- [BLUE HEDGEHOG HILL DASH RIDE](https://studyplaying.github.io/blue-hedgehog-hill-dash-ride.html)
- [CHESS DUEL](https://studyplaying.github.io/chess-duel.html)
- [NUWPYS ADVENTURE](https://studyplaying.github.io/nuwpys-adventure.html)
- [PAWS OFF MY CLUES](https://quizverses.pages.dev/paws-off-my-clues.html)
- [HIDDEN OBJECTS ISLAND SECRETS](https://quizverses.github.io/hidden-objects-island-secrets.html)
- [MERGE HAVEN](https://quizverses.github.io/merge-haven.html)
- [AGENTS IO](https://studyplaying.github.io/agents-io.html)
- [ZOMBIE CHASE](https://quizverses.pages.dev/zombie-chase.html)
- [PUZZLE BLOCKS CLASSIC](https://studyquests.pages.dev/puzzle-blocks-classic.html)
- [CATEGORY ADVENTURE](https://studyquests.pages.dev/category-adventure.html)
- [FIND OBJECTS HIDDEN ITEM](https://studyquests.pages.dev/find-objects-hidden-item.html)
- [JIGSAW M](https://quizverses.github.io/jigsaw-m.html)
- [MERGE 2048 CAKE](https://quizverses.github.io/merge-2048-cake.html)
- [SUPER STOCK STACK](https://studyplaying.github.io/super-stock-stack.html)
- [CATEGORY EDUCATIONAL](https://learnquester.github.io/category-educational.html)
- [SQUISHY TABA PAW ASMR](https://studyplayings.web.app/squishy-taba-paw-asmr.html)
- [SPACE SHOOTER SPEED TYPING CHALLENGE](https://quizverses.github.io/space-shooter-speed-typing-challenge.html)
- [SPRING TILE MASTER](https://studyplayings.web.app/spring-tile-master.html)
- [SEA MONSTERS MAHJONG](https://studyplaying.github.io/sea-monsters-mahjong.html)
