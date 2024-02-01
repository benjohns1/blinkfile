---
title: Contribute
weight: 400
---
## PRs welcome!

Here's how to get started.
1. Find an [existing issue](https://github.com/benjohns1/blinkfile/issues) to work on or create a new one (please label as 'bug' or 'enhancement' as appropriate)
2. Use the issue comments to discuss what approach you want to take
3. Fork the [repo](https://github.com/benjohns1/blinkfile)
4. Get it running locally (see below)
5. Make your changes, including unit and acceptance tests
6. Submit a PR

## Run locally
Prerequisites: Docker, Make

### Build & run
```
make run
```

### Run full test suite
```
make test
```

## Development
Prerequisites: Go, NPM, Docker, Make

### Install dependencies
```
make install
```

### Run unit tests with coverage
```
make unit-test
```

### Run linter
```
make lint
```

### Run acceptance tests
```
make acceptance-test
```

### Open Cypress UI to run and develop acceptance tests
```
make acceptance-test-runner
```
Gherkin-style features are defined in `test/cypress/features`. Any scenarios tagged with `@pending` or `@implementing`
will _not_ be run in the CI pipeline. Scenarios tagged with `@implementing` will show up in the test runner locally.

Cypress implementation steps are defined in `test/cypress/steps`. See
[badeball's cypress-cucumber-preprocessor docs](https://github.com/badeball/cypress-cucumber-preprocessor/blob/master/docs/readme.md)
for more details.

### Run on host
Required environment variables:
- ADMIN_USERNAME
- ADMIN_PASSWORD

```
go run ./...
```

### CI/CD
Pipeline runs in github actions

#### Run the full test suite
This will run a workflow similar to what runs in the pipeline as a final check before push:
```
make test
```
You can also run individual steps in `.local/Makefile`.
