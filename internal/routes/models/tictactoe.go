package models

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jfosburgh/gomes/internal/routes/utils"
	"github.com/jfosburgh/gomes/pkg/tictactoe"
)

type ModelTTT struct {
	WindowParams
	game         *tictactoe.TicTacToeGame
	data         *utils.TwoPlayerGame
	boardCursorX int
	boardCursorY int
	botTurn      bool
	botChan      chan struct{}
}

type responseMsg struct{}

func waitForActivity(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		fmt.Println("waiting for ping")
		msg := responseMsg(<-sub)
		fmt.Println("received ping")
		return msg
	}
}

func (m ModelTTT) Init() tea.Cmd {
	return nil
}

func (m ModelTTT) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
	case responseMsg:
		fmt.Println("processing bot ping")
		move := m.game.BestMove()
		m.game.MakeMove(move)
		var winner int
		m.data.Ended, winner = m.game.GameOver()

		m.data.Active = utils.TTTPieces[m.game.State.Active]
		m.data.Cells = utils.FillTTTCells(m.game, m.data)

		if m.data.Ended {
			if winner == 0 {
				m.data.Status = "It's a tie!"
			} else {
				m.data.Status = fmt.Sprintf("%s Wins!", utils.TTTPieces[winner])
			}
		} else {
			m.data.Status = fmt.Sprintf("%s's Turn!", m.data.Active)
		}
		m.botTurn = false
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			if !m.data.Ended {
				break
			}

			m.game = tictactoe.NewGame()

			m.data.Ended = false
			m.data.Active = "X"
			m.data.Status = "X goes first!"

			m.data.Cells = utils.FillTTTCells(m.game, m.data)

			m.botTurn = m.data.Active != m.data.Player && m.data.Player != "" && !m.data.Ended
			if m.botTurn {
				go func() {
					time.Sleep(1 * time.Second)
					fmt.Println("sending ping")
					m.botChan <- responseMsg{}
				}()
				return m, tea.Batch(waitForActivity(m.botChan), nil)
			}
		case "up", "k":
			if m.boardCursorY > 0 {
				m.boardCursorY--
			}
		case "down", "j":
			if m.boardCursorY < 2 {
				m.boardCursorY++
			}
		case "left", "h":
			if m.boardCursorX > 0 {
				m.boardCursorX--
			}
		case "right", "l":
			if m.boardCursorX < 3 {
				m.boardCursorX++
			}
		case "enter", " ":
			// switch {
			// case m.botTurn:
			// default:
			move := m.boardCursorY*3 + m.boardCursorX
			if !m.data.Cells[move].Clickable {
				return m, nil
			}

			m.game.MakeMove(move)
			// }

			var winner int
			m.data.Ended, winner = m.game.GameOver()

			m.data.Active = utils.TTTPieces[m.game.State.Active]
			m.data.Cells = utils.FillTTTCells(m.game, m.data)

			if m.data.Ended {
				if winner == 0 {
					m.data.Status = "It's a tie!"
				} else {
					m.data.Status = fmt.Sprintf("%s Wins!", utils.TTTPieces[winner])
				}
			} else {
				m.data.Status = fmt.Sprintf("%s's Turn!", m.data.Active)
			}

			m.botTurn = m.data.Active != m.data.Player && m.data.Player != "" && !m.data.Ended
			if m.botTurn {
				go func() {
					time.Sleep(1 * time.Second)
					fmt.Println("sending ping")
					m.botChan <- responseMsg{}
				}()
				return m, tea.Batch(waitForActivity(m.botChan), nil)
			}
		case "q", "ctrl+c":
			return ModelHome{
				WindowParams: m.WindowParams,
				Games:        []string{"Chess", "Tic-Tac-Toe"},
			}, nil
		}
	}

	return m, nil
}

func (m ModelTTT) View() string {
	t := ""
	cells := m.data.Cells
	cursorIndex := m.boardCursorX + m.boardCursorY*3
	row := ""
	for i, cell := range cells {
		block := lipgloss.Place(2, 1, lipgloss.Center, lipgloss.Center, cell.Content)
		bg := m.TxtStyle.GetForeground()

		if i%2 == 0 {
			bg = m.QuitStyle.GetForeground()
		}

		if i == cursorIndex {
			if cell.Clickable {
				bg = lipgloss.Color("12")
			} else {
				bg = lipgloss.Color("99")
			}
		}

		row = lipgloss.JoinHorizontal(lipgloss.Center, row, lipgloss.NewStyle().Background(bg).Foreground(lipgloss.Color("16")).Inherit(m.TxtStyle).Render(block))

		if i%3 == 2 {
			t = lipgloss.JoinVertical(lipgloss.Center, t, row)
			row = ""
		}
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

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.TxtStyle.Render("Tic-Tac-Toe")+fmt.Sprintf("\n%+s\n", t)+m.QuitStyle.Render(optionText))
}

type ModelTTTSettings struct {
	WindowParams
	page int

	modes      []string
	modeCursor int

	players      []string
	playerCursor int

	difficulties     []int
	difficultyCursor int
}

func (m ModelTTTSettings) Init() tea.Cmd {
	return nil
}

func (m ModelTTTSettings) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				if m.difficultyCursor > 0 {
					m.difficultyCursor--
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
				if m.difficultyCursor < len(m.difficulties)-1 {
					m.difficultyCursor++
				}
			}
		case "enter", " ", "n":
			switch m.page {
			case 0:
				if m.modes[m.modeCursor] == "Player vs. Player" {
					next := ModelTTT{
						WindowParams: m.WindowParams,
						game:         tictactoe.NewGame(),
						data:         &utils.TwoPlayerGame{},
					}

					next.data.Active = "X"

					next.data.Started = true
					next.data.Cells = utils.FillTTTCells(next.game, next.data)

					next.data.Status = "X goes first!"

					return next, nil
				}
				m.page = 1
			case 1:
				m.page = 2
			case 2:
				next := ModelTTT{
					WindowParams: m.WindowParams,
					game:         tictactoe.NewGame(),
					data:         &utils.TwoPlayerGame{},
					botChan:      make(chan struct{}),
				}

				next.data.Active = "X"
				if m.modes[m.modeCursor] == "Player vs. Bot" {
					next.data.Player = m.players[m.playerCursor]
					next.game.SearchDepth = m.difficulties[m.difficultyCursor]
				}

				next.data.Started = true
				next.data.Cells = utils.FillTTTCells(next.game, next.data)

				next.data.Status = "X goes first!"
				next.botTurn = next.data.Active != next.data.Player && next.data.Player != "" && !next.data.Ended

				if next.botTurn {
					go func() {
						time.Sleep(1 * time.Second)
						fmt.Println("sending ping")
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

func (m ModelTTTSettings) View() string {
	s := "Tic-Tac-Toe"
	s += "\n\nChoose your settings:\n"

	s = m.TxtStyle.Render(s)

	modeString := "Game Mode:\n"
	for i, mode := range m.modes {
		cursor := " "

		if i == m.modeCursor {
			if m.page > 0 {
				cursor = "*"
			} else {
				cursor = ">"
			}
		}

		modeString += fmt.Sprintf(" %s %s\n", cursor, mode)
	}

	s = lipgloss.JoinVertical(lipgloss.Left, s, m.TxtStyle.Render(modeString))

	playerString := "Play As:\n"
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

	if m.modes[m.modeCursor] == "Player vs. Bot" {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.TxtStyle.Render(playerString))
	} else {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.QuitStyle.Render(playerString))
	}

	difficultyString := "Search Depth:\n"
	for i, difficulty := range m.difficulties {
		cursor := " "

		if i == m.difficultyCursor {
			if m.page < 2 {
				cursor = " "
			} else if m.page == 2 {
				cursor = ">"
			} else {
				cursor = "*"
			}
		}

		difficultyString += fmt.Sprintf(" %s %d\n", cursor, difficulty)
	}

	if m.modes[m.modeCursor] == "Player vs. Bot" {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.TxtStyle.Render(difficultyString))
	} else {
		s = lipgloss.JoinVertical(lipgloss.Left, s, m.QuitStyle.Render(difficultyString))
	}

	optionText := "Press '<space>', '<enter>', or 'n' to select"
	optionText += "\nPress 'N', 'p' to go back"
	optionText += "\nPress 'q' to go home\n"

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, s+"\n\n"+m.QuitStyle.Render(optionText))
}
