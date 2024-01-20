# Container scripts
test:
	$(MAKE) -C ../ build
	cd cypress && docker compose build
	cd cypress && docker compose up --exit-code-from blinkfile-cypress
.PHONY: test

# Host machine scripts
install:
	cd cypress && npm i
.PHONY: install

open: install
	cd cypress && npm run test:open
.PHONY: open
