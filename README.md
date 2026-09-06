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
- [HIDDEN KITTY](https://quizverses.github.io/hidden-kitty.html)
- [INDEX11](https://quizverses-9d2f2.web.app/index11.html)
- [CATEGORY EDUCATIONAL](https://studyquesthub.web.app/category-educational.html)
- [KNEE CASE SIMULATOR](https://studyplaying.github.io/knee-case-simulator.html)
- [BULL RUNNER](https://studyquests.github.io/bull-runner.html)
- [INDEX3](https://thelearnquesters.pages.dev/index3.html)
- [CATEGORY 2D1 060](https://thelearnquesters.pages.dev/category-2d1-060.html)
- [BATTLE ISLAND 2](https://studyquests.github.io/battle-island-2.html)
- [POKER QUEST](https://themindzone.pages.dev/poker-quest.html)
- [SUGAR HEROES](https://studyquests.pages.dev/sugar-heroes.html)
- [CARTOON MOTO STUNT](https://iskillplay.web.app/cartoon-moto-stunt.html)
- [AUTUMN GLAM GALA](https://themindplay.pages.dev/autumn-glam-gala.html)
- [CATEGORY MOBILE2 095](https://themindplaying.web.app/category-mobile2-095.html)
- [CLOCKWORK](https://themindplay.pages.dev/clockwork.html)
- [OFFICE SPIDER SOLITAIRE](https://themindplays.pages.dev/office-spider-solitaire.html)
- [MERGE TIKTOK GRAVITY KNIFE](https://learnquester.pages.dev/merge-tiktok-gravity-knife.html)
- [MINI GAMES RELAX COLLECTION 2](https://themindplaying.web.app/mini-games-relax-collection-2.html)
- [CATEGORY ZOMBIE175](https://thelearnquesters.pages.dev/category-zombie175.html)
- [HIDDEN HORRORS](https://themindplays.pages.dev/hidden-horrors.html)
- [BATTLE ZONE 2D](https://studyquests.pages.dev/battle-zone-2d.html)
- [CRAFTSMAN 3D GANGSTER](https://quizverses.github.io/craftsman-3d-gangster.html)
- [CAR PARK SIMULATOR](https://quizverses.pages.dev/car-park-simulator.html)
- [BEAT THE ZOMBIES](https://learnquester.pages.dev/beat-the-zombies.html)
- [FALLING ART RAGDOLL SIMULATOR](https://studyquests.github.io/falling-art-ragdoll-simulator.html)
- [ROBOT BAND FIND THE DIFFERENCES](https://studyquests.pages.dev/robot-band-find-the-differences.html)
- [INDEX21](https://learnquesters.pages.dev/index21.html)
- [SKYSCRAPER TO THE SKY](https://themindplay.pages.dev/skyscraper-to-the-sky.html)
- [CATEGORY BATTLE524](https://thelearnquesters.pages.dev/category-battle524.html)
- [CATEGORY IDLE445](https://quizverses.github.io/category-idle445.html)
- [INDEX30](https://thelearnquesters.pages.dev/index30.html)
- [ATOMIC MERGE 2048](https://themindplay.pages.dev/atomic-merge-2048.html)
- [HAMSTER COMBO IDLE](https://themindplays.pages.dev/hamster-combo-idle.html)
- [CATEGORY INCREMENTAL388](https://quizverses-9d2f2.web.app/category-incremental388.html)
- [PURSUIT RAMPAGE](https://studyquests.pages.dev/pursuit-rampage.html)
- [GRANDFATHER ROAD CHASE REALISTIC SHOOTER GUNS](https://themindplays.pages.dev/grandfather-road-chase-realistic-shooter-guns.html)
- [CARNAGE BATTLE ARENA](https://themindplays.pages.dev/carnage-battle-arena.html)
- [CATEGORY TITANIUM NETWORK](https://themindplays.pages.dev/category-titanium-network.html)
- [NUGGET MAN SURVIVAL PUZZLE](https://themindplays.pages.dev/nugget-man-survival-puzzle.html)
- [PANDA LU TREEHOUSE](https://themindplay.pages.dev/panda-lu-treehouse.html)
- [BATTLE SIMULATOR SANDBOX](https://thelearnquesters.pages.dev/battle-simulator-sandbox.html)
- [FLICK BASEBALL SUPER HOMERUN](https://learnquester.pages.dev/flick-baseball-super-homerun.html)
- [CATEGORY MERGE224](https://skillplay.github.io/category-merge224.html)
- [CATEGORY OBSTACLE299](https://themindplays.pages.dev/category-obstacle299.html)
- [GOMU GOMAN](https://studyquests.pages.dev/gomu-goman.html)
- [FORCE MASTER 3D](https://themindplays.pages.dev/force-master-3d.html)
- [CATEGORY UNBLOCKED GAMES](https://studyquests.github.io/category-unblocked-games.html)
- [CATEGORY 2048](https://themindplays.pages.dev/category-2048.html)
- [KICK LOSER](https://quizverses-9d2f2.web.app/kick-loser.html)
- [LOVE TILE TRIO](https://themindplays.pages.dev/love-tile-trio.html)
- [SNOW BALL RACING MUTLIPLAYER](https://themindplaying.web.app/snow-ball-racing-mutliplayer.html)
- [CATEGORY ADVENTURE 2](https://quizverses-9d2f2.web.app/category-adventure-2.html)
- [LAVA JUMP](https://themindplays.pages.dev/lava-jump.html)
- [BALL PAINT 3D](https://quizverses.github.io/ball-paint-3d.html)
- [FASHIONISTA AVATAR STUDIO DRESS UP](https://themindplaying.web.app/fashionista-avatar-studio-dress-up.html)
- [QUIZ 10 SECONDS MATH](https://studyquests.github.io/quiz-10-seconds-math.html)
- [CATEGORY BIKE 3](https://themindplays.pages.dev/category-bike-3.html)
- [FORMULA TRAFFIC RACER](https://themindplays.pages.dev/formula-traffic-racer.html)
- [GOBATTLEIO](https://quizverses-9d2f2.web.app/gobattleio.html)
- [CATEGORY CRAFTING45](https://themindplaying.web.app/category-crafting45.html)
- [MONSTERELLA FANTASY MAKEUP](https://themindplay.pages.dev/monsterella-fantasy-makeup.html)
- [LINE ON HOLE](https://themindplays.pages.dev/line-on-hole.html)
- [GOD OF LIGHT](https://thelearnquesters.pages.dev/god-of-light.html)
- [GOON BALL](https://skillplay.github.io/goon-ball.html)
- [COIN STACK UP](https://thelearnquesters.pages.dev/coin-stack-up.html)
- [GEOMETRY STARS](https://thelearnquesters.pages.dev/geometry-stars.html)
- [CATEGORY CRAFTING45](https://thequizzone.pages.dev/category-crafting45.html)
- [CRAZY GOOSE SIMULATOR](https://thelearnquesters.pages.dev/crazy-goose-simulator.html)
- [CATEGORY POOL 3](https://thequizzone.pages.dev/category-pool-3.html)
- [BUSY BEE HIVE](https://quizverses-9d2f2.web.app/busy-bee-hive.html)
- [CATEGORY ARCHERY52](https://learnquester.pages.dev/category-archery52.html)
- [SUPER STOCK STACK](https://themindskillplayplay.pages.dev/super-stock-stack.html)
- [INDEX6](https://themindplays.pages.dev/index6.html)
- [CATEGORY CARE](https://themindskillplayplay.pages.dev/category-care.html)
- [STELLAR GUARDIAN](https://quizverses.pages.dev/stellar-guardian.html)
- [COLOR SORT IMPOSTOR EDITION](https://themindskillplayplay.pages.dev/color-sort-impostor-edition.html)
- [SMOKE TRAIL](https://themindskillplayplay.pages.dev/smoke-trail.html)
- [DINO GAME](https://studyquests.pages.dev/dino-game.html)
- [WATER JUNK WARRIORS](https://thelearnquesters.pages.dev/water-junk-warriors.html)
- [PLANET TAKEOVER](https://studyplaying.github.io/planet-takeover.html)
- [CATEGORY SCIENCE18](https://skillplay.github.io/category-science18.html)
- [COIN EMPIRE](https://iskillplay.web.app/coin-empire.html)
- [FRUIT CAFE MATCH 3](https://quizverses.github.io/fruit-cafe-match-3.html)
- [CATEGORY ZOMBIE175](https://iskillplay.web.app/category-zombie175.html)
- [CATEGORY MINECRAFT 2](https://quizverses.github.io/category-minecraft-2.html)
- [FOOTBALL HEADS 2025](https://studyplaying.github.io/football-heads-2025.html)
- [SKIBIDI SURVIVOR RUSH](https://themindskillplayplay.pages.dev/skibidi-survivor-rush.html)
- [ROLLER COASTER RUSH](https://themindskillplayplay.pages.dev/roller-coaster-rush.html)
- [ASSOCIATIONS](https://quizverses-9d2f2.web.app/associations.html)
- [CATEGORY SIMULATION 3](https://thelearnquesters.pages.dev/category-simulation-3.html)
- [CATEGORY SOLITAIRE](https://learnquester.pages.dev/category-solitaire.html)
- [STEAL BRAINROT DUEL](https://themindplays.pages.dev/steal-brainrot-duel.html)
- [CATEGORY CARE](https://learnquester.pages.dev/category-care.html)
- [GEOMETRY DASH MAZE MAPS V2](https://quizverses.github.io/geometry-dash-maze-maps-v2.html)
- [CATEGORY AIRPLANE](https://thelearnquesters.pages.dev/category-airplane.html)
- [CATEGORY BATTLE 2](https://thelearnquesters.pages.dev/category-battle-2.html)
- [SAVE BABY CAPYBARAS PULL PIN](https://quizverses-9d2f2.web.app/save-baby-capybaras-pull-pin.html)
- [MEMORY MATCH MAGIC](https://thelearnquesters.pages.dev/memory-match-magic.html)
- [SPACE CLEANER](https://themindplay.pages.dev/space-cleaner.html)
- [INDEX3](https://learnquester.pages.dev/index3.html)
- [CATEGORY CAT55](https://iskillplay.web.app/category-cat55.html)
- [CANDY RAIN 5](https://studyquests.github.io/candy-rain-5.html)
- [COCKTAILZ](https://themindskillplayplay.pages.dev/cocktailz.html)
- [TURBO TRUCKS RACE](https://themindskillplayplay.pages.dev/turbo-trucks-race.html)
- [CATEGORY DRESS UP](https://thelearnquesters.pages.dev/category-dress-up.html)
- [SWEET MERGE](https://studyplaying.github.io/sweet-merge.html)
- [RAINBOW FRIENDS HIDE AND SEEK](https://themindplays.pages.dev/rainbow-friends-hide-and-seek.html)
- [ROYAL GARDEN MATCH](https://themindplaying.web.app/royal-garden-match.html)
- [PRESS A TO PARTY](https://skillplay.github.io/press-a-to-party.html)
- [DARLING DOLL](https://learnquester.pages.dev/darling-doll.html)
- [PRINCESS RESCUE SAVE GIRL](https://themindskillplayplay.pages.dev/princess-rescue-save-girl.html)
- [POPCORN STACK](https://themindplay.pages.dev/popcorn-stack.html)
- [FIND IT OUT COLORFUL BOOK](https://learnquester.pages.dev/find-it-out-colorful-book.html)
- [CONNECT BALLS NEW YEAR PUZZLES](https://studyquests.github.io/connect-balls-new-year-puzzles.html)
- [SHINE SEEK](https://themindplay.pages.dev/shine-seek.html)
- [STICKMAN IN SPACE](https://themindskillplayplay.pages.dev/stickman-in-space.html)
- [EMERGENCY JAM](https://learnquester.pages.dev/emergency-jam.html)
- [MATCHING PUZZLE](https://quizverses.pages.dev/matching-puzzle.html)
- [DINO SLIDE](https://themindplay.pages.dev/dino-slide.html)
- [CATEGORY ART32](https://themindplays.pages.dev/category-art32.html)
- [BRAINROT ICE TRUCK](https://thelearnquesters.pages.dev/brainrot-ice-truck.html)
- [CATEGORY MONSTER206](https://themindskillplayplay.pages.dev/category-monster206.html)
- [SOLITAIRE KLONDIKE](https://skillplay.github.io/solitaire-klondike.html)
- [CATEGORY BRAIN260](https://learnquester.pages.dev/category-brain260.html)
- [DARTS JAM](https://quizverses.github.io/darts-jam.html)
- [BLOCK CRAFT 3D SCHOOL](https://studyplaying.github.io/block-craft-3d-school.html)
- [JUST MAHJONG](https://themindskillplayplay.pages.dev/just-mahjong.html)
- [ROYAL KITCHEN THE LOST KING](https://quizverses.github.io/royal-kitchen-the-lost-king.html)
- [CATEGORY PUZZLE 7](https://studyplaying.github.io/category-puzzle-7.html)
- [CRICKET CLASH PONG](https://thelearnquesters.pages.dev/cricket-clash-pong.html)
- [FIND IT OUT COLORFUL BOOK](https://themindplaying.web.app/find-it-out-colorful-book.html)
