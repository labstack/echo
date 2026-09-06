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
- [CATEGORY DRAWING34](https://themindplay.github.io/category-drawing34.html)
- [CATEGORY PUZZLE 2](https://studyplaying.github.io/category-puzzle-2.html)
- [3 TILES](https://quizverses.pages.dev/3-tiles.html)
- [PIRATE PARADISE](https://studyquests.github.io/pirate-paradise.html)
- [CUPIDS STORY LOVE ARCHER BOW](https://quizverses-9d2f2.web.app/cupids-story-love-archer-bow.html)
- [3D BASKETBALLIO DUNK SPORT](https://quizverses-9d2f2.web.app/3d-basketballio-dunk-sport.html)
- [BRAINROT WORLD HOLEIO](https://studyquests.github.io/brainrot-world-holeio.html)
- [MAHJONG FOUR RIVERS](https://studyquests.github.io/mahjong-four-rivers.html)
- [ARCHER DUNGEON HERO](https://quizverses-9d2f2.web.app/archer-dungeon-hero.html)
- [TANKS RACE FOR SURVIVAL](https://studyquests.github.io/tanks-race-for-survival.html)
- [GOOD TO DRIVE](https://quizverses-9d2f2.web.app/good-to-drive.html)
- [NEW YEARS MIRACLES CONNECT THE BALLS](https://studyquests.github.io/new-years-miracles-connect-the-balls.html)
- [CATEGORY PROXY](https://studyquests.github.io/category-proxy.html)
- [CATEGORY TURN BASED30](https://studyquests.github.io/category-turn-based30.html)
- [LOST PUPPY RESCUE AND CARE](https://studyquests.github.io/lost-puppy-rescue-and-care.html)
- [DOGGI](https://studyquests.github.io/doggi.html)
- [MONSTER TRUCK CRUSH](https://quizverses-9d2f2.web.app/monster-truck-crush.html)
- [CATEGORY PUZZLE 2](https://studyquests.github.io/category-puzzle-2.html)
- [CATEGORY SECURLY BYPASS](https://studyquests.github.io/category-securly-bypass.html)
- [CATEGORY SOLDIER](https://studyquests.github.io/category-soldier.html)
- [KIRKA IO](https://studyquests.github.io/kirka-io.html)
- [ROOM SORT](https://thequizzone.pages.dev/room-sort.html)
- [CATEGORY 2D1 070](https://thequizzone.pages.dev/category-2d1-070.html)
- [CONQUERIO](https://thequizzone.pages.dev/conquerio.html)
- [CATEGORY POOL](https://thelearnquesters.pages.dev/category-pool.html)
- [THE COUNTERFEIT BANK](https://thelearnquesters.pages.dev/the-counterfeit-bank.html)
- [SHARK CHOMP CHASE](https://themindzone.pages.dev/shark-chomp-chase.html)
- [RACING BALL ADVENTURE](https://quizverses-9d2f2.web.app/racing-ball-adventure.html)
- [SWORDEDIO SPIN AND RUB](https://studyquests.github.io/swordedio-spin-and-rub.html)
- [HIT BALL](https://studyquests.github.io/hit-ball.html)
- [RAGDOLL ARENA 2 PLAYER](https://quizverses-9d2f2.web.app/ragdoll-arena-2-player.html)
- [POPTROPICA](https://theskillquest.pages.dev/poptropica.html)
- [VEGAMIX MATCH 3 VILLAGE](https://studyquests.github.io/vegamix-match-3-village.html)
- [CATEGORY YOUTUBE](https://studyquests.github.io/category-youtube.html)
- [CATEGORY MAHJONG 2](https://iskillquest.pages.dev/category-mahjong-2.html)
- [FOOTBALL HEADS 2026](https://thequizzone.pages.dev/football-heads-2026.html)
- [MOUNTAIN BUS DRIVER](https://thequizzone.pages.dev/mountain-bus-driver.html)
- [DINO SIMULATOR CITY ATTACK](https://thequizzone.pages.dev/dino-simulator-city-attack.html)
- [DEADFLIP FRENZY](https://quizverses-9d2f2.web.app/deadflip-frenzy.html)
- [COIN MERGE](https://themindzone.pages.dev/coin-merge.html)
- [ANIMAL BLOCKS](https://themindzone.pages.dev/animal-blocks.html)
- [CATEGORY MAHJONG GAMES](https://quizverses.pages.dev/category-mahjong-games.html)
- [STACK TOWER PRO](https://studyquests.github.io/stack-tower-pro.html)
- [ROAD TO 7](https://quizverses-9d2f2.web.app/road-to-7.html)
- [MERGE TOWN](https://quizverses-9d2f2.web.app/merge-town.html)
- [CATEGORY MATCH 3](https://quizverses-9d2f2.web.app/category-match-3.html)
- [SHELF SHIFT MATCH](https://quizverses.pages.dev/shelf-shift-match.html)
- [CATEGORY SECURLY](https://quizverses-9d2f2.web.app/category-securly.html)
- [CRAZY BAR BRAWL](https://studyquests.github.io/crazy-bar-brawl.html)
- [ICONIC HALLOWEEN COSTUMES](https://thequizzone.pages.dev/iconic-halloween-costumes.html)
- [MERGE SESAME](https://studyquests.github.io/merge-sesame.html)
- [MATH KING MATH SKILL GAME](https://thelearnquesters.pages.dev/math-king-math-skill-game.html)
- [HIDE AND SEEK HORROR ESCAPE](https://themindzone.pages.dev/hide-and-seek-horror-escape.html)
- [CATEGORY SOCCER](https://thequizzone.pages.dev/category-soccer.html)
- [CATEGORY MINECRAFT81](https://quizverses.pages.dev/category-minecraft81.html)
- [CATEGORY SHOOTER](https://quizverses.pages.dev/category-shooter.html)
- [DRAW TO SMASH ZOMBIE](https://quizverses.pages.dev/draw-to-smash-zombie.html)
- [CATEGORY ADVENTURE 2](https://iskillquest.pages.dev/category-adventure-2.html)
- [PAPER DOLL DIARY DRESS UP DIY](https://thelearnquesters.pages.dev/paper-doll-diary-dress-up-diy.html)
- [KAWAII CLAW MERGE](https://quizverses.pages.dev/kawaii-claw-merge.html)
- [MYSTICAL BLADE 3D](https://theskillquest.pages.dev/mystical-blade-3d.html)
- [BANK ROBBERY 3](https://studyquests.github.io/bank-robbery-3.html)
- [FOOTBALL HEADS 2025](https://studyquests.github.io/football-heads-2025.html)
- [ICE CUBE](https://theskillquest.pages.dev/ice-cube.html)
- [STICKMAN RAGDOLL PLAYGROUND](https://quizverses-9d2f2.web.app/stickman-ragdoll-playground.html)
- [CATEGORY HORROR 2](https://theskillquest.pages.dev/category-horror-2.html)
- [CATEGORY STICKMAN](https://studyquests.github.io/category-stickman.html)
- [SMASH DEFENSE](https://themindzone.pages.dev/smash-defense.html)
- [CATEGORY 3D1 371](https://iskillquest.pages.dev/category-3d1-371.html)
- [INDEX23](https://iskillquest.pages.dev/index23.html)
- [HEXA SORT WINTER EDITION](https://quizverses.pages.dev/hexa-sort-winter-edition.html)
- [SOLVE THE CUBE WOODEN BLOCKS 2D](https://theskillquest.pages.dev/solve-the-cube-wooden-blocks-2d.html)
- [MAGIC BUBBLES](https://quizverses.pages.dev/magic-bubbles.html)
- [FOOTBALL PENALTY 2026](https://studyquests.github.io/football-penalty-2026.html)
- [ULTIMATE PLANTS TD](https://thelearnquesters.pages.dev/ultimate-plants-td.html)
- [CATEGORY MATCH 3117](https://theskillquest.pages.dev/category-match-3117.html)
- [PLANET HOPPER](https://thelearnquesters.pages.dev/planet-hopper.html)
- [MISSION SANTA DELIVER THE GIFTS](https://studyquests.github.io/mission-santa-deliver-the-gifts.html)
- [TWO SUPRA DRIFTERS](https://quizverses.pages.dev/two-supra-drifters.html)
- [CATEGORY SHOOTER 2](https://quizverses.pages.dev/category-shooter-2.html)
- [GEOMETRY VIBES X BALL](https://studyquests.github.io/geometry-vibes-x-ball.html)
- [PICK BRAINROT 3D BATTLE](https://thequizzone.pages.dev/pick-brainrot-3d-battle.html)
- [GARDEN BLOCK PUZZLE](https://quizverses.pages.dev/garden-block-puzzle.html)
- [CATEGORY WAR](https://studyquests.github.io/category-war.html)
- [CATEGORY WEB PROXY](https://quizverses-9d2f2.web.app/category-web-proxy.html)
- [CAPYBARA XMAS MERGE](https://quizverses.pages.dev/capybara-xmas-merge.html)
- [PONGOAL](https://studyquests.github.io/pongoal.html)
- [MATCH ARENA](https://quizverses.pages.dev/match-arena.html)
- [MERGE TIKTOK GRAVITY KNIFE](https://themindzone.pages.dev/merge-tiktok-gravity-knife.html)
- [ISOMETRIC ESCAPE 2](https://studyquests.github.io/isometric-escape-2.html)
- [SEEK FIND](https://quizverses.pages.dev/seek-find.html)
- [SNAKE 2048IO](https://quizverses-9d2f2.web.app/snake-2048io.html)
- [INK SHOP DRESS TATTOO](https://thequizzone.pages.dev/ink-shop-dress-tattoo.html)
- [CATEGORY SPACE](https://quizverses-9d2f2.web.app/category-space.html)
- [CATEGORY CRAFTING45](https://themindzone.pages.dev/category-crafting45.html)
- [INDEX21](https://iskillquest.pages.dev/index21.html)
- [CATEGORY SOLDIER11](https://quizverses.pages.dev/category-soldier11.html)
- [UNICORN PRINCESS DRESS UP](https://quizverses-9d2f2.web.app/unicorn-princess-dress-up.html)
- [ANIMALS MERGE](https://thelearnquesters.pages.dev/animals-merge.html)
- [INDEX36](https://thequizzone.pages.dev/index36.html)
- [DOP PUZZLE ERASE MASTER](https://studyquests.github.io/dop-puzzle-erase-master.html)
- [CATEGORY SHOOTER 2](https://quizverses-9d2f2.web.app/category-shooter-2.html)
- [CATEGORY TOWER DEFENSE](https://quizverses-9d2f2.web.app/category-tower-defense.html)
- [CATEGORY COLLECT565](https://thequizzone.pages.dev/category-collect565.html)
- [MAFIA ROULETTE](https://studyquests.github.io/mafia-roulette.html)
- [INDEX23](https://thequizzone.pages.dev/index23.html)
- [SAVE HER TOUR](https://themindzone.pages.dev/save-her-tour.html)
- [TRAFFIC LIGHT SIMULATOR 3D](https://quizverses-9d2f2.web.app/traffic-light-simulator-3d.html)
- [PLANTS WARFARE](https://thelearnquesters.pages.dev/plants-warfare.html)
- [CATEGORY TOWER DEFENSE118](https://thequizzone.pages.dev/category-tower-defense118.html)
- [BOXING FIGHTER](https://studyquests.github.io/boxing-fighter.html)
- [ARROW HIT](https://studyquests.github.io/arrow-hit.html)
- [BUBBLE AROUND](https://quizverses-9d2f2.web.app/bubble-around.html)
- [SNIPER 3D ZOMBIE](https://studyquests.github.io/sniper-3d-zombie.html)
- [PLANET EVOLUTION IDLE CLICKER](https://studyquests.github.io/planet-evolution-idle-clicker.html)
- [CATEGORY UNBLOCKED WEBSITES](https://quizverses-9d2f2.web.app/category-unblocked-websites.html)
- [NUBIK IN THE MONSTER WORLD](https://thequizzone.pages.dev/nubik-in-the-monster-world.html)
- [STRIKE IT](https://quizverses-9d2f2.web.app/strike-it.html)
- [DOWNHILL CAR RIDE CRASH TEST](https://thequizzone.pages.dev/downhill-car-ride-crash-test.html)
- [CATEGORY SNAKE40](https://quizverses.pages.dev/category-snake40.html)
- [STICK FIGHT THE CHAOS](https://thequizzone.pages.dev/stick-fight-the-chaos.html)
- [PRACTICE ON ME](https://thequizzone.pages.dev/practice-on-me.html)
- [ASSASSIN COMMANDO CAR DRIVING](https://theskillquest.pages.dev/assassin-commando-car-driving.html)
- [AROUND ELBRUS](https://thelearnquesters.pages.dev/around-elbrus.html)
- [CATEGORY TANK58](https://studyquests.github.io/category-tank58.html)
- [CATEGORY WAR137](https://quizverses-9d2f2.web.app/category-war137.html)
- [2048 MERGE CIRCLE](https://quizverses-9d2f2.web.app/2048-merge-circle.html)
- [CARS DERBY ARENA](https://thequizzone.pages.dev/cars-derby-arena.html)
- [CUPIDS STORY LOVE ARCHER BOW](https://quizverses.pages.dev/cupids-story-love-archer-bow.html)
- [FIGHT TO THE END](https://quizverses.pages.dev/fight-to-the-end.html)
