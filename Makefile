.PHONY: build build-frontend build-backend test-ui plot run dev clean migrate reseed

build: build-frontend build-backend

build-frontend:
	test -f frontend/index.html
	test -f frontend/index.css
	test -f frontend/app/main.js
	test -f frontend/vendor/marked.esm.js
	test -f frontend/vendor/monaco/min/vs/loader.js

build-backend:
	cd backend && go build -o tracker ./cmd/tracker

test-ui: build
	cd backend && go test -tags=e2e ./e2e

MAP ?= error

plot:
	cd backend && go run ./cmd/plot -map $(MAP)

run:
	cd backend && ./tracker

dev: build-frontend
	cd backend && go run ./cmd/tracker

clean:
	rm -f backend/tracker
	rm -rf backend/tmp

migrate:
	cd backend && go run ./cmd/migrate
