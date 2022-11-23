# To start the servers and clients
1. To start a server type `go run server/server.go -port <port> -name <name>`
2. If you haven't changed the file server-addresses.txt you need to run 3 servers with the ports: 5000, 5001, and 5002 (each with their unique name)
3. Now open as many clients as you want by writing `go run client/client.go -id <id>` each with its own unique id (and in a seperate terminal)

# Commands in client terminal:
To bid in client type: `bid <amount>`
EXAMPLE: `bid 12`

To see status of auction type: `result`

To start a new auction type: `start`
