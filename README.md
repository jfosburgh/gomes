# gomes
A small collection of simple games written in Go.
## How to Play
### Clone the Repo
```bash
git clone https://github.com/jfosburgh/gomes
cd gomes
```
### Setup the Server
First, fetch dependencies:
```bash
go mod tidy
```
Then run the server:
```bash
go build -o ./bin/ ./cmd/server
./bin/server
```
or
```bash
go run ./cmd/server
```
This will by default host the web server at `http://localhost:8080` and an ssh server for TUI connections on port `23234`
To change the web port, create a `.env` file in the `gomes` directory, and add the following:
```
SERVER_PORT=<port number>
```
### Connect to the Server
Connect to the server and start playing!
Currently supported clients are:
- [x] Browser 
- [x] TUI (terminal user interface)

To connect via TUI, enter the following in a terminal:
```bash
ssh localhost -p 23234
```
## The Games
- [x] Tic-Tac-Toe
- [x] Chess
## Technical Details
This project uses Go as the foundation, with [HTMX](https://htmx.org/) for the browser frontend, and [Wish](https://github.com/charmbracelet/wish) to provide the ssh server functionality with [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI.

## Contributing
If you would like to contribute, fork the repo and make a pull request against main with your changes.
