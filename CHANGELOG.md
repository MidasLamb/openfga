# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Try to keep listed changes to a concise bulleted list of simple explanations of changes. Aim for the amount of information needed so that readers can understand where they would look in the codebase to investigate the changes' implementation, or where they would look in the documentation to understand how to make use of the change in practice - better yet, link directly to the docs and provide detailed information there. Only elaborate if doing so is required to avoid breaking changes or experimental features from ruining someone's day.

## [Unreleased]

## [0.3.2] - 2023-01-18

[Full changelog](https://github.com/openfga/openfga/compare/v0.3.1...v0.3.2)


### Added
* OpenTelemetry metrics integration with an `otlp` exporter (#360) - thanks @AlexandreBrg!

  To export OpenTelemetry metrics from an OpenFGA instance you can now provide the `otel-metrics` experimental flag along with the `--otel-telemetry-endpoint` and `--otel-telemetry-protocol` flags. For example,

  ```
  ./openfga run --experimentals=otel-metrics --otel-telemetry-endpoint=127.0.0.1:4317 --otel-telemetry-protocol=http
  ```

  For more information see the official documentation on [Experimental Features](https://openfga.dev/docs/getting-started/setup-openfga#experimental-features) and [Telemetry](https://openfga.dev/docs/getting-started/setup-openfga#telemetry-metrics-and-tracing).

* Type-bound public access support in the optimized ListObjects implementation (when the `list-objects-optimized` experimental feature is enabled) (#444)

### Fixed
* Tuple validations for models with schema version 1.1 (#446, #457)
* Evaluate rewrites on nested usersets in the optimized ListObjects implementation (#432)

## [0.3.1] - 2022-12-19

[Full changelog](https://github.com/openfga/openfga/compare/v0.3.0...v0.3.1)

### Added
* Datastore configuration flags to control connection pool settings
  `--datastore-max-open-conns`
  `--datastore-max-idle-conns`
  `--datastore-conn-max-idle-time`
  `--datastore-conn-max-lifetime`
  These flags can be used to fine-tune database connections for your specific deployment of OpenFGA.

* Log level configuration flags
  `--log-level` (can be one of ['none', 'debug', 'info', 'warn', 'error', 'panic', 'fatal'])

* Support for Experimental Feature flags
  A new flag `--experimentals` has been added to enable certain experimental features in OpenFGA. For more information see [Experimental Features](https://openfga.dev/docs/getting-started/setup-openfga#experimental-features).

### Security
* Patches [CVE-2022-23542](https://github.com/openfga/openfga/security/advisories/GHSA-m3q4-7qmj-657m) - relationship reads now respect type restrictions from prior models (#422).

## [0.3.0] - 2022-12-12

[Full changelog](https://github.com/openfga/openfga/compare/v0.2.5...v0.3.0)

This release comes with a few big changes:

### Support for [v1.1 JSON Schema](https://github.com/openfga/rfcs/blob/feat/add-type-restrictions-to-json-syntax/20220831-add-type-restrictions-to-json-syntax.md)

- You can now write your models in the [new DSL](https://github.com/openfga/rfcs/blob/type-restriction-dsl/20221012-add-type-restrictions-to-dsl-syntax.md)
which the Playground and the [syntax transformer](https://github.com/openfga/syntax-transformer) can convert to the
JSON syntax. Schema v1.1 allows for adding type restrictions to each assignable relation, and it can be used to
indicate cases such as "The folder's parent must be a folder" (and so not a user or a document).
  - This change also comes with breaking changes to how `*` and `<type>:*` are treated:
  - `<type>:*` is interpreted differently according to the model version. v1.0 will interpret it as a object of type
    `<type>` and id `*`, whereas v1.1 will interpret is as all objects of type `<type>`.
  - `*` is still supported in v1.0 models, but not supported in v1.1 models. A validation error will be thrown when
    used in checks or writes and it will be ignored when evaluating.
- Additionally, the change to v1.1 models allows us to provide more consistent validation when writing the model
instead of when issuing checks.

:warning: Note that with this release **models with schema version 1.0 are now considered deprecated**, with the plan to
drop support for them over the next couple of months, please migrate to version 1.1 when you can. Read more about
[migrating to the new syntax](https://openfga.dev/docs/modeling/migrating/migrating-schema-1-1).

### ListObjects changes

The response has changed to include the object type, for example:
```json
{ "object_ids": [ "a", "b", "c" ] }
```
to
```json
{ "objects": [ "document:a", "document:b", "document:c" ] }
```

We have also improved validation and fixed support for Contextual Tuples that were causing inaccurate responses to be
returned.

### ReadTuples deprecation

:warning:This endpoint is now marked as deprecated, and support for it will be dropped shortly. Please use Read with
no tuple key instead.


## [0.2.5] - 2022-11-07
### Security
* Patches [CVE-2022-39352](https://github.com/openfga/openfga/security/advisories/GHSA-3gfj-fxx4-f22w)

### Added
* Multi-platform container build manifests to releases (#323)

### Fixed
* Read RPC returns correct error when authorization model id is not found (#312)
* Throw error if `http.upstreamTimeout` config is less than `listObjectsDeadline` (#315)

## [0.2.4] - 2022-10-24
### Security
* Patches [CVE-2022-39340](https://github.com/openfga/openfga/security/advisories/GHSA-95x7-mh78-7w2r), [CVE-2022-39341](https://github.com/openfga/openfga/security/advisories/GHSA-vj4m-83m8-xpw5), and [CVE-2022-39342](https://github.com/openfga/openfga/security/advisories/GHSA-f4mm-2r69-mg5f)

### Fixed
* TLS certificate config path mappings (#285)
* Error message when a `user` field is invalid (#278)
* host:port mapping with unspecified host (#275)
* Wait for connection to postgres before starting (#270)


### Added
* Update Go to 1.19

## [0.2.3] - 2022-10-05
### Added
* Support for MySQL storage backend (#210). Thank you @MidasLamb!
* Allow specification of type restrictions in authorization models (#223). Note: Type restriction is not enforced yet, this just allows storing them.
* Tuple validation against type restrictions in Write API (#232)
* Upgraded the Postgres storage backend to use pgx v5 (#225)

### Fixed
* Close database connections after migration (#252)
* Race condition in streaming ListObjects (#255, #256)


## [0.2.2] - 2022-09-15
### Fixed
* Reject direct writes if only indirect relationship allowed (#114). Thanks @dblclik!
* Log internal errors at the grpc layer (#222)
* Authorization model validation (#224)
* Bug in `migrate` command (#236)
* Skip malformed tuples involving tuple to userset definitions (#234)

## [0.2.1] - 2022-08-30
### Added
* Support Check API calls on userset types of users (#146)
* Add backoff when connecting to Postgres (#188)

### Fixed
* Improve logging of internal server errors (#193)
* Use Postgres in the sample Docker Compose file (#195)
* Emit authorization errors (#144)
* Telemetry in Check and ListObjects APIs (#177)
* ListObjects API: respect the value of ListObjectsMaxResults (#181)


## [0.2.0] - 2022-08-12
### Added
* [ListObjects API](https://openfga.dev/api/service#/Relationship%20Queries/ListObjects)

  The ListObjects API provides a way to list all of the objects (of a particular type) that a user has a relationship with. It provides a solution to the [Search with Permissions (Option 3)](https://openfga.dev/docs/interacting/search-with-permissions#option-3-build-a-list-of-ids-then-search) use case for access-aware filtering on smaller object collections. It implements the [ListObjects RFC](https://github.com/openfga/rfcs/blob/main/20220714-listObjects-api.md).

  This addition brings with it two new server configuration options `--listObjects-deadline` and `--listObjects-max-results`. These configurations help protect the server from excessively long lived and large responses.

  > ⚠️ If `--listObjects-deadline` or `--listObjects-max-results` are provided, the endpoint may only return a subset of the data. If you provide the deadline but returning all of the results would take longer than the deadline, then you may not get all of the results. If you limit the max results to 1, then you'll get at most 1 result.

* Support for presharedkey authentication in the Playground (#141)

  The embedded Playground now works if you run OpenFGA using one or more preshared keys for authentication. OIDC authentication remains unsupported for the Playground at this time.


## [0.1.7] - 2022-07-29
### Added
* `migrate` CLI command (#56)

  The `migrate` command has been added to the OpenFGA CLI to assist with bootstrapping and managing database schema migrations. See the usage for more info.

  ```
  ➜ openfga migrate -h
  The migrate command is used to migrate the database schema needed for OpenFGA.

  Usage:
    openfga migrate [flags]

  Flags:
        --datastore-engine string   (required) the database engine to run the migrations for
        --datastore-uri string      (required) the connection uri of the database to run the migrations against (e.g. 'postgres://postgres:password@localhost:5432/postgres')
    -h, --help                      help for migrate
        --version uint              the version to migrate to. If omitted, the latest version of the schema will be used
  ```

## [0.1.6] - 2022-07-27
### Fixed
* Issue with embedded Playground assets found in the `v0.1.5` released docker image (#129)

## [0.1.5] - 2022-07-27
### Added
* Support for defining server configuration in `config.yaml`, CLI flags, or env variables (#63 #92 #100)

  `v0.1.5` introduces multiple ways to support a variety of server configuration strategies. You can configure the server with CLI flags, env variables, or a `config.yaml` file.

  Server config will be loaded in the following order of precedence:

    * CLI flags (e.g. `--datastore-engine`)
    * env variables (e.g. `OPENFGA_DATASTORE_ENGINE`)
    * `config.yaml`

  If a `config.yaml` file is provided, the OpenFGA server will look for it in `"/etc/openfga"`, `"$HOME/.openfga"`, or `"."` (the current working directory), in that order.

* Support for grpc health checks (#86)

  `v0.1.5` introduces support for the [GRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md). The server's health can be checked with the grpc or HTTP health check endpoints (the `/healthz` endpoint is just a proxy to the grpc health check RPC).

  For example,
  ```
  grpcurl -plaintext \
    -d '{"service":"openfga.v1.OpenFGAService"}' \
    localhost:8081 grpc.health.v1.Health/Check
  ```
  or, if the HTTP server is enabled, with the `/healthz` endpoint:
  ```
  curl --request GET -d '{"service":"openfga.v1.OpenFGAService"}' http://localhost:8080/healthz
  ```

* Profiling support (pprof) (#111)

  You can now profile the OpenFGA server while it's running using the [pprof](https://github.com/google/pprof/blob/main/doc/README.md) profiler. To enable the pprof profiler set `profiler.enabled=true`. It is served on the `/debug/pprof` endpoint and port `3001` by default.

* Configuration to enable/disable the HTTP server (#84)

  You can now enable/disable the HTTP server by setting `http.enabled=true/false`. It is enabled by default.

### Changed
* Env variables have a new mappings.

  Please refer to the [`.config-schema.json`](https://github.com/openfga/openfga/blob/main/.config-schema.json) file for a description of the new configurations or `openfga run -h` for the CLI flags. Env variables are   mapped by prefixing `OPENFGA` and converting dot notation into underscores (e.g. `datastore.uri` becomes `OPENFGA_DATASTORE_URI`). 

### Fixed
* goroutine leaks in Check resolution. (#113)

## [0.1.4] - 2022-06-27
### Added
* OpenFGA Playground support (#68)
* CORS policy configuration (#65)

## [0.1.2] - 2022-06-20
### Added
* Request validation middleware
* Postgres startup script

## [0.1.1] - 2022-06-16
### Added
* TLS support for both the grpc and HTTP servers
* Configurable logging formats including `text` and `json` formats
* OpenFGA CLI with a preliminary `run` command to run the server

## [0.1.0] - 2022-06-08
### Added
* Initial working implementation of OpenFGA APIs (Check, Expand, Write, Read, Authorization Models, etc..)
* Postgres storage adapter implementation
* Memory storage adapter implementation
* Early support for preshared key or OIDC authentication methods

[Unreleased]: https://github.com/openfga/openfga/compare/v0.3.2...HEAD
[0.3.2]: https://github.com/openfga/openfga/releases/tag/v0.3.2
[0.3.1]: https://github.com/openfga/openfga/releases/tag/v0.3.1
[0.3.0]: https://github.com/openfga/openfga/releases/tag/v0.3.0
[0.2.5]: https://github.com/openfga/openfga/releases/tag/v0.2.5
[0.2.4]: https://github.com/openfga/openfga/releases/tag/v0.2.4
[0.2.3]: https://github.com/openfga/openfga/releases/tag/v0.2.3
[0.2.2]: https://github.com/openfga/openfga/releases/tag/v0.2.2
[0.2.1]: https://github.com/openfga/openfga/releases/tag/v0.2.1
[0.2.0]: https://github.com/openfga/openfga/releases/tag/v0.2.0
[0.1.7]: https://github.com/openfga/openfga/releases/tag/v0.1.7
[0.1.6]: https://github.com/openfga/openfga/releases/tag/v0.1.6
[0.1.5]: https://github.com/openfga/openfga/releases/tag/v0.1.5
[0.1.4]: https://github.com/openfga/openfga/releases/tag/v0.1.4
[0.1.2]: https://github.com/openfga/openfga/releases/tag/v0.1.2
[0.1.1]: https://github.com/openfga/openfga/releases/tag/v0.1.1
[0.1.0]: https://github.com/openfga/openfga/releases/tag/v0.1.0
