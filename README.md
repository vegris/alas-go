# alas-go

**alas-go** is an example Go application built using microservice architecture.

The application consists of two services:
- **kiwi** receives signed analytics events from users and saves them to Kafka.
- **orcrist** generates session tokens which users use to sign their events. 

## Repository overview

This repository contains 4 separate folders, each is a separate Go module:

- `kiwi/` - **kiwi** service
- `orcrist/` - **orcrist** service
- `shared/` - common bode used by both services:
    
    + `shared/application` contains app startup and shutdown routines
    + `shared/schemas` contains JSON schema compilation function
    + `shared/token` contains Token definition and functions to encode/decode it

- `integration_tests/` contains integration tests

## Usage

### Run locally

1. Setup required infrastructure (Redis, Postgres, Kafka) using provided `docker-compose.yaml`.

`make up`

2. Navigate to application folder (`kiwi/` or `orcrist/`) and start the service.

`make run`

### Run as a single Docker Compose deployment

Start `docker-compose.yaml` with the `integration` profile to deploy the infrastructure alongside the **kiwi** and **orcrist** services.

`make up-integration`

### Run integration tests

1. Start the application (either approach work).

Ensure `KAFKA_SYNC=1` is set to both services (can be added to `.env`) to disable in-service message buffering.

2. Navigate to `integration_tests/` and run tests.

`make test`
