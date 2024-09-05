package models

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jfosburgh/gomes/internal/routes/utils"
	"github.com/jfosburgh/gomes/pkg/chess"
)

type ModelChess struct {
	WindowParams
	game *chess.ChessGame
	data *utils.TwoPlayerGame

	boardCursorX int
	boardCursorY int

	botTurn bool
	botChan chan struct{}

	promoteCursor int
	promote       bool
	promoteData   []string

	moveSrc int
}

func (m ModelChess) Init() tea.Cmd {
	return nil
}

func (m ModelChess) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
	case responseMsg:
		move := m.game.BestMove()
		m.game.MakeMove(move)
		m.data.Active = utils.ChessNames[m.game.EBE.Active]

		m.data.Cells = utils.FillChessCells(m.game, m.data, -1, false)
		m.data.Ended = len(m.game.GetLegalMoves()) == 0 || m.game.EBE.Halfmoves >= 100
		if m.data.Ended {
			if m.game.EBE.Halfmoves >= 100 {
				m.data.Status = "Enforcing 50-move rule, it's a tie"
			} else {
				m.data.Status = fmt.Sprintf("%s Wins!", utils.ChessNames[^(m.game.EBE.Active)&0b1])
			}
		} else {
			m.data.Status = fmt.Sprintf("%s played %s, %s's Turn!", utils.ChessNames[^m.game.EBE.Active&0b1], move, m.data.Active)
		}

		m.botTurn = m.data.Active != m.data.Player && m.data.Player != "" && !m.data.Ended
		if m.botTurn {
			go func() {
				m.botChan <- responseMsg{}
			}()
			return m, tea.Batch(waitForActivity(m.botChan), nil)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			if !m.data.Ended {
				break
			}

			nextGame := chess.NewGame()
			nextGame.SearchTime = m.game.SearchTime
			m.game = nextGame

			m.data.Ended = false
			m.data.Active = "White"
			m.data.Status = "White goes first!"

			m.data.Cells = utils.FillChessCells(m.game, m.data, -1, false)

			m.promote = false
			m.moveSrc = -1

			m.botTurn = m.data.Active != m.data.Player && m.data.Player != "" && !m.data.Ended
			if m.botTurn {
				go func() {
					m.botChan <- responseMsg{}
				}()
				return m, tea.Batch(waitForActivity(m.botChan), nil)
			}
		case "up", "k":
			switch {
			case m.promote:
			case m.botTurn:
			default:
				if m.boardCursorY > 0 {
					m.boardCursorY--
				}
			}
		case "down", "j":
			switch {
			case m.promote:
			case m.botTurn:
			default:
				if m.boardCursorY < 7 {
					m.boardCursorY++
				}
			}
		case "left", "h":
			switch {
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
			case m.botTurn:
			case m.promote:
				if m.promoteCursor < 3 {
					m.promoteCursor++
				}
			default:
				if m.boardCursorX < 7 {
					m.boardCursorX++
				}
			}
		case "enter", " ":
			// switch {
			// case m.botTurn:
			// default:
			move := m.boardCursorY*8 + m.boardCursorX
			if !m.data.Cells[move].Clickable && !m.promote {
				return m, nil
			}

			switch {
			case m.moveSrc != -1:
				move = utils.FlipRank(move)
				src := utils.FlipRank(m.moveSrc)

				gameMove, valid := m.game.MoveFromLocations(src, move)
				if !valid {
					m.moveSrc = utils.FlipRank(move)
					m.data.Cells = utils.FillChessCells(m.game, m.data, move, false)
					break
				}

				if m.promote {
					fmt.Println("applying promotion")
					m.promote = false
					side := m.game.EBE.Active << 3
					gameMove.Promotion = side | []int{chess.KNIGHT, chess.ROOK, chess.BISHOP, chess.QUEEN}[m.promoteCursor]
				} else if gameMove.Promotion != chess.EMPTY {
					fmt.Println("promoting next choice")
					m.promote = true
					gameMove.Promotion = gameMove.Piece
				}
				m.game.MakeMove(gameMove)
				if !m.promote {
					m.data.Active = utils.ChessNames[m.game.EBE.Active]
				}

				m.data.Cells = utils.FillChessCells(m.game, m.data, -1, m.promote)
				checkmate := len(m.game.GetLegalMoves()) == 0
				draw := !checkmate && m.game.EBE.Halfmoves >= 100
				m.data.Ended = checkmate || draw
				if m.data.Ended {
					m.data.Status = fmt.Sprintf("%s Wins!", utils.ChessNames[^(m.game.EBE.Active)&0b1])
					if draw {
						m.data.Status = "Enforcing 50-move rule, it's a tie!"
					}
				} else {
					m.data.Status = fmt.Sprintf("%s played %s, %s's Turn!", utils.ChessNames[^m.game.EBE.Active&0b1], gameMove, m.data.Active)
				}

				if m.promote {
					m.data.Status = fmt.Sprintf("%s, choose your promotion!", m.data.Active)
					m.game.UnmakeMove(gameMove)
					side := m.game.EBE.Active << 3
					m.promoteData = []string{
						utils.ChessPieces[side|chess.KNIGHT],
						utils.ChessPieces[side|chess.ROOK],
						utils.ChessPieces[side|chess.BISHOP],
						utils.ChessPieces[side|chess.QUEEN],
					}
				} else {
					m.moveSrc = -1
				}
			case m.moveSrc == -1:
				m.moveSrc = move
				location := utils.FlipRank(move)
				m.data.Cells = utils.FillChessCells(m.game, m.data, location, false)
			}
			m.botTurn = m.data.Active != m.data.Player && m.data.Player != "" && !m.data.Ended
			if m.botTurn {
				go func() {
					m.botChan <- responseMsg{}
				}()
				return m, tea.Batch(waitForActivity(m.botChan), nil)
			}
			// }
		case "q", "ctrl+c":
			return ModelHome{
				WindowParams: m.WindowParams,
				Games:        []string{"Chess", "Tic-Tac-Toe"},
			}, nil
		}
	}

	return m, nil
}

func (m ModelChess) View() string {
	cells := m.data.Cells
	cursorIndex := m.boardCursorX + m.boardCursorY*8
	t := ""
	row := ""
	for i, cell := range cells {
		block := lipgloss.Place(2, 1, lipgloss.Center, lipgloss.Center, cell.Content)
		bg := m.TxtStyle.GetForeground()
		switch {
		case strings.Contains(cell.Classes, "target"):
			bg = lipgloss.Color("52")
			break
		case strings.Contains(cell.Classes, "selected"):
			bg = lipgloss.Color("22")
			break
		case strings.Contains(cell.Classes, "black"):
			bg = m.QuitStyle.GetForeground()
			break
		}
		if i == cursorIndex {
			if cell.Clickable {
				bg = lipgloss.Color("12")
			} else {
				bg = lipgloss.Color("99")
			}
		}

		row = lipgloss.JoinHorizontal(lipgloss.Center, row, lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("16")).Inherit(m.TxtStyle).Render(block))

		if i%8 == 7 {
			t = lipgloss.JoinVertical(lipgloss.Center, t, row)
			row = ""
		}
	}

	if m.promote {
		row := ""
		for i, val := range m.promoteData {
			block := lipgloss.Place(2, 1, lipgloss.Center, lipgloss.Center, val)
			bg := m.TxtStyle.GetForeground()

			if i%2 == 0 {
				bg = m.QuitStyle.GetForeground()
			}
			if i == m.promoteCursor {
				bg = lipgloss.Color("12")
			}
			row = lipgloss.JoinHorizontal(lipgloss.Center, row, lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("16")).Inherit(m.TxtStyle).Render(block))
		}
		t += "\n"
		t = lipgloss.JoinVertical(lipgloss.Center, t, row)
	}

	t += "\n"

	status := m.data.Status
	if m.botTurn {
		status += " Bot is thinking..."
	}
	t = lipgloss.JoinVertical(lipgloss.Center, t, m.TxtStyle.Render(status))

	optionText := ""
	if m.data.Ended {
		optionText += "\nPress 'r' to replay"
	}
	optionText += "\nPress 'q' to return home\n"

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.TxtStyle.Render("Chess")+fmt.Sprintf("\n%+s\n", t)+m.QuitStyle.Render(optionText))
}

type ModelChessSettings struct {
	WindowParams
	page int

	modes      []string
	modeCursor int

	players      []string
	playerCursor int

	depths      []int
	depthCursor int

	times      []int
	timeCursor int
}

func (m ModelChessSettings) Init() tea.Cmd {
	return nil
}

func (m ModelChessSettings) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			switch m.page {
			case 0:
				if m.modeCursor > 0 {
					m.modeCursor--
				}
			case 1:
				if m.playerCursor > 0 {
					m.playerCursor--
				}
			case 2:
				if m.depthCursor > 0 {
					m.depthCursor--
				}
			case 3:
				if m.timeCursor > 0 {
					m.timeCursor--
				}
			}
		case "down", "j":
			switch m.page {
			case 0:
				if m.modeCursor < len(m.modes)-1 {
					m.modeCursor++
				}
			case 1:
				if m.playerCursor < len(m.players)-1 {
					m.playerCursor++
				}
			case 2:
				if m.depthCursor < len(m.depths)-1 {
					m.depthCursor++
				}
			case 3:
				if m.timeCursor < len(m.times)-1 {
					m.timeCursor++
				}
			}
		case "enter", " ", "n":
			switch m.page {
			case 0:
				if m.modes[m.modeCursor] == "Player vs. Player" {
					next := ModelChess{
						WindowParams: m.WindowParams,
						game:         chess.NewGame(),
						data:         &utils.TwoPlayerGame{},

						moveSrc: -1,
					}

					next.data.Active = "White"

					next.data.Started = true
					next.data.Cells = utils.FillChessCells(next.game, next.data, -1, false)

					next.data.Status = "White goes first!"

					return next, nil
				}
				m.page = 1
			case 1:
				m.page = 2
			case 2:
				m.page = 3
			case 3:
				next := ModelChess{
					WindowParams: m.WindowParams,
					game:         chess.NewGame(),
					data:         &utils.TwoPlayerGame{},

					moveSrc: -1,
					botChan: make(chan struct{}),
				}

				next.data.Active = "White"
				next.game.MaxSearchDepth = m.depths[m.depthCursor]
				next.game.SearchTime = time.Duration(m.times[m.timeCursor]) * time.Second

				if m.modes[m.modeCursor] == "Player vs. Bot" {
					next.data.Player = m.players[m.playerCursor]
				} else {
					next.data.Player = "neither"
				}

				next.data.Started = true
				next.data.Cells = utils.FillChessCells(next.game, next.data, -1, false)

				next.data.Status = "White goes first!"
				next.botTurn = next.data.Active != next.data.Player && next.data.Player != "" && !next.data.Ended

				if next.botTurn {
					go func() {
						next.botChan <- responseMsg{}
					}()
				}

				return next, waitForActivity(next.botChan)
			}
		case "N", "p":
			if m.page == 0 {
				return ModelHome{
					WindowParams: m.WindowParams,
					Games:        []string{"Chess", "Tic-Tac-Toe"},
				}, nil
			}
			m.page = max(0, m.page-1)
		case "q", "ctrl+c":
			return ModelHome{
				WindowParams: m.WindowParams,
				Games:        []string{"Chess", "Tic-Tac-Toe"},
			}, nil
		}
	}

	return m, nil
}

func (m ModelChessSettings) View() string {
	s := "Chess"
	s += "\n\nChoose your settings:\n"

	s = m.TxtStyle.Render(s)

	modeString := "\nGame Mode:\n"
	for i, mode := range m.modes {
		cursor := " "

		if i == m.modeCursor && m.page >= 0 {
			if m.page > 0 {
				cursor = "*"
			} else {
				cursor = ">"
			}
		}

		modeString += fmt.Sprintf(" %s %s\n", cursor, mode)
	}

	s = lipgloss.JoinVertical(lipgloss.Left, s, m.TxtStyle.Render(modeString))

	playerString := "\nPlay As:\n"
	for i, player := range m.players {
		cursor := " "

		if i == m.playerCursor {
			if m.page < 1 {
				cursor = " "
			} else if m.page == 1 {
				cursor = ">"
			} else {
				cursor = "*"
			}
		}

		playerString += fmt.Sprintf(" %s %s\n", cursor, player)
	}

	if m.modes[m.modeCursor] != "Player vs. Player" {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.TxtStyle.Render(playerString))
	} else {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.QuitStyle.Render(playerString))
	}

	depthString := "\nSearch Depth:\n"
	for i, depth := range m.depths {
		cursor := " "

		if i == m.depthCursor {
			if m.page < 2 {
				cursor = " "
			} else if m.page == 2 {
				cursor = ">"
			} else {
				cursor = "*"
			}
		}

		depthString += fmt.Sprintf(" %s %d ply\n", cursor, depth)
	}

	if m.modes[m.modeCursor] != "Player vs. Player" {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.TxtStyle.Render(depthString))
	} else {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.QuitStyle.Render(depthString))
	}

	timeString := "\nSearch Time:\n"
	for i, seconds := range m.times {
		cursor := " "

		if i == m.timeCursor {
			if m.page < 3 {
				cursor = " "
			} else if m.page == 3 {
				cursor = ">"
			} else {
				cursor = "*"
			}
		}

		timeString += fmt.Sprintf(" %s %ds\n", cursor, seconds)
	}

	if m.modes[m.modeCursor] != "Player vs. Player" {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.TxtStyle.Render(timeString))
	} else {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.QuitStyle.Render(timeString))
	}

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, s+"\n\n"+m.QuitStyle.Render("Press 'q' to go home\n"))
}
