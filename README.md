# level_zero
**To run:**

Go to level_zero directory:<br>
`cd path/to/level_zero`

Start container with PostgreSQL and a NATS:<br>
`docker-compose -f deploy/docker-compose.yml up`

Run server:<br>
`go run cmd/server/server.go`

Run generator:<br>
`go run cmd/generator/generator.go`