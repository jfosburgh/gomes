package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModelHome struct {
	cursor int
	Games  []string
	WindowParams
}

func (m ModelHome) Init() tea.Cmd {
	return nil
}

func (m ModelHome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.Width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.Games)-1 {
				m.cursor++
			}
		case "enter", " ":
			switch m.Games[m.cursor] {
			case "Tic-Tac-Toe":
				next := ModelTTTSettings{
					WindowParams: m.WindowParams,
					modes: []string{
						"Player vs. Player",
						"Player vs. Bot",
					},
					players: []string{
						"X",
						"O",
					},
					difficulties: []int{
						1,
						2,
						3,
						4,
						5,
						6,
						7,
						8,
						9,
					},
				}
				return next, nil
			case "Chess":
				next := ModelChessSettings{
					WindowParams: m.WindowParams,
					modes: []string{
						"Player vs. Player",
						"Player vs. Bot",
						"Bot vs. Bot",
					},
					players: []string{
						"White",
						"Black",
					},
					depths: []int{
						1,
						2,
						3,
						4,
						5,
						6,
					},
					times: []int{
						1,
						2,
						3,
						4,
						5,
						6,
					},
				}
				return next, nil
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ModelHome) View() string {
	s := fmt.Sprintf("Welcome to gomes.sh, the best place to play games in the terminal")
	s += "\n\nWhat would you like to play?\n"

	for i, game := range m.Games {
		cursor := " "

		if i == m.cursor {
			cursor = ">"
		}

		s += fmt.Sprintf(" %s %s\n", cursor, game)
	}

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.TxtStyle.Render(s)+"\n\n"+m.QuitStyle.Render("Press 'q' to quit\n"))
}
