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
- [SPRUNKI 3D SHOOTER](https://studyplayings.pages.dev/sprunki-3d-shooter.html)
- [CATEGORY BUSINESS135](https://iskillquest.pages.dev/category-business135.html)
- [CATEGORY ARENA255](https://studyplayings.pages.dev/category-arena255.html)
- [ALCHEMY PUZZLE](https://studyplaying.github.io/alchemy-puzzle.html)
- [DYNAMONS 11](https://learnquester.github.io/dynamons-11.html)
- [LITTLE LILY HALLOWEEN PREP](https://studyplayings.web.app/little-lily-halloween-prep.html)
- [INDEX9](https://studyplayings.pages.dev/index9.html)
- [PIXEL SHOOT](https://studyplaying.github.io/pixel-shoot.html)
- [CYBERPUNK CITY FASHION](https://studyquests.pages.dev/cyberpunk-city-fashion.html)
- [CATEGORY SKILL256](https://studyplayings.pages.dev/category-skill256.html)
- [BOBB S WORLD](https://studyquests.pages.dev/bobb-s-world.html)
- [TWO BLOCKS](https://studyquesthub.web.app/two-blocks.html)
- [CATEGORY 2D1 060](https://studyplayings.web.app/category-2d1-060.html)
- [INDIAN SUV OFFROAD SIMULATOR](https://studyquests.pages.dev/indian-suv-offroad-simulator.html)
- [STEAM SORTER](https://studyquesthub.web.app/steam-sorter.html)
- [CATEGORY MERGE GAME](https://studyplayings.pages.dev/category-merge-game.html)
- [SINGLE LINE PUZZLE DRAWING](https://studyplayings.web.app/single-line-puzzle-drawing.html)
- [CATEGORY SHOOTER 3](https://studyplaying.github.io/category-shooter-3.html)
- [MERGE NUMBERS](https://studyquests.pages.dev/merge-numbers.html)
- [CATEGORY ONE BUTTON84](https://studyplayings.pages.dev/category-one-button84.html)
- [THE EARTH EVOLUTION](https://studyplayings.web.app/the-earth-evolution.html)
- [BLOCKAPOLYPSE ZOMBIE SHOOTER](https://studyquesthub.web.app/blockapolypse-zombie-shooter.html)
- [TIED UP](https://studyplayings.web.app/tied-up.html)
- [CATEGORY PUZZLE 5](https://studyplayings.pages.dev/category-puzzle-5.html)
- [CATEGORY STICKMAN](https://studyquests.pages.dev/category-stickman.html)
- [SECRETS OF CHARMLAND](https://studyplaying.github.io/secrets-of-charmland.html)
- [HIDDEN OBJECT CLUES AND MYSTERIES](https://studyplayings.web.app/hidden-object-clues-and-mysteries.html)
- [RAGDOLL JUMP](https://studyplayings.web.app/ragdoll-jump.html)
- [CATEGORY MAKEUP](https://studyplayings.pages.dev/category-makeup.html)
- [NEKOS ADVENTURE](https://studyplayings.web.app/nekos-adventure.html)
- [CATEGORY WAR137](https://studyplayings.pages.dev/category-war137.html)
- [ICE FISHING 3D](https://studyplaying.github.io/ice-fishing-3d.html)
- [THE PRISM CITY DETECTIVES](https://studyplayings.web.app/the-prism-city-detectives.html)
- [CATEGORY MMO24](https://studyplayings.pages.dev/category-mmo24.html)
- [TAP TO COLOR PAINTING BOOK](https://studyplaying.github.io/tap-to-color-painting-book.html)
- [ROBYBOX SPACE STATION WAREHOUSE](https://studyplaying.github.io/robybox-space-station-warehouse.html)
- [EUROPE AT WAR](https://studyplayings.web.app/europe-at-war.html)
- [CATEGORY ROGUELIKE38](https://studyplayings.pages.dev/category-roguelike38.html)
- [PHOTO BLOCK JOURNEY](https://studyquests.pages.dev/photo-block-journey.html)
- [BLUE HEDGEHOG HILL DASH RIDE](https://quizverses.github.io/blue-hedgehog-hill-dash-ride.html)
- [CATEGORY MOUSE1 707](https://quizverses.github.io/category-mouse1-707.html)
- [MOTO TRAFFIC RIDER](https://quizverses.pages.dev/moto-traffic-rider.html)
- [CATEGORY SOCCER](https://quizverses-9d2f2.web.app/category-soccer.html)
- [STICKMAN KOMBAT 2D](https://studyplayings.pages.dev/stickman-kombat-2d.html)
- [GUN RUSH](https://quizverses.pages.dev/gun-rush.html)
- [SMART DOTS RELOADED](https://studyquesthub.web.app/smart-dots-reloaded.html)
- [THE GRENCH COUPLE HOLIDAY DRESS UP](https://quizverses-9d2f2.web.app/the-grench-couple-holiday-dress-up.html)
- [BRAIN FIND CAN YOU FIND IT](https://quizverses.pages.dev/brain-find-can-you-find-it.html)
- [BUBBLE SHOOTER VINTAGE](https://studyquests.github.io/bubble-shooter-vintage.html)
- [TOW N GO](https://quizverses-9d2f2.web.app/tow-n-go.html)
- [LOST IN THE FOREST](https://studyplayings.web.app/lost-in-the-forest.html)
- [KALULU TANHULU ASMR MUKBANG](https://quizverses.pages.dev/kalulu-tanhulu-asmr-mukbang.html)
- [STICKER JAM PEEL OFF MATCH](https://quizverses.github.io/sticker-jam-peel-off-match.html)
- [DETECTIVE LOGIC PUZZLES](https://studyplaying.github.io/detective-logic-puzzles.html)
- [CATEGORY SANDBOX41](https://studyplayings.web.app/category-sandbox41.html)
- [PUZZLE GAMES OUTING DAY](https://studyquests.github.io/puzzle-games-outing-day.html)
- [SECRET ROOMS](https://studyplaying.github.io/secret-rooms.html)
- [STUDENT AND TEACHER](https://studyquests.pages.dev/student-and-teacher.html)
- [SHELF SHIFT MATCH](https://studyquests.github.io/shelf-shift-match.html)
- [MAHJONG PET QUEST](https://quizverses.pages.dev/mahjong-pet-quest.html)
- [COZY GARDEN IDLE](https://studyplayings.web.app/cozy-garden-idle.html)
- [CATEGORY MOUSE1 697](https://studyplayings.pages.dev/category-mouse1-697.html)
- [SORT WORKS NUTS ORDER](https://studyplaying.github.io/sort-works-nuts-order.html)
- [MEGA RAMP CAR STUNTS](https://studyplaying.github.io/mega-ramp-car-stunts.html)
- [CATEGORY AGILITY](https://thelearnquester.web.app/category-agility.html)
- [CANDY JEWELS](https://studyquests.github.io/candy-jewels.html)
- [HIGH HEELS 2](https://studyquesthub.web.app/high-heels-2.html)
- [PARIS KISS](https://studyplaying.github.io/paris-kiss.html)
- [BRAINROT EVOLUTION GAME](https://quizverses-9d2f2.web.app/brainrot-evolution-game.html)
- [PHONE CASE DIY 5](https://quizverses-9d2f2.web.app/phone-case-diy-5.html)
- [CATEGORY SPORTS](https://studyplayings.pages.dev/category-sports.html)
- [CATEGORY DRESS UP97](https://studyplaying.github.io/category-dress-up97.html)
- [CATEGORY JIGSAW](https://studyquests.pages.dev/category-jigsaw.html)
- [FRUIT MATCH JUICY PUZZLE](https://quizverses.pages.dev/fruit-match-juicy-puzzle.html)
- [TUNG TUNG SAHUR IN GEOMETRY DASH](https://studyquests.pages.dev/tung-tung-sahur-in-geometry-dash.html)
- [SQUID GAME HUNTER](https://quizverses.pages.dev/squid-game-hunter.html)
- [CRASH THE ROBOT](https://quizverses-9d2f2.web.app/crash-the-robot.html)
- [SOCCER DASH](https://studyplayings.web.app/soccer-dash.html)
- [MAZE ESCAPE CHALLENGE](https://quizverses.pages.dev/maze-escape-challenge.html)
- [SHANGHAI CHEF](https://studyplaying.github.io/shanghai-chef.html)
- [PLANETARIUM 2](https://studyplaying.github.io/planetarium-2.html)
- [CATEGORY SHOOTER](https://quizverses-9d2f2.web.app/category-shooter.html)
- [WAR STATE IO CONQUER BATTLES](https://learnquester.github.io/war-state-io-conquer-battles.html)
- [WORMS ZONE](https://studyplayings.web.app/worms-zone.html)
- [TREASURE SEEKER](https://quizverses-9d2f2.web.app/treasure-seeker.html)
- [BASE JUMP WINGSUIT FLYING](https://studyquests.pages.dev/base-jump-wingsuit-flying.html)
- [1945 AIR FORCE AIRPLANE](https://studyquests.pages.dev/1945-air-force-airplane.html)
- [PLANE CRASH RAGDOLL SIMULATOR](https://studyplaying.github.io/plane-crash-ragdoll-simulator.html)
- [JINN DASH](https://quizverses.pages.dev/jinn-dash.html)
- [LADDER MASTER COLOR RUN](https://quizverses-9d2f2.web.app/ladder-master-color-run.html)
- [SURVIVE LAVA FOR BRAINROTS](https://quizverses-9d2f2.web.app/survive-lava-for-brainrots.html)
- [CATEGORY PUZZLE 3](https://studyquests.pages.dev/category-puzzle-3.html)
- [MERGE BRAINROT](https://quizverses-9d2f2.web.app/merge-brainrot.html)
- [BACKYARD DIG HOLE 3D SIMULATOR](https://quizverses.pages.dev/backyard-dig-hole-3d-simulator.html)
- [CATEGORY SECURLY BYPASS](https://quizverses.pages.dev/category-securly-bypass.html)
- [CATEGORY OBSTACLE](https://quizverses-9d2f2.web.app/category-obstacle.html)
- [CATCH THE GOOSE](https://studyquesthub.web.app/catch-the-goose.html)
- [ENCHANTED EASTER ADVENTURE](https://studyquests.github.io/enchanted-easter-adventure.html)
- [LABUBU COLORING ADVENTURE](https://quizverses.pages.dev/labubu-coloring-adventure.html)
- [ITALIAN BRAINROT BIKE RUSH](https://quizverses.github.io/italian-brainrot-bike-rush.html)
- [PAWS PALS DINER](https://studyquests.github.io/paws-pals-diner.html)
- [COLOR NONOGRAM PUZZLE 2](https://quizverses.pages.dev/color-nonogram-puzzle-2.html)
- [MONA LISA FASHION EXPERIMENTS](https://learnquester.github.io/mona-lisa-fashion-experiments.html)
- [MERGE WAR](https://quizverses.github.io/merge-war.html)
- [RENT OUT LANDLORD TYCOON](https://studyplaying.github.io/rent-out-landlord-tycoon.html)
- [2048 MERGE CIRCLE](https://quizverses-9d2f2.web.app/2048-merge-circle.html)
- [BLOCKS AND THATS IT](https://studyplayings.web.app/blocks-and-thats-it.html)
- [FUNNY FEVER HOSPITAL](https://quizverses.pages.dev/funny-fever-hospital.html)
- [CATEGORY RPG80](https://studyplaying.github.io/category-rpg80.html)
- [OBBY PRISON CRAFT ESCAPE](https://studyplayings.web.app/obby-prison-craft-escape.html)
- [POXEL IO](https://studyquests.github.io/poxel-io.html)
- [NG FLOW LINES](https://studyquests.github.io/ng-flow-lines.html)
- [SOLITAIRE QUEST](https://studyquests.github.io/solitaire-quest.html)
- [JAIL PRISON VAN POLICE GAME](https://studyplaying.github.io/jail-prison-van-police-game.html)
- [POWERFUL PUNCH](https://studyplayings.web.app/powerful-punch.html)
- [BUS JAM](https://quizverses-9d2f2.web.app/bus-jam.html)
- [ARROW HIT](https://quizverses.github.io/arrow-hit.html)
- [LATUTU HOLIDAY GIFT HUNT](https://studyplayings.pages.dev/latutu-holiday-gift-hunt.html)
- [CATEGORY RPG](https://studyplaying.github.io/category-rpg.html)
- [ROBLOX CHRISTMAS DRESSUP](https://quizverses-9d2f2.web.app/roblox-christmas-dressup.html)
- [CATEGORY PUZZLE 2](https://studyplayings.web.app/category-puzzle-2.html)
- [BOMB HEAD HOT POTATO](https://studyquests.github.io/bomb-head-hot-potato.html)
- [FLOWBALL](https://studyplaying.github.io/flowball.html)
- [REACH 2048](https://quizverses.github.io/reach-2048.html)
- [EPIC RACING DESCENT ON CARS](https://quizverses.github.io/epic-racing-descent-on-cars.html)
- [MR BEAN JUMP](https://studyquesthub.web.app/mr-bean-jump.html)
- [MERGE SQUARES](https://studyplaying.github.io/merge-squares.html)
- [CATEGORY PUZZLE 2](https://thelearnquester.web.app/category-puzzle-2.html)
- [ROOFTOP CHALLENGE](https://learnquester.github.io/rooftop-challenge.html)
- [ITALIAN BRAINROT QUIZ](https://studyquests.github.io/italian-brainrot-quiz.html)
