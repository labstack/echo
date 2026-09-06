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
- [POPPY STRIKE 5](https://learnquester.pages.dev/poppy-strike-5.html)
- [GANGSTA ISLAND CRIME CITY](https://quizverses.pages.dev/gangsta-island-crime-city.html)
- [CATEGORY RUNNING107](https://quizverses.pages.dev/category-running107.html)
- [BRICK BLAZE](https://studyquests.github.io/brick-blaze.html)
- [CATEGORY HORROR](https://studyplayings.pages.dev/category-horror.html)
- [CATEGORY FASHION105](https://thelearnquester.web.app/category-fashion105.html)
- [MARBLE BUBBLE LEGEND](https://learnquester.github.io/marble-bubble-legend.html)
- [ROAD TO 7](https://theskillquest.pages.dev/road-to-7.html)
- [FOOD JAM](https://quizverses.pages.dev/food-jam.html)
- [FIND THE SPRUNKI](https://studyquests.github.io/find-the-sprunki.html)
- [CATEGORY INCREMENTAL388](https://studyquests.github.io/category-incremental388.html)
- [LAST UFO DEFENSE](https://studyquests.github.io/last-ufo-defense.html)
- [ASSOCIATIONS](https://quizverses-9d2f2.web.app/associations.html)
- [CATEGORY MAHJONG CONNECT](https://studyplayings.pages.dev/category-mahjong-connect.html)
- [TERMS](https://brainquests.github.io/terms.html)
- [DIAMONDZ](https://studyplayings.web.app/diamondz.html)
- [INDEX15](https://quizverses.pages.dev/index15.html)
- [CATEGORY WEB PROXY](https://studyquests.github.io/category-web-proxy.html)
- [CATEGORY INTERSTELLAR](https://learnquester.github.io/category-interstellar.html)
- [CREEPY DRESS UP](https://quizverses-9d2f2.web.app/creepy-dress-up.html)
- [NO PAIN NO GAIN RAGDOLL SANDBOX](https://quizverses.pages.dev/no-pain-no-gain-ragdoll-sandbox.html)
- [TRAFFIC RUN PUZZLE](https://thelearnquester.web.app/traffic-run-puzzle.html)
- [UNLOCK THE BOLTS](https://learnquester.github.io/unlock-the-bolts.html)
- [WINTER MAHJONG](https://studyquests.github.io/winter-mahjong.html)
- [DIG FLOW SAVE WATER](https://quizverses.github.io/dig-flow-save-water.html)
- [FRUIT NINJA](https://quizverses.github.io/fruit-ninja.html)
- [CATEGORY MATCH 3](https://learnquester.pages.dev/category-match-3.html)
- [CATEGORY BASKETBALL 2](https://learnquester.pages.dev/category-basketball-2.html)
- [ASMR TATTOO TREATMENT](https://studyquests.github.io/asmr-tattoo-treatment.html)
- [CATEGORY SECURLY BYPASS](https://studyquests.github.io/category-securly-bypass.html)
- [CATEGORY PIXEL313](https://studyquests.github.io/category-pixel313.html)
- [PIRATES MAHJONG](https://quizverses.pages.dev/pirates-mahjong.html)
- [CATEGORY BATTLE524](https://thelearnquester.web.app/category-battle524.html)
- [THEO MORINIS MAGICAL RESORT](https://learnquester.github.io/theo-morinis-magical-resort.html)
- [SUGAR POP LAND](https://studyquests.github.io/sugar-pop-land.html)
- [OFFLINE FPS ROYALE](https://quizverses.github.io/offline-fps-royale.html)
- [CATEGORY BIKE 2](https://thelearnquester.web.app/category-bike-2.html)
- [DIGITAL CIRCUS FIND THE DIFFERENCES](https://quizverses.github.io/digital-circus-find-the-differences.html)
- [INDEX7](https://studyquests.github.io/index7.html)
- [KINGDOM WARS TD](https://studyquests.pages.dev/kingdom-wars-td.html)
- [CANDY POP MANIA](https://studyquests.github.io/candy-pop-mania.html)
- [CATEGORY MAGIC46](https://learnquester.github.io/category-magic46.html)
- [CATEGORY 204828](https://thelearnquester.web.app/category-204828.html)
- [CATEGORY SCHOOL](https://studyquests.github.io/category-school.html)
- [CATEGORY IO](https://studyplaying.github.io/category-io.html)
- [CATEGORY MMO24](https://learnquester.github.io/category-mmo24.html)
- [PRACTICE ON ME](https://studyquests.github.io/practice-on-me.html)
- [SCP 173 ESCAPE](https://studyquests.github.io/scp-173-escape.html)
- [UNBLOCK IT 3D](https://thelearnquester.web.app/unblock-it-3d.html)
- [UNSCREW THEM ALL](https://studyplayings.web.app/unscrew-them-all.html)
- [RUN FROM BABA YAGA](https://learnquester.github.io/run-from-baba-yaga.html)
- [CATEGORY BYPASS](https://learnquester.pages.dev/category-bypass.html)
- [SOLITAIRE TAIL](https://learnquester.github.io/solitaire-tail.html)
- [SNEAKY FRIENDS](https://studyquests.github.io/sneaky-friends.html)
- [FALLING ART RAGDOLL SIMULATOR](https://learnquester.github.io/falling-art-ragdoll-simulator.html)
- [SLIDE BLOCK PUZZLE](https://studyquesthub.web.app/slide-block-puzzle.html)
- [CATEGORY SOLITAIRE](https://learnquester.pages.dev/category-solitaire.html)
- [BATTALION COMMANDER 2](https://studyquests.github.io/battalion-commander-2.html)
- [MAGIC BUBBLES](https://studyquests.pages.dev/magic-bubbles.html)
- [SUPERWINGS SUBWAY](https://quizverses.github.io/superwings-subway.html)
- [HEX SENSE](https://studyquests.github.io/hex-sense.html)
- [CUBE DROP PUZZLE](https://quizverses.github.io/cube-drop-puzzle.html)
- [SUPER RACING GT DRAG PRO](https://studyplayings.web.app/super-racing-gt-drag-pro.html)
- [STICKMAN FIGHT PRO](https://learnquester.github.io/stickman-fight-pro.html)
- [URUS CITY DRIVER](https://studyquests.pages.dev/urus-city-driver.html)
- [MR DUDE KING OF THE HILL](https://quizverses.github.io/mr-dude-king-of-the-hill.html)
- [VENETIAN LOVE AFFAIR](https://studyquesthub.web.app/venetian-love-affair.html)
- [CATEGORY BOOKMARK](https://thelearnquester.web.app/category-bookmark.html)
- [PET FALL](https://learnquester.github.io/pet-fall.html)
- [MR CAPPUCCINO ASSASSINO](https://studyquesthub.web.app/mr-cappuccino-assassino.html)
- [CATEGORY TITANIUMNETWORK](https://studyquests.github.io/category-titaniumnetwork.html)
- [UNSCREW WOOD PUZZLE](https://studyquests.github.io/unscrew-wood-puzzle.html)
- [ANIME COUPLE AVATAR MAKER](https://quizverses-9d2f2.web.app/anime-couple-avatar-maker.html)
- [SIEGE BREAK](https://learnquester.github.io/siege-break.html)
- [LABUBU DOLL MUKBANG ASMR UNBLOCKED](https://studyplayings.web.app/labubu-doll-mukbang-asmr-unblocked.html)
- [BOW AND ARROW](https://learnquester.github.io/bow-and-arrow.html)
- [CATEGORY FPS 2](https://quizverses.pages.dev/category-fps-2.html)
- [DUMMIES WORLD CUP](https://studyplayings.pages.dev/dummies-world-cup.html)
- [STICK FIGHT THE CHAOS](https://studyplayings.web.app/stick-fight-the-chaos.html)
- [CATEGORY UNBLOCKED WEBSITE](https://quizverses.pages.dev/category-unblocked-website.html)
- [STICKMAN ESCAPES FROM PRISON](https://quizverses.pages.dev/stickman-escapes-from-prison.html)
- [PING PONG BATTLE TABLE TENNIS](https://studyquesthub.web.app/ping-pong-battle-table-tennis.html)
- [CATEGORY CASUAL 7](https://learnquester.pages.dev/category-casual-7.html)
- [HIT BALL](https://quizverses.github.io/hit-ball.html)
- [TOKA BOKA HOME CLEAN UP DESIGN](https://studyquests.github.io/toka-boka-home-clean-up-design.html)
- [BRAWL STARS SOUND](https://studyquesthub.web.app/brawl-stars-sound.html)
- [THE WHITE ROOM 5](https://learnquester.github.io/the-white-room-5.html)
- [SITEMAP](https://brainquests.github.io/sitemap.html)
- [ZOMBIE RAFT](https://studyquests.pages.dev/zombie-raft.html)
- [RIDE SHOOTER](https://learnquester.github.io/ride-shooter.html)
- [CATEGORY CASUAL 3](https://learnquester.pages.dev/category-casual-3.html)
- [321 CHOOSE THE DIFFERENT](https://thelearnquester.web.app/321-choose-the-different.html)
- [CATEGORY PUZZLE 5](https://learnquester.pages.dev/category-puzzle-5.html)
- [RUNNING IN FOAM](https://studyplaying.github.io/running-in-foam.html)
- [MEMORY WARS](https://quizverses.github.io/memory-wars.html)
- [CATEGORY MINECRAFT](https://studyplayings.pages.dev/category-minecraft.html)
- [BALL TOWER OF HELL](https://learnquester.github.io/ball-tower-of-hell.html)
- [CATEGORY ESCAPE 2](https://studyplayings.web.app/category-escape-2.html)
- [FOOTBALL PENALTY 2026](https://studyquests.github.io/football-penalty-2026.html)
- [ALIEN HUNTERS](https://studyquests.github.io/alien-hunters.html)
- [JELLY RUN 2048](https://quizverses.github.io/jelly-run-2048.html)
- [SOLITAIRE EMPEROR SECRETS OF FATE](https://learnquester.github.io/solitaire-emperor-secrets-of-fate.html)
- [ZIP ZAP](https://quizverses.github.io/zip-zap.html)
- [FASHION PRINCESS DRESS UP](https://thelearnquester.web.app/fashion-princess-dress-up.html)
- [COLOR WAVEE](https://thelearnquester.web.app/color-wavee.html)
- [HOTFOOT BASEBALL](https://quizverses.github.io/hotfoot-baseball.html)
- [SUPER SLIME](https://quizverses.github.io/super-slime.html)
- [MINI GAMES RELAX COLLECTION 2](https://studyquests.github.io/mini-games-relax-collection-2.html)
- [CATEGORY MISSION207](https://studyplayings.pages.dev/category-mission207.html)
- [BLOCK LEGENDS](https://quizverses.github.io/block-legends.html)
- [CATEGORY HERO](https://learnquester.pages.dev/category-hero.html)
- [CANDY CRUNCH SUGAR ESCAPE](https://studyplaying.github.io/candy-crunch-sugar-escape.html)
- [ASMR WASHING FIXING](https://studyplaying.github.io/asmr-washing-fixing.html)
- [CATEGORY ESCAPE 2](https://learnquester.pages.dev/category-escape-2.html)
- [CATEGORY COOKING46](https://studyquests.github.io/category-cooking46.html)
- [CATEGORY IO](https://thelearnquester.web.app/category-io.html)
- [CANDY SMASH](https://quizverses-9d2f2.web.app/candy-smash.html)
- [CATEGORY SOLITAIRE27](https://studyquests.github.io/category-solitaire27.html)
- [CATEGORY HORROR](https://studyquests.github.io/category-horror.html)
- [CATEGORY FPS](https://thelearnquester.web.app/category-fps.html)
- [CATEGORY IDLE448](https://learnquester.pages.dev/category-idle448.html)
- [MERGE GALAXY](https://studyquests.pages.dev/merge-galaxy.html)
- [PRIVACY](https://cryptotify9.onrender.com/privacy.html)
- [CATEGORY SHOOTER](https://thelearnquester.web.app/category-shooter.html)
- [FRUIT GOALS MATCH](https://studyplaying.github.io/fruit-goals-match.html)
- [CARD MASTER](https://learnquester.github.io/card-master.html)
- [CATEGORY HORROR](https://learnquester.pages.dev/category-horror.html)
- [CATEGORY SNAKE](https://studyplayings.pages.dev/category-snake.html)
- [BULL RUNNER](https://studyquests.pages.dev/bull-runner.html)
- [STICK HERO BATTLE](https://learnquester.github.io/stick-hero-battle.html)
