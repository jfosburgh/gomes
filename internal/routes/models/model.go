package models

import "github.com/charmbracelet/lipgloss"

type WindowParams struct {
	Width     int
	Height    int
	TxtStyle  lipgloss.Style
	QuitStyle lipgloss.Style
}
