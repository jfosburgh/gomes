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
- [x] Browser 
- [x] TUI (terminal user interface)
- [ ] Generic JSON api
<!-- #### TUI Instructions -->
<!-- ```bash -->
<!-- make build tui -->
<!-- make run tui -->
<!-- ``` -->
## The Games
- [x] Tic-Tac-Toe
- [ ] Connect4
- [ ] Checkers
- [x] Chess
- [ ] Wordle
and more to come!
## Technical Details
This project uses Go as the foundation, with [HTMX](https://htmx.org/) for the browser frontend, and (will probably use) [bubbletea](https://github.com/charmbracelet/bubbletea) for the TUI.

## Contributing
If you would like to contribute, fork the repo and make a pull request against main with your changes.
