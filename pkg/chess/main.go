package chess

type ChessGame struct {
	EBE      EBE
	Bitboard BitBoard
	Moves    []Move
	Captured []int
}

func NewGame() *ChessGame {
	c := ChessGame{
		EBE:      DefaultBoard(),
		Bitboard: make(BitBoard),
	}

	c.Bitboard.FromEBE(c.EBE.Board)

	return &c
}

func (c *ChessGame) SetStateFromFEN(fen string) {
	c.EBE.FromFEN(fen)
	c.Bitboard.FromEBE(c.EBE.Board)
	c.Moves = []Move{}
	c.Captured = []int{}
}
