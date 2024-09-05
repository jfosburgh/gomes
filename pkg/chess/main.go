package chess

import (
	"fmt"
	"sync"
	"time"
)

type ChessGame struct {
	EBE                EBE
	Bitboard           *BitBoard
	Moves              []Move
	Captured           []int
	MaxSearchDepth     int
	Transpositions     map[EBEBoard]TranspositionNode
	TranspositionMutex *sync.RWMutex
	SearchStart        time.Time
	SearchTime         time.Duration
	SearchTimer        *time.Timer
}

type TranspositionNode struct {
	Value float64
	Depth int
}

func NewGame() *ChessGame {
	c := ChessGame{
		EBE:                DefaultBoard(),
		Bitboard:           &BitBoard{},
		MaxSearchDepth:     4,
		SearchTime:         2,
		TranspositionMutex: &sync.RWMutex{},
	}

	c.Bitboard.FromEBE(c.EBE.Board)

	return &c
}

func (c *ChessGame) Clone() *ChessGame {
	clone := NewGame()

	copyBitboard(c.Bitboard, clone.Bitboard)
	clone.EBE.Board = copyBoard(c.EBE.Board)
	clone.EBE.Active = c.EBE.Active
	clone.EBE.CastlingRights = c.EBE.CastlingRights
	clone.EBE.EnPassantTarget = c.EBE.EnPassantTarget
	clone.EBE.Halfmoves = c.EBE.Halfmoves
	clone.EBE.Moves = c.EBE.Moves
	clone.Moves = append(clone.Moves, c.Moves...)

	clone.Transpositions = c.Transpositions
	clone.SearchTimer = c.SearchTimer
	clone.TranspositionMutex = c.TranspositionMutex

	return clone
}

func (c *ChessGame) SetStateFromFEN(fen string) {
	c.EBE.FromFEN(fen)
	c.Bitboard.FromEBE(c.EBE.Board)
	c.Moves = []Move{}
	c.Captured = []int{}
}

func copyBitboard(source, dest *BitBoard) {
	for piece := range source {
		dest[piece] = source[piece]
	}
}

func copyBoard(source EBEBoard) EBEBoard {
	board := EBEBoard{}
	for i := range 64 {
		board[i] = source[i]
	}

	return board
}

func (c *ChessGame) BestMove() Move {
	// codebookMove, ok := ChooseFromCodebook(c.EBE.Board, c.EBE.Active)
	// if ok {
	// 	fmt.Printf("selected move from codebook: %+v\n", codebookMove)
	// 	return codebookMove
	// }
	//
	// fmt.Printf("board not in codebook, searching\n")

	options, _ := c.Search()
	if c.EBE.Active<<3 == WHITE {
		return options[len(options)-1]
	}

	return options[0]
}

func (c *ChessGame) MoveFromLocations(start, end int) (Move, bool) {
	pseudoLegal := c.GeneratePseudoLegal()
	active := c.EBE.Active << 3
	for _, move := range pseudoLegal {
		if move.Start != start || move.End != end {
			continue
		}

		legal := false
		c.MakeMove(move)
		if !c.Bitboard.InCheck(active) {
			legal = true
		}
		c.UnmakeMove(move)

		return move, legal
	}

	return Move{}, false
}

func (c *ChessGame) GetLegalMoves() []Move {
	pseudoLegal := c.GeneratePseudoLegal()

	moves := []Move{}
	active := c.EBE.Active << 3
	for _, move := range pseudoLegal {
		c.MakeMove(move)
		if !c.Bitboard.InCheck(active) {
			moves = append(moves, move)
		}
		c.UnmakeMove(move)
	}

	return moves
}

func (c *ChessGame) GetMoveTargets(pieceLocation int) []int {
	fmt.Printf("getting moves for %d\n", pieceLocation)
	pseudoLegal := c.GeneratePseudoLegal()
	fmt.Printf("found %d pseudolegal moves\n", len(pseudoLegal))

	moves := []int{}
	active := c.EBE.Active << 3
	for _, move := range pseudoLegal {
		if move.Start != pieceLocation {
			fmt.Printf("discarding %+v, %d->%d\n", move, move.Start, move.End)
			continue
		}

		c.MakeMove(move)
		if !c.Bitboard.InCheck(active) {
			moves = append(moves, move.End)
		}
		c.UnmakeMove(move)
	}

	return moves
}

func (c *ChessGame) Perft(depth, startDepth int, debug bool) (int, string) {
	if depth == 0 {
		return 1, ""
	}

	start := time.Time{}
	wg := sync.WaitGroup{}
	if depth == startDepth {
		start = time.Now()
	}

	resultString := ""

	count := 0
	moves := c.GeneratePseudoLegal()
	res := make(chan int, len(moves))

	for _, move := range moves {
		if depth >= 5 || depth == startDepth {
			wg.Add(1)
			go func() {
				defer wg.Done()
				clone := c.Clone()
				active := clone.EBE.Active << 3
				childMoveCount := 0

				clone.MakeMove(move)
				if !clone.Bitboard.InCheck(active) {
					childMoveCount, _ = clone.Perft(depth-1, startDepth, debug)
				}
				clone.UnmakeMove(move)

				res <- childMoveCount
				if depth == startDepth && childMoveCount != 0 {
					resultString += fmt.Sprintf("%s: %d\n", move, childMoveCount)
				}
			}()
		} else {
			active := c.EBE.Active << 3
			childMoveCount := 0

			c.MakeMove(move)
			if !c.Bitboard.InCheck(active) {
				childMoveCount, _ = c.Perft(depth-1, startDepth, debug)
			}
			c.UnmakeMove(move)

			res <- childMoveCount
			if depth == startDepth && childMoveCount != 0 {
				resultString += fmt.Sprintf("%s: %d\n", move, childMoveCount)
			}
		}
	}

	for range moves {
		count += <-res
	}

	if depth == startDepth {
		runTime := time.Since(start)
		fmt.Printf("perft evaluated to depth of %d in %05dms, %08d moves/second\n", startDepth, runTime.Milliseconds(), int(float32(count)/float32(runTime.Microseconds())*1e6))
	}

	return count, resultString
}
