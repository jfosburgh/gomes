package routes

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/elapsed"
	"github.com/charmbracelet/wish/logging"
	"github.com/jfosburgh/gomes/pkg/chess"
	"github.com/jfosburgh/gomes/pkg/tictactoe"
)

const (
	host = "localhost"
	port = "23234"
)

var sshGames map[string]interface{}
var sshGameData map[string]*twoplayergame

func ServeSSH() {
	sshGames = make(map[string]interface{})
	sshGameData = make(map[string]*twoplayergame)

	srv, err := wish.NewServer(
		// The address the server will listen to.
		wish.WithAddress(net.JoinHostPort(host, port)),

		// The SSH server need its own keys, this will create a keypair in the
		// given path if it doesn't exist yet.
		// By default, it will create an ED25519 key.
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		// Middlewares do something on a ssh.Session, and then call the next
		// middleware in the stack.
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.

			// The last item in the chain is the first to be called.
			logging.Middleware(),
			elapsed.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			// We ignore ErrServerClosed because it is expected.
			log.Error("Could not start server", "error", err)
		}
	}()
}

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// tea.WithAltScreen) on a session by session basis.
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	// This should never fail, as we are using the activeterm middleware.
	pty, _, _ := s.Pty()

	// When running a Bubble Tea app over SSH, you shouldn't use the default
	// lipgloss.NewStyle function.
	// That function will use the color profile from the os.Stdin, which is the
	// server, not the client.
	// We provide a MakeRenderer function in the bubbletea middleware package,
	// so you can easily get the correct renderer for the current session, and
	// use it to create the styles.
	// The recommended way to use these styles is to then pass them down to
	// your Bubble Tea model.
	renderer := bubbletea.MakeRenderer(s)
	txtStyle := renderer.NewStyle().Foreground(lipgloss.Color("10"))
	quitStyle := renderer.NewStyle().Foreground(lipgloss.Color("8"))

	m := model{
		width:     pty.Window.Width,
		height:    pty.Window.Height,
		txtStyle:  txtStyle,
		quitStyle: quitStyle,
		games: []string{
			"Chess",
			"Tic-Tac-Toe",
		},
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

// Just a generic tea.Model to demo terminal information of ssh.
type model struct {
	width     int
	height    int
	txtStyle  lipgloss.Style
	quitStyle lipgloss.Style

	cursor int
	games  []string
	gameID string

	boardSize    int
	boardCursorX int
	boardCursorY int

	moveSrc int

	promoteCursor int
	promote       bool
	promoteData   []string

	botTurn bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		// The "up" and "k" keys move the cursor up
		case "up", "k":
			switch {
			case m.gameID == "":
				if m.cursor > 0 {
					m.cursor--
				}
			case m.promote:
			case m.botTurn:
			default:
				if m.boardCursorY > 0 {
					m.boardCursorY--
				}
			}
		// The "down" and "j" keys move the cursor down
		case "down", "j":
			switch {
			case m.gameID == "":
				if m.cursor < len(m.games)-1 {
					m.cursor++
				}
			case m.promote:
			case m.botTurn:
			default:
				if m.boardCursorY < m.boardSize-1 {
					m.boardCursorY++
				}
			}
		case "left", "h":
			switch {
			case m.gameID == "":
			case m.botTurn:
			case m.promote:
				if m.promoteCursor > 0 {
					m.promoteCursor--
				}
			default:
				if m.boardCursorX > 0 {
					m.boardCursorX--
				}
			}
		case "right", "l":
			switch {
			case m.gameID == "":
			case m.botTurn:
			case m.promote:
				if m.promoteCursor < 3 {
					m.promoteCursor++
				}
			default:
				if m.boardCursorX < m.boardSize-1 {
					m.boardCursorX++
				}
			}
		case "enter", " ":
			fmt.Println(m.botTurn)
			switch {
			case m.gameID == "":
				game := twoplayergame{}
				game.ID = generateID()
				game.Player = ""

				game.Started = true
				game.Ended = false

				m.gameID = game.ID

				switch m.games[m.cursor] {
				case "Tic-Tac-Toe":
					ttt := tictactoe.NewGame()
					sshGames[m.gameID] = ttt

					game.Active = "X"
					game.Player = "X"

					game.State = ttt.ToGameString()
					game.Cells = fillTTTCells(ttt, &game)

					m.boardSize = 3

					game.Status = "X goes first!"
				case "Chess":
					chessGame := chess.NewGame()
					chessGame.MaxSearchDepth = 12
					chessGame.SearchTime = time.Second * 1
					sshGames[game.ID] = chessGame

					game.Active = "White"
					game.Player = "White"

					game.State = chessGame.EBE.ToFEN()
					game.Cells = fillChessCells(chessGame, &game, -1, false)

					m.boardSize = 8
					game.Status = "White goes first!"

					m.moveSrc = -1
				}

				sshGameData[game.ID] = &game
				m.gameID = game.ID
			case m.botTurn:
				data := sshGameData[m.gameID]
				gameInterface := sshGames[m.gameID]

				switch gameInterface.(type) {
				case *tictactoe.TicTacToeGame:
					game := gameInterface.(*tictactoe.TicTacToeGame)
					move := game.BestMove()
					game.MakeMove(move)

					var winner int
					data.Ended, winner = game.GameOver()

					data.Active = tttPieces[game.State.Active]
					data.Cells = fillTTTCells(game, data)

					if data.Ended {
						if winner == 0 {
							data.Status = "It's a tie!"
						} else {
							data.Status = fmt.Sprintf("%s Wins!", tttPieces[winner])
						}
					} else {
						data.Status = fmt.Sprintf("%s's Turn!", data.Active)
					}
				case *chess.ChessGame:
					game := gameInterface.(*chess.ChessGame)
					move := game.BestMove()
					game.MakeMove(move)
					data.Active = chessNames[game.EBE.Active]

					data.Cells = fillChessCells(game, data, -1, false)
					data.Ended = len(game.GetLegalMoves()) == 0
					if data.Ended {
						data.Status = fmt.Sprintf("%s Wins!", chessNames[^(game.EBE.Active)&0b1])
					} else {
						data.Status = fmt.Sprintf("%s played %s, %s's Turn!", chessNames[^game.EBE.Active&0b1], move, data.Active)
					}
				}
				m.botTurn = false
			case !sshGameData[m.gameID].Ended:
				data := sshGameData[m.gameID]
				move := m.boardCursorY*m.boardSize + m.boardCursorX
				if sshGameData[m.gameID].Cells[move].Clickable || m.promote {
					gameInterface := sshGames[m.gameID]

					switch gameInterface.(type) {
					case *tictactoe.TicTacToeGame:
						game := gameInterface.(*tictactoe.TicTacToeGame)
						game.MakeMove(move)

						var winner int
						data.Ended, winner = game.GameOver()

						data.Active = tttPieces[game.State.Active]
						data.Cells = fillTTTCells(game, data)

						if data.Ended {
							if winner == 0 {
								data.Status = "It's a tie!"
							} else {
								data.Status = fmt.Sprintf("%s Wins!", tttPieces[winner])
							}
						} else {
							data.Status = fmt.Sprintf("%s's Turn!", data.Active)
						}
					case *chess.ChessGame:
						fmt.Printf("src: %d, target: %d, promote: %+v\n", m.moveSrc, move, m.promote)
						switch {
						case m.moveSrc != -1:
							game := gameInterface.(*chess.ChessGame)

							move = flipRank(move)
							src := flipRank(m.moveSrc)

							gameMove, valid := game.MoveFromLocations(src, move)
							if !valid {
								m.moveSrc = move
								game := gameInterface.(*chess.ChessGame)
								data.Cells = fillChessCells(game, data, move, false)
								break
							}

							if m.promote {
								fmt.Println("applying promotion")
								m.promote = false
								side := game.EBE.Active << 3
								gameMove.Promotion = side | []int{chess.KNIGHT, chess.ROOK, chess.BISHOP, chess.QUEEN}[m.promoteCursor]
							} else if gameMove.Promotion != chess.EMPTY {
								fmt.Println("promoting next choice")
								m.promote = true
								gameMove.Promotion = gameMove.Piece
							}
							game.MakeMove(gameMove)
							if !m.promote {
								data.Active = chessNames[game.EBE.Active]
							}

							data.Cells = fillChessCells(game, data, -1, m.promote)
							checkmate := len(game.GetLegalMoves()) == 0
							draw := !checkmate && game.EBE.Halfmoves >= 100
							data.Ended = checkmate || draw
							if data.Ended {
								data.Status = fmt.Sprintf("%s Wins!", chessNames[^(game.EBE.Active)&0b1])
								if draw {
									data.Status = "It's a tie!"
								}
							} else {
								data.Status = fmt.Sprintf("%s played %s, %s's Turn!", chessNames[^game.EBE.Active&0b1], gameMove, data.Active)
							}

							if m.promote {
								data.Status = fmt.Sprintf("%s, choose your promotion!", data.Active)
								game.UnmakeMove(gameMove)
								side := game.EBE.Active << 3
								m.promoteData = []string{
									chessPieces[side|chess.KNIGHT],
									chessPieces[side|chess.ROOK],
									chessPieces[side|chess.BISHOP],
									chessPieces[side|chess.QUEEN],
								}
							} else {
								m.moveSrc = -1
							}
						case m.moveSrc == -1:
							m.moveSrc = move
							location := flipRank(move)
							game := gameInterface.(*chess.ChessGame)
							data.Cells = fillChessCells(game, data, location, false)
						}
					default:
						fmt.Println("didn't type properly")
						fmt.Println(gameInterface)
						fmt.Println(sshGames)
					}
				}
				m.botTurn = data.Active != data.Player && data.Player != "" && !data.Ended
			}
		case "q", "ctrl+c":
			switch m.gameID {
			case "":
				return m, tea.Quit
			}

			delete(sshGameData, m.gameID)
			delete(sshGames, m.gameID)

			newM := model{}
			newM.width = m.width
			newM.height = m.height
			newM.txtStyle = m.txtStyle
			newM.quitStyle = m.quitStyle
			newM.games = m.games

			return newM, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.gameID == "" {
		s := fmt.Sprintf("Welcome to gomes.sh, the best place to play games in the terminal")
		s += "\n\nWhat would you like to play?\n"

		for i, game := range m.games {
			cursor := " "

			if i == m.cursor {
				cursor = ">"
			}

			s += fmt.Sprintf(" %s %s\n", cursor, game)
		}

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.txtStyle.Render(s)+"\n\n"+m.quitStyle.Render("Press 'q' to quit\n"))
	} else {
		s := m.games[m.cursor]
		data := sshGameData[m.gameID]

		var t string
		switch s {
		case "Tic-Tac-Toe":
			t = ""
			cells := data.Cells
			cursorIndex := m.boardCursorX + m.boardCursorY*3
			row := ""
			for i, cell := range cells {
				block := lipgloss.Place(2, 1, lipgloss.Center, lipgloss.Center, cell.Content)
				bg := m.txtStyle.GetForeground()

				if i%2 == 0 {
					bg = m.quitStyle.GetForeground()
				}

				if i == cursorIndex {
					if cell.Clickable {
						bg = lipgloss.Color("12")
					} else {
						bg = lipgloss.Color("99")
					}
				}

				row = lipgloss.JoinHorizontal(lipgloss.Center, row, lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("16")).Inherit(m.txtStyle).Render(block))

				if i%3 == 2 {
					t = lipgloss.JoinVertical(lipgloss.Center, t, row)
					row = ""
				}
			}

			t += "\n"

			status := sshGameData[m.gameID].Status
			if m.botTurn {
				status += " [<enter>/<space>] to trigger bot's turn"
			}
			t = lipgloss.JoinVertical(lipgloss.Center, t, m.txtStyle.Render(status))
		case "Chess":
			cells := data.Cells
			cursorIndex := m.boardCursorX + m.boardCursorY*8
			t = ""
			row := ""
			for i, cell := range cells {
				block := lipgloss.Place(2, 1, lipgloss.Center, lipgloss.Center, cell.Content)
				bg := m.txtStyle.GetForeground()
				switch {
				case strings.Contains(cell.Classes, "target"):
					bg = lipgloss.Color("52")
					break
				case strings.Contains(cell.Classes, "selected"):
					bg = lipgloss.Color("22")
					break
				case strings.Contains(cell.Classes, "black"):
					bg = m.quitStyle.GetForeground()
					break
				}
				if i == cursorIndex {
					if cell.Clickable {
						bg = lipgloss.Color("12")
					} else {
						bg = lipgloss.Color("99")
					}
				}

				row = lipgloss.JoinHorizontal(lipgloss.Center, row, lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("16")).Inherit(m.txtStyle).Render(block))

				if i%8 == 7 {
					t = lipgloss.JoinVertical(lipgloss.Center, t, row)
					row = ""
				}
			}

			if m.promote {
				row := ""
				for i, val := range m.promoteData {
					block := lipgloss.Place(2, 1, lipgloss.Center, lipgloss.Center, val)
					bg := m.txtStyle.GetForeground()

					if i%2 == 0 {
						bg = m.quitStyle.GetForeground()
					}
					if i == m.promoteCursor {
						bg = lipgloss.Color("12")
					}
					row = lipgloss.JoinHorizontal(lipgloss.Center, row, lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("16")).Inherit(m.txtStyle).Render(block))
				}
				t += "\n"
				t = lipgloss.JoinVertical(lipgloss.Center, t, row)
			}

			t += "\n"

			status := sshGameData[m.gameID].Status
			if m.botTurn {
				status += " [<enter>/<space>] to trigger bot's turn"
			}
			t = lipgloss.JoinVertical(lipgloss.Center, t, m.txtStyle.Render(status))
		}

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.txtStyle.Render(s)+fmt.Sprintf("\n%+s\n", t)+m.quitStyle.Render("\nPress 'q' to return home\n"))
	}
}
