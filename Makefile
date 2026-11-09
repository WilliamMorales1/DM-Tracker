.PHONY: build build-frontend typecheck-frontend test-frontend build-backend plot run dev clean migrate reseed

build: build-frontend build-backend

build-frontend:
	cd frontend && npm ci && npm run build

typecheck-frontend:
	cd frontend && npm run typecheck

test-frontend:
	cd frontend && npm test

build-backend:
	cd backend && go build -o tracker ./cmd/tracker

MAP ?= error

plot:
	cd backend && go run ./cmd/plot -map $(MAP)

run:
	cd backend && ./tracker

dev: build-frontend
	cd backend && go run ./cmd/tracker

clean:
	rm -f backend/tracker
	rm -rf backend/tmp frontend/dist

migrate:
	cd backend && go run ./cmd/migrate
