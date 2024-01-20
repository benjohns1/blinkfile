# Container scripts
ci-test:
	$(MAKE) -C ../ build
	cd cypress && docker compose build
	cd cypress && docker compose up --exit-code-from blinkfile-cypress
.PHONY: test

# Host machine scripts
install:
	cd cypress && npm i
.PHONY: install

test: install
	cd cypress && npm run test
.PHONY: test

open: install
	cd cypress && npm run test:open
.PHONY: open
