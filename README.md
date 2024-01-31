# gomes
A collection of simple games written in Go.
## How to Play
### Clone the Repo
```bash
git clone https://github.com/jfosburgh/gomes
cd gomes
```
### Setup the Server
```bash
make build server
make run server
```
This will by default host the server at `http://localhost:8080`
To change the port, create a `.env` file in the `gomes` directory, and add the following:
```
SERVER_PORT=<port number>
```
### Connect to the Server
Connect to the server and start playing!
Currently supported clients are:
- [ ] Browser 
- [ ] TUI (terminal user interface)
- [ ] Generic JSON api
#### TUI Instructions
```bash
make build tui
make run tui
```
## The Games
- [ ] Tic-Tac-Toe
- [ ] Connect4
- [ ] Checkers
- [ ] Chess
- [ ] Wordle
and more to come!
## Technical Details
This project uses Go as the foundation, with [HTMX](https://htmx.org/) and [templ](https://templ.guide/) for the browser frontend, [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI, and [chi](https://go-chi.io/#/) (probably, still tbd) to handle the api calls. 
