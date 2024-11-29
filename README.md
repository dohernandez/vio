# vio

[![Build Status](https://github.com/dohernandez/vio/workflows/test-unit/badge.svg)](https://github.com/dohernandez/vio/actions?query=branch%3Amain+workflow%3Atest-unit)
![Code lines](https://sloc.xyz/github/dohernandez/vio/?category=code)
![Comments](https://sloc.xyz/github/dohernandez/vio/?category=comments)


## Table of Contents
- [Table of Contents](#table-of-contents)
- [Overview](#overview)
- [Getting started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Development](#development)
        - [Running the service locally](#running-the-service-locally)
        - [Generate code from proto file](#generate-code-from-proto-file)
    - [Testing](#testing)
        - [Testing locally](#testing-locally)
        - [Benchmarking](#benchmarking)
    - [Metrics](#metrics)
    - [Migrations](#migrations)
- [Enhancement](#enhancement)

## Overview

The current architecture of the service is described in the [ARCHITECTURE.md](./ARCHITECTURE.md) document.

For more in-depth explanations and considerations on architectural choices for the service, please refer to our [Architecture Decision Records](./resources/adr) folder.

If you want to submit an architectural change to the service, please create a new entry in the ADR folder [using the template provided](./resources/adr/template.md) and open a new Pull Request for review. Each ADR should have a prefix with the consecutive number and a name. For example `002-implement-server-streaming-rpc-get-geolocation-data.md`

## Getting started

### Prerequisites

To develop and run this application on your machine, you must have `make` &&` docker` && `docker-compose` installed.

The service uses `dep` to manage its dependencies. All the dependencies can be installed using the following `make` command:

```shell
make deps
```

[[table of contents]](#table-of-contents)

### Development

#### Running the service locally

Run app with `docker-compose` dependencies.

First generate an `.env` file which the environment values required by the service such as `APP_GRPC_PORT` and `DATABASE_DSN`. You can run the following `make` command:

```
make envfile
```

This command will generate the `.env` file from the `.env.template`. Make sure the env variables defined in the file `.env` meet your expectation.

**Note:** No need to edit the `.env.template` before running the command, the flow is that you generate the `.env` file from `.env.template` and after that edit the `.env` if needed.

After the `.env` file is generated, you can start the app by running

```shell
docker-compose up -d
```

or using the `make` command

```shell
make start-dc
```

To destroy the app run the command:

```shell
docker-compose down -v
```

or using the `make` command

```shell
make stop-dc
```

[[table of contents]](#table-of-contents)

#### Generate code from proto file

The service implements grpc based on the proto definition. The proto file with the service definition can be found `resources/proto/service.proto`.

To generate the go code based on the proto definition and the api documentation run the `make` command

```shell
make proto-gen
```

The Go files generated based on the command can be found `pkg/proto` folder and the swagger definition can be found `resources/swagger/service.swagger.json`.

[[table of contents]](#table-of-contents)

### Testing

The server follows unit testing and integration test with [godog]/(https://github.com/cucumber/godog) the official Cucumber BDD framework for Golang. Unit testing make sure the logic of the application is sounds and integrations test make sure the business logic of the different uses cases covered are sound.

Unit tests reside with the application source code, as per Golang recommendation. Use the command `make` command to run the unit tests:

```shell
make test-unit
```

Integration tests reside in `features` folder, using [godog]/(https://github.com/cucumber/godog). Use the command `make` command to run the integration tests:

```shell
make test-integration
```

**Note**: Use the command `make stop-deps` to stop the containers.

To run the whole test suites of unit and integration tests:

```shell
make test
```

You can also run

```shell
make lint
```

to make sure your changes follow our coding standards.

[[table of contents]](#table-of-contents)

#### Testing locally

In case you wanna test the service using the sample data located in `resources/sample_data/data_dump.csv`, you do the following:

```makefile
make test-dc
```

- Play with the api by doing:
    - Using the gRPC client using [Evans](#evans).
    - Using the REST api by doing `curl http://localhost:8080/v1/geolocations/89.207.159.171`.
    - Or using the Swagger interface by going to http://localhost:8080/docs.

**Note**: Use the command `make stop-dc` to stop the containers.

[[table of contents]](#table-of-contents)

### Benchmarking

Benchmarking is a high level benchmark to assess global end to end performance impact of internal changes. To run the integration benchmarks:

```shell
  make bench-integration
```

**Note**: Use the command `make stop-deps` to stop the containers.

[[table of contents]](#table-of-contents)

#### Evans

For manual gRPC API inspection, the service allows gRPC reflection in dev environment.

To install Evans following the instructions from it GitHub page https://github.com/ktr0731/evans#installation.

[[table of contents]](#table-of-contents)

#### REST

Also, you can do test thro REST calls. For that you can use the service REST api documentation which uses Swagger interface.

Launch the service by [Running the service locally](#running-the-service-locally). This will make the service available in http://localhost:8080 (remember that the port is base on the configuration you provide in the `.env` file. This example is based on the `.env.template` configuration) and REST api documentation can be accessible on http://localhost:8080/docs.

[[table of contents]](#table-of-contents)

### Metrics

The service exposes some metrics such as:

- Database
- Go build info
- Current Go process
- Calls started/completed
- Histogram of response latency (seconds).

Metrics are available on http://localhost:8010/metrics

[[table of contents]](#table-of-contents)

### Migrations

Database migrations are stored in [`resources/migrations`](./resources/migrations) folder.

Migrations are run using [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate) tool,
embedded in the service's `Dockerfile` under `/bin/migrate`.

Each migration should have an `<name>.up.sql` and `<name>.down.sql` variants, further information can be seen here: https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md

The layout of the migration name should be as follows:

```
<current-date-string>_<migration-name>.(up|down).sql

Example (created in 2024-01-01 00:00:00):

    20240101000000_my_migration.up.sql
    20240101000000_my_migration.down.sql
```

For creating migration files you can use the following `make` command:

```shell
make create-migration NAME=<migration-name>
```

To run migration up use the following `make` command:

```shell
make migrate
```

**Note:** `DATABASE_DSN` env variable should be defined. In case it does not you can use the following `make` command:

```shell
make env migrate
```

If you run the above command from outside to docker network, make sure to have `127.0.0.1 postgres` in `/etc/hosts`.

[[table of contents]](#table-of-contents)

### Enhancement

* Add a cache layer to the service to avoid unnecessary calls to the database.
* Add a command to parse the geolocation data from an url.

[[table of contents]](#table-of-contents)