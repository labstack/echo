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
- [PIN DETECTIVE](https://studyplaying.github.io/pin-detective.html)
- [CATEGORY TITANIUM NETWORK](https://quizverses-9d2f2.web.app/category-titanium-network.html)
- [STELLAR FUSION](https://studyquests.pages.dev/stellar-fusion.html)
- [CATEGORY ARENA254](https://themindzone.pages.dev/category-arena254.html)
- [PIMPLE SQUEEZE](https://studyplaying.github.io/pimple-squeeze.html)
- [PACKING LINE](https://quizverses-9d2f2.web.app/packing-line.html)
- [CATEGORY IDLE445](https://studyplaying.github.io/category-idle445.html)
- [HOLE EAT GROW ATTACK](https://studyplaying.github.io/hole-eat-grow-attack.html)
- [K WEDDING DREAM](https://quizverses.github.io/k-wedding-dream.html)
- [CATEGORY CONTROLLER59](https://thelearnquester.web.app/category-controller59.html)
- [ICE CREAM INC](https://studyplayings.web.app/ice-cream-inc.html)
- [RAGDOLL JUMP](https://studyplayings.web.app/ragdoll-jump.html)
- [CATEGORY RUNNING](https://studyquests.github.io/category-running.html)
- [CATEGORY UNBLOCKED](https://studyplayings.pages.dev/category-unblocked.html)
- [BLOCK MASTER SUPER PUZZLE](https://studyplayings.pages.dev/block-master-super-puzzle.html)
- [MINI GOLF BATTLE](https://studyplayings.web.app/mini-golf-battle.html)
- [CATEGORY SOCCER](https://studyplayings.web.app/category-soccer.html)
- [NOOB FUN FISHING](https://studyplayings.web.app/noob-fun-fishing.html)
- [FPS TOY REALISM](https://studyplayings.web.app/fps-toy-realism.html)
- [CATEGORY FASHION105](https://studyquests.github.io/category-fashion105.html)
- [CATEGORY MAKEUP](https://studyquests.github.io/category-makeup.html)
- [FUN IQ PUZZLE](https://studyplayings.web.app/fun-iq-puzzle.html)
- [CATEGORY RPG GAMES](https://studyplayings.pages.dev/category-rpg-games.html)
- [AVATAR MASTER FIX UP FACE](https://studyplayings.web.app/avatar-master-fix-up-face.html)
- [KITTY MATCH 3 PUZZLE GAME](https://studyquests.github.io/kitty-match-3-puzzle-game.html)
- [GROW WARSIO](https://quizverses-9d2f2.web.app/grow-warsio.html)
- [CATEGORY UNBLOCKERS](https://quizverses.pages.dev/category-unblockers.html)
- [CATEGORY SPORTS](https://quizverses-9d2f2.web.app/category-sports.html)
- [CATEGORY SIDE SCROLLING184](https://studyplayings.pages.dev/category-side-scrolling184.html)
- [HOME BLOCK STORY](https://studyplayings.pages.dev/home-block-story.html)
- [ISLAND PUZZLE BUILD SOLVE](https://quizverses-9d2f2.web.app/island-puzzle-build-solve.html)
- [GRADUATION MAKEUP TRENDS](https://quizverses-9d2f2.web.app/graduation-makeup-trends.html)
- [EGG DASH](https://studyplayings.web.app/egg-dash.html)
- [MERGE SMITH](https://studyquests.github.io/merge-smith.html)
- [CATEGORY ANIMAL216](https://quizverses-9d2f2.web.app/category-animal216.html)
- [CATEGORY MAHJONG](https://studyplayings.pages.dev/category-mahjong.html)
- [BLACKRIVER MYSTERY HIDDEN OBJECTS](https://quizverses-9d2f2.web.app/blackriver-mystery-hidden-objects.html)
- [CATEGORY MERGE224](https://quizverses.pages.dev/category-merge224.html)
- [JIGSAW CARDS DAILY PUZZLES](https://studyplayings.web.app/jigsaw-cards-daily-puzzles.html)
- [QUEENS ROYAL SUDOKU PUZZLE](https://studyplayings.web.app/queens-royal-sudoku-puzzle.html)
- [CATEGORY MMO25](https://studyplayings.pages.dev/category-mmo25.html)
- [VEGAMIX2 WILD WEST](https://studyquesthub.web.app/vegamix2-wild-west.html)
- [CATEGORY MONSTER207](https://quizverses.pages.dev/category-monster207.html)
- [GROW CASTLE DEFENCE](https://quizverses-9d2f2.web.app/grow-castle-defence.html)
- [MAFIA ROULETTE](https://studyplayings.web.app/mafia-roulette.html)
- [ZOMBIE HIGHWAY RAMPAGE](https://quizverses.pages.dev/zombie-highway-rampage.html)
- [BOLTS UNSCREW IT](https://studyplayings.web.app/bolts-unscrew-it.html)
- [BUBBLE SHOOTER VINTAGE](https://quizverses-9d2f2.web.app/bubble-shooter-vintage.html)
- [GUN CRAFT RUN WEAPON FIRE](https://studyplayings.pages.dev/gun-craft-run-weapon-fire.html)
- [DIGIT SHOOTER](https://quizverses.github.io/digit-shooter.html)
- [MEOW MARKET](https://quizverses.github.io/meow-market.html)
- [CATEGORY STICKMAN 2](https://studyplayings.pages.dev/category-stickman-2.html)
- [POWER PUZZLE](https://studyquesthub.web.app/power-puzzle.html)
- [CATEGORY SHOOTER 2](https://studyplayings.pages.dev/category-shooter-2.html)
- [CATEGORY MANAGEMENT](https://quizverses.pages.dev/category-management.html)
- [BLOCK UP](https://studyplayings.web.app/block-up.html)
- [CATEGORY TRAFFIC34](https://studyquesthub.web.app/category-traffic34.html)
- [CATEGORY RPG](https://studyquests.github.io/category-rpg.html)
- [CATEGORY PUZZLE 4](https://studyplayings.pages.dev/category-puzzle-4.html)
- [DART HERO](https://studyplayings.web.app/dart-hero.html)
- [CATEGORY CAN T STOP PLAYING212](https://quizverses.pages.dev/category-can-t-stop-playing212.html)
- [WORLD WARS TANKS](https://studyquests.github.io/world-wars-tanks.html)
- [CATEGORY PROXY](https://studyquests.github.io/category-proxy.html)
- [SNIPING ALIENS](https://studyplayings.pages.dev/sniping-aliens.html)
- [KNOCK AND RUN 100 DOORS ESCAPE](https://studyquests.github.io/knock-and-run-100-doors-escape.html)
- [TAIL GUN CHARLIE](https://studyplayings.web.app/tail-gun-charlie.html)
- [CATEGORY TITANIUM NETWORK](https://studyquests.github.io/category-titanium-network.html)
- [CORNHOLE LEAGUE BOARD GAMES](https://studyplayings.web.app/cornhole-league-board-games.html)
- [CAR ESCAPE](https://studyplayings.web.app/car-escape.html)
- [CATEGORY SHOOTER](https://studyplayings.pages.dev/category-shooter.html)
- [MONSTER TRUCK CRUSH](https://studyplayings.pages.dev/monster-truck-crush.html)
- [ZOMBIE DEFENSE WAR](https://studyplayings.web.app/zombie-defense-war.html)
- [CATEGORY GOGUARDIAN](https://thelearnquester.web.app/category-goguardian.html)
- [CATEGORY COLOR](https://studyquesthub.web.app/category-color.html)
- [CATEGORY TOOLS](https://studyplayings.pages.dev/category-tools.html)
- [CATEGORY COOKING](https://thelearnquester.web.app/category-cooking.html)
- [CATEGORY LISTS](https://thelearnquester.web.app/category-lists.html)
- [CATEGORY ROGUELIKE38](https://studyplayings.pages.dev/category-roguelike38.html)
- [TOKA BOKA HOME CLEAN UP DESIGN](https://studyquests.github.io/toka-boka-home-clean-up-design.html)
- [SUPER KID ADVENTURE](https://studyplayings.web.app/super-kid-adventure.html)
- [CANNONS BLAST 3D](https://studyquesthub.web.app/cannons-blast-3d.html)
- [PERFECT CAKE MAKER](https://studyquests.github.io/perfect-cake-maker.html)
- [CATEGORY SIDE SCROLLING184](https://studyplayings.web.app/category-side-scrolling184.html)
- [CATEGORY RELAXING221](https://thelearnquester.web.app/category-relaxing221.html)
- [NUMBER MERGE 10](https://learnquester.github.io/number-merge-10.html)
- [MR LONG LEGS](https://studyquesthub.web.app/mr-long-legs.html)
- [REACH 2048](https://studyplayings.pages.dev/reach-2048.html)
- [CATEGORY PROXY](https://studyplayings.pages.dev/category-proxy.html)
- [PHONE CASE DIY 5](https://studyplayings.pages.dev/phone-case-diy-5.html)
- [RUN 3D](https://studyquests.github.io/run-3d.html)
- [SCREW COLOR SORTING MASTER](https://quizverses-9d2f2.web.app/screw-color-sorting-master.html)
- [SORTING FROGS](https://studyplayings.web.app/sorting-frogs.html)
- [CANDY JEWELS](https://quizverses-9d2f2.web.app/candy-jewels.html)
- [SECRETS OF CHARMLAND](https://quizverses-9d2f2.web.app/secrets-of-charmland.html)
- [JAILBREAK ASSAULT](https://learnquester.github.io/jailbreak-assault.html)
- [REVOXEL 3D VOXEL RPG SHOOTER](https://studyplayings.pages.dev/revoxel-3d-voxel-rpg-shooter.html)
- [CATEGORY FOOD](https://studyquests.github.io/category-food.html)
- [BATTLE ISLAND 2](https://studyquests.github.io/battle-island-2.html)
- [CAPYBARA COIN MASTER](https://quizverses-9d2f2.web.app/capybara-coin-master.html)
- [PLANTS VS ZOMBIES WAR](https://studyquesthub.web.app/plants-vs-zombies-war.html)
- [AIDAN IN DANGER](https://quizverses-9d2f2.web.app/aidan-in-danger.html)
- [XMAS HEXA SORT](https://studyquesthub.web.app/xmas-hexa-sort.html)
- [POGO MASTERS](https://quizverses-9d2f2.web.app/pogo-masters.html)
- [CATEGORY LOGIC](https://thequizzone.pages.dev/category-logic.html)
- [NUTS AND BOLTS SCREW PUZZLE](https://thelearnquester.web.app/nuts-and-bolts-screw-puzzle.html)
- [OVERPROTECTIVE BOYFRIEND](https://learnquester.pages.dev/overprotective-boyfriend.html)
- [MONA LISA FASHION EXPERIMENTS](https://learnquester.github.io/mona-lisa-fashion-experiments.html)
- [CRASH THE ROBOT](https://quizverses-9d2f2.web.app/crash-the-robot.html)
- [BACTERIA LIFE DEATH](https://thelearnquester.web.app/bacteria-life-death.html)
- [ONU LIVE](https://learnquester.github.io/onu-live.html)
- [THE GRENCH COUPLE HOLIDAY DRESS UP](https://thequizzone.pages.dev/the-grench-couple-holiday-dress-up.html)
- [2 PLAYER ONLINE CHESS](https://learnquester.pages.dev/2-player-online-chess.html)
- [PRACTICE ON ME](https://studyquests.github.io/practice-on-me.html)
- [WORLDGUESSR](https://studyplaying.github.io/worldguessr.html)
- [CATEGORY PHYSICS371](https://studyquests.pages.dev/category-physics371.html)
- [CATEGORY UNBLOCKED](https://thequizzone.pages.dev/category-unblocked.html)
- [SNEAKER ART](https://studyplaying.github.io/sneaker-art.html)
- [CELEBRITY SPRING FASHION TRENDS](https://thequizzone.pages.dev/celebrity-spring-fashion-trends.html)
- [CATEGORY SPACE57](https://quizverses-9d2f2.web.app/category-space57.html)
- [CATEGORY MANAGEMENT210](https://thequizzone.pages.dev/category-management210.html)
- [BASKET SWAP](https://studyquests.github.io/basket-swap.html)
- [MALATANG MASTER STACK RUN 3D](https://thequizzone.pages.dev/malatang-master-stack-run-3d.html)
- [CATEGORY UNBLOCKED](https://quizverses-9d2f2.web.app/category-unblocked.html)
- [GIANT SUSHI MERGE MASTER GAME](https://quizverses.github.io/giant-sushi-merge-master-game.html)
- [GETTING OVER IT](https://quizverses-9d2f2.web.app/getting-over-it.html)
- [RESTAURANT VIP MASTERCHEF](https://learnquester.github.io/restaurant-vip-masterchef.html)
- [THE ROAD HOME GRANNY ESCAPE](https://thelearnquesters.pages.dev/the-road-home-granny-escape.html)
- [OIL DIGGING](https://thelearnquesters.pages.dev/oil-digging.html)
- [CATEGORY INTERSTELLARUNBLOCKER](https://thequizzone.pages.dev/category-interstellarunblocker.html)
- [MY CASTLE MERGE STORY](https://thequizzone.pages.dev/my-castle-merge-story.html)
