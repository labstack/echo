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
- [BRAT GIRL SUMMER](https://studyplaying.github.io/brat-girl-summer.html)
- [STREET RACING MOTO DRIFT](https://thelearnquesters.pages.dev/street-racing-moto-drift.html)
- [TYPE SPRINT](https://studyquests.github.io/type-sprint.html)
- [SQUARE PUNKI LONG HAND](https://quizverses.pages.dev/square-punki-long-hand.html)
- [SOLITAIRE SUMMER KLONDIKE](https://quizverses-9d2f2.web.app/solitaire-summer-klondike.html)
- [ADDICTION SOLITAIRE](https://quizverses-9d2f2.web.app/addiction-solitaire.html)
- [CATEGORY FPS175](https://quizverses-9d2f2.web.app/category-fps175.html)
- [WORDMEISTER HD](https://studyquesthub.web.app/wordmeister-hd.html)
- [PET SALON](https://quizverses-9d2f2.web.app/pet-salon.html)
- [VENETIAN LOVE AFFAIR](https://studyquests.github.io/venetian-love-affair.html)
- [URUS CITY DRIVER](https://quizverses.pages.dev/urus-city-driver.html)
- [CATEGORY GROW99](https://quizverses.pages.dev/category-grow99.html)
- [FRUITSLAND ESCAPE FROM THE AMUSEMENT PARK](https://quizverses.pages.dev/fruitsland-escape-from-the-amusement-park.html)
- [MATH DUCK](https://quizverses.pages.dev/math-duck.html)
- [CATEGORY COLLECT565](https://quizverses-9d2f2.web.app/category-collect565.html)
- [ART SALON](https://quizverses.pages.dev/art-salon.html)
- [OBBY ESCAPE BARRYS JAIL PARKOUR](https://studyquests.github.io/obby-escape-barrys-jail-parkour.html)
- [FLOW BLOCK](https://quizverses.pages.dev/flow-block.html)
- [HARVESTING VEGGIES](https://quizverses.github.io/harvesting-veggies.html)
- [THE OFFICE ESCAPE](https://quizverses.pages.dev/the-office-escape.html)
- [CATEGORY CASUAL 6](https://quizverses-9d2f2.web.app/category-casual-6.html)
- [CATEGORY TITANIUMNETWORK](https://quizverses.pages.dev/category-titaniumnetwork.html)
- [LIGHTS OUT](https://quizverses.pages.dev/lights-out.html)
- [MR LONG HAND](https://quizverses-9d2f2.web.app/mr-long-hand.html)
- [LAMPHEAD](https://studyquests.github.io/lamphead.html)
- [TAP ARROW AWAY](https://quizverses.pages.dev/tap-arrow-away.html)
- [ROYAL GARDEN MATCH](https://quizverses-9d2f2.web.app/royal-garden-match.html)
- [COUNT MASTERS SUPERHERO](https://quizverses.pages.dev/count-masters-superhero.html)
- [MONSTER TRUCK CRUSH](https://quizverses-9d2f2.web.app/monster-truck-crush.html)
- [MARBLE BLAST](https://studyquests.github.io/marble-blast.html)
- [MINI GRAND THEFT CITY](https://quizverses.pages.dev/mini-grand-theft-city.html)
- [CATEGORY ESCAPE](https://quizverses.pages.dev/category-escape.html)
- [ITALIAN ANIMALS CREATE YOUR OWN BRAINROT](https://studyquests.github.io/italian-animals-create-your-own-brainrot.html)
- [HOOP RIVALS](https://quizverses.pages.dev/hoop-rivals.html)
- [COOL CARS RACING AT ALTITUDE](https://quizverses.pages.dev/cool-cars-racing-at-altitude.html)
- [SERIOUS HEAD 2](https://quizverses.github.io/serious-head-2.html)
- [CATEGORY ESCAPE](https://quizverses-9d2f2.web.app/category-escape.html)
- [CATEGORY CASUAL 4](https://thequizzone.pages.dev/category-casual-4.html)
- [ARCHER DUNGEON HERO](https://quizverses-9d2f2.web.app/archer-dungeon-hero.html)
- [MAGIC BUBBLES](https://quizverses.pages.dev/magic-bubbles.html)
- [KINGDOM OF PIXELS](https://studyquesthub.web.app/kingdom-of-pixels.html)
- [CATEGORY SECURLY](https://quizverses.pages.dev/category-securly.html)
- [MINI GOLF BATTLE](https://themindzone.pages.dev/mini-golf-battle.html)
- [GIN RUMMY](https://thequizzone.pages.dev/gin-rummy.html)
- [FARM MATCH SEASONS 3](https://thelearnquesters.pages.dev/farm-match-seasons-3.html)
- [LITTLE COMMANDER RED VS BLUE](https://thelearnquesters.pages.dev/little-commander-red-vs-blue.html)
- [CATEGORY ADVENTURE 4](https://thequizzone.pages.dev/category-adventure-4.html)
- [CATEGORY THINKY](https://thelearnquesters.pages.dev/category-thinky.html)
- [SUDOKU CLASSIC DAILY BRAIN PUZZLE](https://thequizzone.pages.dev/sudoku-classic-daily-brain-puzzle.html)
- [CATEGORY MMO25](https://quizverses.pages.dev/category-mmo25.html)
- [PURRFECT PUZZLE](https://studyquesthub.web.app/purrfect-puzzle.html)
- [CATEGORY TOWER DEFENSE](https://thelearnquesters.pages.dev/category-tower-defense.html)
- [LAQUEUS ESCAPE CHAPTER III](https://thequizzone.pages.dev/laqueus-escape-chapter-iii.html)
- [CATEGORY ALIEN34](https://quizverses.pages.dev/category-alien34.html)
- [WEDNESDAY LIGHT ACADEMIA](https://thelearnquesters.pages.dev/wednesday-light-academia.html)
- [METAL GUNS FURY](https://quizverses.pages.dev/metal-guns-fury.html)
- [MAHJONG ZEN GARDEN](https://thelearnquesters.pages.dev/mahjong-zen-garden.html)
- [IDLE LEGEND](https://quizverses.pages.dev/idle-legend.html)
- [LAST TO LEAVE CIRCLE OBBY](https://studyquesthub.web.app/last-to-leave-circle-obby.html)
- [BOMBER BATTLE ARENA](https://quizverses.pages.dev/bomber-battle-arena.html)
- [HEXA TILE MASTER](https://thelearnquesters.pages.dev/hexa-tile-master.html)
- [CATEGORY HORDE SURVIVAL67](https://quizverses-9d2f2.web.app/category-horde-survival67.html)
- [COZY KITCHEN MERGE](https://thelearnquesters.pages.dev/cozy-kitchen-merge.html)
- [LADY POOL](https://thelearnquesters.pages.dev/lady-pool.html)
- [KITTY SQUAD WINTER DRESS UP](https://thelearnquesters.pages.dev/kitty-squad-winter-dress-up.html)
- [SNEAKY FRIENDS](https://thelearnquesters.pages.dev/sneaky-friends.html)
- [LEAP AND AVOID 2](https://thelearnquesters.pages.dev/leap-and-avoid-2.html)
- [GLOVES GROW RUSH](https://studyquests.github.io/gloves-grow-rush.html)
- [HEXA PUZZLE](https://thequizzone.pages.dev/hexa-puzzle.html)
- [LIPSTICK COLLECTOR RUN](https://quizverses.github.io/lipstick-collector-run.html)
- [GOLF MINI](https://quizverses.pages.dev/golf-mini.html)
- [MY LITTLE CAR WASH](https://quizverses.github.io/my-little-car-wash.html)
- [PURRFECT SCOOPS](https://thelearnquesters.pages.dev/purrfect-scoops.html)
- [OCTONAUTS BUBBLES](https://thelearnquesters.pages.dev/octonauts-bubbles.html)
- [SCOOTER TOUCHGRIND TRICKS 3D](https://thelearnquesters.pages.dev/scooter-touchgrind-tricks-3d.html)
- [CATEGORY AVOID295](https://quizverses.pages.dev/category-avoid295.html)
- [ASMR BEAUTY SUPERSTAR](https://quizverses-9d2f2.web.app/asmr-beauty-superstar.html)
- [CATEGORY WAR137](https://quizverses.pages.dev/category-war137.html)
- [INDEX18](https://studyquesthub.web.app/index18.html)
- [HOUSE ROBBER](https://quizverses.github.io/house-robber.html)
- [COWBOYS DUEL](https://studyquests.github.io/cowboys-duel.html)
- [SCHOOLBOY RUNAWAY ROOM ESCAPE](https://quizverses.pages.dev/schoolboy-runaway-room-escape.html)
- [HERO FIGHT CLASH](https://thelearnquesters.pages.dev/hero-fight-clash.html)
- [HALLOWEEN MAKEUP TRENDS](https://quizverses.github.io/halloween-makeup-trends.html)
- [NINE CARDS OF WINTER](https://thequizzone.pages.dev/nine-cards-of-winter.html)
- [ANGRY CITY SMASHER](https://thelearnquesters.pages.dev/angry-city-smasher.html)
- [MINE FPS SHOOTER NOOB ARENA](https://thelearnquesters.pages.dev/mine-fps-shooter-noob-arena.html)
- [STICKMAN THE FLASH](https://studyquesthub.web.app/stickman-the-flash.html)
- [KEY QUEST](https://thequizzone.pages.dev/key-quest.html)
- [IDLE MERGE CAR AND RACE](https://quizverses.pages.dev/idle-merge-car-and-race.html)
- [YUMMY TRAILS](https://thequizzone.pages.dev/yummy-trails.html)
- [INTERIOR DESIGNER DECOR LIFE](https://thequizzone.pages.dev/interior-designer-decor-life.html)
- [UNLOCK THE BOLTS](https://thequizzone.pages.dev/unlock-the-bolts.html)
- [BATTLE SIMULATOR SANDBOX](https://thelearnquesters.pages.dev/battle-simulator-sandbox.html)
- [TAP AWAY](https://thelearnquesters.pages.dev/tap-away.html)
- [JET FIGHTER AIRPLANE RACING](https://quizverses.github.io/jet-fighter-airplane-racing.html)
- [THROUGH THE WALL](https://quizverses-9d2f2.web.app/through-the-wall.html)
- [CATEGORY JUMPING147](https://studyquesthub.web.app/category-jumping147.html)
- [MIND GAMBIT](https://thelearnquesters.pages.dev/mind-gambit.html)
- [MERMAID PRINCESS AVATER CASTLE](https://thelearnquesters.pages.dev/mermaid-princess-avater-castle.html)
- [TB AVATARIA LIFE GIRL](https://quizverses.pages.dev/tb-avataria-life-girl.html)
- [HEXON RUSH](https://studyquests.github.io/hexon-rush.html)
- [TIE DYE EXPLOSION OF COLOR](https://quizverses.github.io/tie-dye-explosion-of-color.html)
- [CATEGORY STRATEGY](https://quizverses.github.io/category-strategy.html)
- [XMAS PRESENTS MAHJONG](https://thelearnquesters.pages.dev/xmas-presents-mahjong.html)
- [CARGO SKATES](https://thelearnquesters.pages.dev/cargo-skates.html)
- [BOAT MANIA](https://thelearnquesters.pages.dev/boat-mania.html)
- [SORT TILES](https://thelearnquesters.pages.dev/sort-tiles.html)
- [OBBY YARD SALE](https://thequizzone.pages.dev/obby-yard-sale.html)
- [HIGH HEEL DESIGN](https://thequizzone.pages.dev/high-heel-design.html)
- [SCREW SPIN](https://thequizzone.pages.dev/screw-spin.html)
- [CATEGORY BATTLE524](https://quizverses-9d2f2.web.app/category-battle524.html)
- [INDEX18](https://quizverses-9d2f2.web.app/index18.html)
- [SLOPE SNOWBALL](https://thequizzone.pages.dev/slope-snowball.html)
- [SUMMER CONNECT](https://thequizzone.pages.dev/summer-connect.html)
- [WATER SHOOTER](https://studyquests.github.io/water-shooter.html)
- [KITTY MATCH 3 PUZZLE GAME](https://quizverses.github.io/kitty-match-3-puzzle-game.html)
- [BLACK PINK STPATRICKS DAY CONCERT](https://quizverses.github.io/black-pink-stpatricks-day-concert.html)
- [NINJA CLIMB](https://studyquests.github.io/ninja-climb.html)
- [ZOO RESTAURANT](https://thequizzone.pages.dev/zoo-restaurant.html)
- [SNIPER FOR BRAINROT](https://quizverses.pages.dev/sniper-for-brainrot.html)
- [ROLLANCE GOING BALLS](https://thelearnquesters.pages.dev/rollance-going-balls.html)
- [ROBBY BOMBERMAN](https://thequizzone.pages.dev/robby-bomberman.html)
- [HERO TRANSFORM RUN](https://thelearnquesters.pages.dev/hero-transform-run.html)
- [INDEX14](https://studyquesthub.web.app/index14.html)
- [OFFLINE FPS ROYALE](https://quizverses-9d2f2.web.app/offline-fps-royale.html)
- [CATEGORY RUNNING107](https://quizverses.github.io/category-running107.html)
- [MEMEVOIO](https://quizverses.github.io/memevoio.html)
- [CATEGORY SNAKE GAMES](https://quizverses.github.io/category-snake-games.html)
- [ONLINE CAR DESTRUCTION SIMULATOR 3D](https://quizverses.pages.dev/online-car-destruction-simulator-3d.html)
