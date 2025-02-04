all: client server

client: ./frontend/node_modules
	cd frontend && npm run build

server:
	go build -o server .

test:
	go test ./database
	go test ./handlers

clean:
	rm -f server
	rm -f brainwave_db.db
	rm -rf frontend/node_modules
	rm -rf frontend/dist

./frontend/node_modules:
	cd frontend && npm ci

.PHONY: client server test clean
