---
title: Contribute
layout: home
nav_order: 4
---
# Contribute
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
#### Run full test suite
This will run a flow similar to what runs in the pipeline as a final check before push:
```
make test
```
You can also run individual steps in `.local/Makefile`.

### Docs
Prerequisites: Ruby, [Setup steps for Jekyll](https://docs.github.com/en/pages/setting-up-a-github-pages-site-with-jekyll/testing-your-github-pages-site-locally-with-jekyll)
```
make build-docs
make run-docs
```