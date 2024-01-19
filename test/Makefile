# Container scripts
test:
	$(MAKE) -C ../ build
	cd cypress && docker compose build
	cd cypress && docker compose up --exit-code-from blinkfile-cypress

# Host machine scripts
install:
	cd cypress && npm i

open: install
	cd cypress && npm run test:open
