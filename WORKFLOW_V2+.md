# Transition of Gobot to v2+

## Problem Description

see <https://github.com/hybridgroup/gobot/issues/920>

## Analysis of uuid

### Version history

All is retrieved from repo (e.g. README.md). No tests done.

|Version|Release|go.mod                    |Date(code/release)   |Setup (go get)/Usage(import)|Go  |
|-------|-------|--------------------------|---------------------|----------------------------|----|
|v1.0.0 |-      |-                         |2016-03-24           |"github.com/gofrs/uuid"     |-   |
|v1.1.0 |-      |-                         |2016-06-07           |"github.com/gofrs/uuid"     |-   |
|v1.2.0 |-      |-                         |2018-01-03           |"github.com/gofrs/uuid"     |1.2 |
|v2.0.0 |x      |-                         |2018-07-18           |"github.com/gofrs/uuid"     |1.7 |
|v2.0.1 |x      |-                         |2018-07-19           |"github.com/gofrs/uuid"     |1.7 |
|v2.1.0 |x      |-                         |2018-07-21           |"github.com/gofrs/uuid"     |1.7 |
|v3.0.0 |x      |-                         |2018-07-18           |"github.com/gofrs/uuid"     |1.7 |
|v3.1.0 |x      |-                         |2018-08-22           |"github.com/gofrs/uuid"     |1.7 |
|v3.1.1 |x      |"github.com/gofrs/uuid/v3"|2018-08-30/2018-09-01|"github.com/gofrs/uuid"     |1.7 |
|v3.1.2 |x      |"github.com/gofrs/uuid/v3"|2018-10-30           |"github.com/gofrs/uuid"     |1.7 |
|v3.2.0 |x      |(1*)                      |2019-01-11           |"github.com/gofrs/uuid"     |1.7 |
|v3.3.0 |x      |-                         |2019-05-10/2020-05-05|"github.com/gofrs/uuid"     |1.7 |
|v3.4.0 |x      |-                         |2019-05-10/2020-12-30|"github.com/gofrs/uuid"     |1.7 |
|v4.0.0 |x      |-                         |2020-12-30           |"github.com/gofrs/uuid"     |1.7 |
|v4.1.0 |x      |-                         |2021-10-16           |"github.com/gofrs/uuid"     |1.7 |
|v4.2.0 |x      |-                         |2021-11-26           |"github.com/gofrs/uuid"     |1.7 |
|v4.3.0 |x      |-                         |2022-09-10           |"github.com/gofrs/uuid"     |1.7 |
|v4.3.1 |x      |-                         |2022-10-31           |"github.com/gofrs/uuid"     |1.7 |
|v4.4.0 |x      |-                         |2023-01-26           |"github.com/gofrs/uuid"     |1.7 |
|v5.0.0 |x      |"github.com/gofrs/uuid/v5"|2023-02-10           |                            |1.19|

This section was added to the README (1*):

```md
Go 1.11 Modules

As of v3.2.0, this repository no longer adopts Go modules, and v3.2.0 no longer has a go.mod file. As a result, v3.2.0
also drops support for the github.com/gofrs/uuid/v3 import path. Only module-based consumers are impacted. With the
v3.2.0 release, all gofrs/uuid consumers should use the github.com/gofrs/uuid import path.

An existing module-based consumer will continue to be able to build using the github.com/gofrs/uuid/v3 import path using
any valid consumer go.mod that worked prior to the publishing of v3.2.0, but any module-based consumer should start using
the github.com/gofrs/uuid import path when possible and must use the github.com/gofrs/uuid import path prior to upgrading
to v3.2.0.

Please refer to Issue #61 and Issue #66 for more details.
```

<https://github.com/gofrs/uuid/issues/61>
<https://github.com/gofrs/uuid/pull/66>
<https://github.com/gofrs/uuid/pull/67>

### Differences of UUID to gobot

UUID starts without modules and runs **into v2+ versions without modules**. Afterwards it tries to switch to modules within
v3.1.x and make a rollback. The modules v3.1.1 and v3.1.2 exists as tags and releases, but can not be found as versions in
the [history of pre-modules](https://pkg.go.dev/github.com/gofrs/uuid?tab=versions), but in the
[history of v3](https://pkg.go.dev/github.com/gofrs/uuid/v3). And of course the
[v5 history](https://pkg.go.dev/github.com/gofrs/uuid/v5?tab=versions) is present.

Gobot starts without go modules and migrated to it with v1.14.0 (released on 2019-10-15), before v2+.

> All versions since v2 without a go.mod file will automatically marked with "+incompatible", although the release (tag)
> is created without this extension.

## Analysis of datadog-api-client-go

In contrast to gobot, UUID has no module releases prior v2+, so another project was chosen which equals in this point.

### Description of Version History

The module mode was introduces together with the [initial creation](https://github.com/DataDog/datadog-api-client-go/commit/f6f306c7543a1fa08fd1f4949dd7adc02cb370ee)
of the project. The switch to v2+ was done on 2022-08-01 from [v1.16.0 to v2.0.0](https://github.com/DataDog/datadog-api-client-go/compare/v1.16.0...v2.0.0),
this equality to gobot is only a coincidence. The same things were done like in gobot (adjust go.mod to "/v2" and in
substitute in over ~700 files all internal paths to "/v2"), see this [commit](
https://github.com/DataDog/datadog-api-client-go/commit/c926bb85c001e473c360b537c4bd7d861188cef1).

### Differences of datadog-api-client-go to gobot

Not any version without module mode was released for datadog-api-client-go. Gobot has some releases prior v1.14.0
without module mode.

The module name is a common "github.com/DataDog/datadog-api-client-go", gobot uses "gobot.io/x/gobot" as module name.

## Analysis of go-micro

Package [go-micro](https://github.com/go-micro/go-micro) has releases without module mode prior v2+. It has also releases
prior v2+ with module mode. And finally it uses a, from github location different, module name "go-micro.dev". This seems
perfectly fits the gobot situation. Because there are 320 releases right now, we skip all non relevant releases/tags.

|Version|Release|go.mod                        |Date(code/release)   |Go  |Notes|
|-------|-------|------------------------------|---------------------|----|-----|
|v0.1.0 |x      |-                             |2017-05-11           |-   | |
|v0.23.0|x      |-                             |2019-01-29           |-   |information can be found in [pkg.go.dev](https://pkg.go.dev/github.com/asim/go-micro)|
|v0.24.0|x      |"github.com/micro/go-micro"   |2019-01-30           |-   | |
|v1.0.0 |x      |"github.com/micro/go-micro"   |2019-03-05           |-   | |
|v1.6.0 |x      |"github.com/micro/go-micro"   |2019-06-07           |1.12| |
|v1.11.0|x      |"github.com/micro/go-micro"   |2019-10-01           |1.13| |
|v2.0.0 |x      |"github.com/micro/go-micro/v2"|2020-01-30           |1.13|[347 files changed](https://github.com/go-micro/go-micro/commit/f23638c0360a542c1b814d5ef3429439ba6d8807), no cut off or removing of go.mod found|
|v3.0.0 |x      |"github.com/asim/go-micro/v3" |2021-01-20           |1.13|[534 files changed](https://github.com/go-micro/go-micro/commit/dc8236ec05edf485eb6b4c8540787f81157b30e8), renaming done by intention|
|v3.5.2 |x      |"github.com/asim/go-micro/v3" |2021-07-06           |1.16| |
|v4.0.0 |x      |"go-micro.dev/v4"             |2021-10-12           |1.16|[705 files changed](https://github.com/go-micro/go-micro/commit/1cd7cfaa6cbb7d4cd9cba76ae2374469649a81d4), renaming done by intention|

> Possibly the repository name has changed over the time from "github.com/micro" to "github.com/go-micro". Although the
> module name was changed over the time, the relevant part can be seen in [pkg.go.dev](https://pkg.go.dev/github.com/micro/go-micro?tab=versions).
> Which means there coexist v0, v1, v2 and v3 modules with the same name. Together with the investigation on github,
> there was most likely no problem with smooth transition of this versions. Gobot transition was done in the same way,
> but do not work.
> Do not confuse with <https://github.com/micro/micro>, which is a slightly different package, but has the same fit to
> gobot, except the module name (it uses "github.com/micro/micro").

## Projects switching to v2+ before module mode

* <https://github.com/godbus/dbus/tree/v5.0.0>
* <https://github.com/labstack/echo/releases/tag/v3.3.6>
* <https://github.com/redis/go-redis/tree/v7.0.0-beta> (tagged, first release with 8.x)
* <https://github.com/DataDog/datadog-go/releases/tag/v5.0.0>

## Projects switching to module mode before v2.0.0

* <https://github.com/urfave/cli> (use different branches for v1, v2 and master, can't be found in [pkg.go.dev](https://pkg.go.dev/search?q=urface&m=))
* <https://github.com/DataDog/dd-trace-go> (not switched yet, latest 1.51.0)
* <https://github.com/DataDog/datadog-api-client-go/tree/v1.0.0-beta.1> (but not any version without go.mod)
* <https://pkg.go.dev/github.com/micro/micro/v3>

## When a v2 Subfolder will help

If you would like to support older Go releases than Go 1.9.7+, 1.10.3+, the additional mechanics to release a v2+
module are:

* Create a v2 directory and place a new go.mod file in that directory. The module path must end with /v2.
* Copy or move the code into that v2 directory.
* Update import paths to include /v2.
* Tag the release with v2.0.0.

reference: <https://github.com/golang/go/issues/25967>

## When a v2 Branch will help

In GOPATH mode, go get will download the default branch. If it is set to v2 for the repo, so GOPATH users will get
something pretty recent.

> This will only work if the v2 branch is the default one.

## Conclusion

If gobot do not need to support users with older Go versions than 1.11 (before module mode), and also do not support
GOPATH mode, there is no need for introducing a v2 branch or a v2 subfolder.

The combination (or any) of the following seems to cause the not working v2+ module release:

* introducing of module mode before v2.0.0 ([v1.14.0](https://github.com/hybridgroup/gobot/releases/tag/v1.14.0))
* there were some releases without go.mod before (e.g. [v1.13.0](https://github.com/hybridgroup/gobot/tree/v1.13.0))
* the path is not "github.com/hybridgroup/gobot" but "gobot.io/x/gobot"

> Prior v2+, all released versions (0.11.0..v1.16.0) are listed in [pkg.go.dev](https://pkg.go.dev/gobot.io/x/gobot?tab=versions),
> including the defective version "v2.0.0+incompatible", but no module "https://pkg.go.dev/gobot.io/x/gobot/v2" exists.

## Further ideas

* release a v2.x without modules and afterwards release a v3.0.0 with modules (e.g. together with go and module upgrade)

## Package recognition for different methods

This will result in a "basic" version (GOPATH mode compatible):

* release a v0 or v1 version with or without go.mod file
* release a v2+ version without a go.mod file
* release a v2+ version with a go.mod file, but not adjusted module name to "/v2+" (most likely problems on download)

This should result in a v2+ version in module mode:

* create a go.mod with "/v2+" module name
* adjust "internal" dependencies to the new name
