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

## Build & run in containers
Prerequisites: Docker, Node, NPM
```
npm i
npm start
```
Starts the Blinkfile app at [http://localhost:8020](http://localhost:8020) and the documentation at [http://localhost:8021](http://localhost:8021)  

## Install & run on host
Requires Go  
Required environment variables:
- ADMIN_USERNAME
- ADMIN_PASSWORD
```
npm run host:install
go run ./...
```

## Lint
### Run linters
```
npm run lint
```

## Test
### Run the full test suite
```
npm test
```

### Run unit tests with coverage
```
npm run test:unit
```
Open `coverage.html` in a browser to view the coverage report.

### Run acceptance tests
This also builds & runs the app in a containerized environment
```
npm run test:acceptance
```

### Open Cypress UI to run and develop acceptance tests
```
npm run cypress:open
```
This requires the app to be running locally on port 8020 with ENABLE_TEST_AUTOMATION=true, ADMIN_USERNAME=admin, ADMIN_PASSWORD=1234123412341234 which is what is run by default with `npm start`. The Cypress UI will open and you can select the feature file you want to run. 

Gherkin-style features are defined in `test/cypress/features`. Any scenarios tagged with `@pending` or `@implementing`
will _not_ be run in the CI pipeline. Scenarios tagged with `@implementing` will show up in the test runner locally.

Cypress implementation steps are defined in `test/cypress/steps`. See
[badeball's cypress-cucumber-preprocessor docs](https://github.com/badeball/cypress-cucumber-preprocessor/blob/master/docs/readme.md)
for more details.

All new features should have a minimal set of acceptance tests to cover the happy path and important edge cases.

## Documentation
Built with [Hugo](https://gohugo.io/) in the /docs directory using McShelby's [relearn theme](https://github.com/McShelby/hugo-theme-relearn).

### Run Hugo dev server
```
npm run docs
```
Starts the server on [http://localhost:8021](http://localhost:8021) and watches for any changes. You can pass arguments to the [Hugo CLI](https://gohugo.io/commands/hugo/) through NPM after the `--` without needing to install it locally, like so:
```
npm run hugo -- version
npm run hugo -- --help
```

## Conventional Commits
This project uses [Conventional Commits](https://www.conventionalcommits.org) with the [commitlint conventional presets](https://github.com/conventional-changelog/commitlint).

After staging your changes in Git, you can use `npm run commit`. This will prompt you to fill in the commit message fields and then generate the commit message for you.

## Git Hooks
This project uses [Husky](https://typicode.github.io/husky/) to run Git hooks, defined in the `.husky/` directory. This is recommended to reduce the chance of pipeline failures.

To install the hooks, run:
```
npm i
npm run prepare
```
To skip the hooks for a specific git command, use the `-n/--no-verify` flag, for instance:
```
git commit -m "..." -n
```
Or for multiple commands set the env var HUSKY=0:
```
export HUSKY=0 # Disables all Git hooks
git ...
git ...
unset HUSKY # Re-enables hooks
```

To remove them, run:
```
git config --unset core.hooksPath
```
