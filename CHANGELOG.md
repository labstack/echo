# Changelog

## v4.2.1 - 2020-03-08

**Important notes**

Due to a datarace the config parameters for the newly added timeout middleware required a change.
See the [docs](https://echo.labstack.com/middleware/timeout).
A performance regression has been fixed, even bringing better performance than before for some routing scenarios.

**Fixes**

* Fix performance regression caused by path escaping (#1777, #1798, #1799, aldas)
* Avoid context canceled errors (#1789, clwluvw)
* Improve router to use on stack backtracking (#1791, aldas, stffabi)
* Fix panic in timeout middleware not being not recovered and cause application crash (#1794, aldas)
* Fix Echo.Serve() not serving on HTTP port correctly when TLSListener is used (#1785, #1793, aldas)
* Apply go fmt (#1788, Le0tk0k)
* Uses strings.Equalfold (#1790, rkilingr)
* Improve code quality (#1792, withshubh)

This release was made possible by our **contributors**:
aldas, clwluvw, lammel, Le0tk0k, maciej-jezierski, rkilingr, stffabi, withshubh

## v4.2.0 - 2020-02-11

**Important notes**

The behaviour for binding data has been reworked for compatibility with echo before v4.1.11 by
enforcing `explicit tagging` for processing parameters. This **may break** your code if you 
expect combined handling of query/path/form params.
Please see the updated documentation for [request](https://echo.labstack.com/guide/request) and [binding](https://echo.labstack.com/guide/request)

The handling for rewrite rules has been slightly adjusted to expand `*` to a non-greedy `(.*?)` capture group. This is only relevant if multiple asterisks are used in your rules.
Please see [rewrite](https://echo.labstack.com/middleware/rewrite) and [proxy](https://echo.labstack.com/middleware/proxy) for details.

**Security**

* Fix directory traversal vulnerability for Windows (#1718, little-cui)
* Fix open redirect vulnerability with trailing slash (#1771,#1775 aldas,GeoffreyFrogeye)

**Enhancements**

* Add Echo#ListenerNetwork as configuration (#1667, pafuent)
* Add ability to change the status code using response beforeFuncs (#1706, RashadAnsari)
* Echo server startup to allow data race free access to listener address
* Binder: Restore pre v4.1.11 behaviour for c.Bind() to use query params only for GET or DELETE methods (#1727, aldas)
* Binder: Add separate methods to bind only query params, path params or request body (#1681, aldas)
* Binder: New fluent binder for query/path/form parameter binding (#1717, #1736, aldas)
* Router: Performance improvements for missed routes (#1689, pafuent)
* Router: Improve performance for Real-IP detection using IndexByte instead of Split (#1640, imxyb)
* Middleware: Support real regex rules for rewrite and proxy middleware (#1767)
* Middleware: New rate limiting middleware (#1724, iambenkay)
* Middleware: New timeout middleware implementation for go1.13+ (#1743, )
* Middleware: Allow regex pattern for CORS middleware (#1623, KlotzAndrew)
* Middleware: Add IgnoreBase parameter to static middleware (#1701, lnenad, iambenkay)
* Middleware: Add an optional custom function to CORS middleware to validate origin (#1651, curvegrid)
* Middleware: Support form fields in JWT middleware (#1704, rkfg)
* Middleware: Use sync.Pool for (de)compress middleware to improve performance (#1699, #1672, pafuent)
* Middleware: Add decompress middleware to support gzip compressed requests (#1687, arun0009)
* Middleware: Add ErrJWTInvalid for JWT middleware (#1627, juanbelieni)
* Middleware: Add SameSite mode for CSRF cookies to support iframes (#1524, pr0head)

**Fixes**

* Fix handling of special trailing slash case for partial prefix (#1741, stffabi)
* Fix handling of static routes with trailing slash (#1747)
* Fix Static files route not working (#1671, pwli0755, lammel)
* Fix use of caret(^) in regex for rewrite middleware (#1588, chotow)
* Fix Echo#Reverse for Any type routes (#1695, pafuent)
* Fix Router#Find panic with infinite loop (#1661, pafuent)
* Fix Router#Find panic fails on Param paths (#1659, pafuent)
* Fix DefaultHTTPErrorHandler with Debug=true (#1477, lammel)
* Fix incorrect CORS headers (#1669, ulasakdeniz)
* Fix proxy middleware rewritePath to use url with updated tests (#1630, arun0009)
* Fix rewritePath for proxy middleware to use escaped path in (#1628, arun0009)
* Remove unless defer (#1656, imxyb)

**General**

* New maintainers for Echo: Roland Lammel (@lammel) and Pablo Andres Fuente (@pafuent)
* Add GitHub action to compare benchmarks (#1702, pafuent)
* Binding query/path params and form fields to struct only works for explicit tags (#1729,#1734, aldas)
* Add support for Go 1.15 in CI (#1683, asahasrabuddhe)
* Add test for request id to remain unchanged if provided (#1719, iambenkay)
* Refactor echo instance listener access and startup to speed up testing (#1735, aldas)
* Refactor and improve various tests for binding and routing
* Run test workflow only for relevant changes (#1637, #1636, pofl)
* Update .travis.yml (#1662, santosh653)
* Update README.md with an recents framework benchmark (#1679, pafuent)

This release was made possible by **over 100 commits** from more than **20 contributors**:
asahasrabuddhe, aldas, AndrewKlotz, arun0009, chotow, curvegrid, iambenkay, imxyb, 
juanbelieni,  lammel, little-cui, lnenad, pafuent, pofl, pr0head, pwli, RashadAnsari, 
rkfg, santosh653, segfiner, stffabi, ulasakdeniz
