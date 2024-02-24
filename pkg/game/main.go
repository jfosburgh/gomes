package game

type Handler interface {
	Info() (string, string)
	NewGame() (Game, string)
	GameOptions() GameOptions
}

type Game interface {
	TemplateData() (string, interface{})
	ProcessTurn(string) error
	Start(GameOptions)
}

type SelectOption struct {
	Value   string
	Content string
}

type GameOptions struct {
	Modes              []SelectOption
	Difficulties       []SelectOption
	SelectedMode       string
	SelectedDifficulty string
	Name               string
	Bot                bool
	PlayerID           string
}
