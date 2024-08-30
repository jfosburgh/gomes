package chess

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	"gopkg.in/freeeve/pgn.v1"
)

var PieceConverter = map[pgn.Piece]int{
	pgn.NoPiece:     EMPTY,
	pgn.BlackPawn:   BLACK | PAWN,
	pgn.BlackRook:   BLACK | ROOK,
	pgn.BlackKnight: BLACK | KNIGHT,
	pgn.BlackBishop: BLACK | BISHOP,
	pgn.BlackQueen:  BLACK | QUEEN,
	pgn.BlackKing:   BLACK | KING,
	pgn.WhitePawn:   WHITE | PAWN,
	pgn.WhiteRook:   WHITE | ROOK,
	pgn.WhiteKnight: WHITE | KNIGHT,
	pgn.WhiteBishop: WHITE | BISHOP,
	pgn.WhiteQueen:  WHITE | QUEEN,
	pgn.WhiteKing:   WHITE | KING,
}

var (
	Codebook map[EBEBoard]map[int][]Move

	PGN_SOURCES = []string{
		"./pkg/chess/Carlsen.pgn",
	}
)

func InitCodebook() {
	Codebook = make(map[EBEBoard]map[int][]Move)

	for _, s := range PGN_SOURCES {
		err := ReadPGNToCodebook(s, 12)
		if err != nil {
			panic(err)
		}
	}
}

func addToCodebook(state EBEBoard, move Move, active int) {
	_, ok := Codebook[state]
	if !ok {
		Codebook[state] = map[int][]Move{}
	}

	for _, existing := range Codebook[state][active] {
		if existing.String() == move.String() {
			return
		}
	}

	Codebook[state][active] = append(Codebook[state][active], move)
}

func ChooseFromCodebook(state EBEBoard, active int) (Move, bool) {
	entries, ok := Codebook[state]
	if !ok {
		return Move{}, false
	}

	activePlayerEntries, ok := entries[active]
	if !ok {
		return Move{}, false
	}

	fmt.Printf("Selecting move from codebook entries: %+v\n", activePlayerEntries)
	return activePlayerEntries[rand.Intn(len(activePlayerEntries))], true
}

func ReadPGNToCodebook(filepath string, moveLimit int) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	ps := pgn.NewPGNScanner(f)
	for ps.Next() {
		g := NewGame()

		pgnGame, err := ps.Scan()
		if err != nil {
			return err
		}

		for i, pgnMove := range pgnGame.Moves {
			if i >= moveLimit {
				break
			}

			fmt.Printf("PGN Move: %d -> %d, promote %d\n", pgnMove.From, pgnMove.To, pgnMove.Promote)
			start := algebraic2Int(pgnMove.From.String())
			end := algebraic2Int(pgnMove.To.String())
			piece := g.EBE.Board[start]

			capture := g.EBE.Board[end]
			if piece&0b0111 == PAWN && start%8 != end%8 {
				offset := 8 - (16 * g.EBE.Active)
				capture = g.EBE.Board[start+offset]
			}

			move := Move{
				Piece: piece,
				Start: start,
				End:   end,

				Capture:   capture,
				Castle:    piece&0b0111 == KING && math.Abs(float64(end)-float64(start)) == 2,
				Promotion: PieceConverter[pgnMove.Promote],

				Halfmoves:       g.EBE.Halfmoves,
				CastlingRights:  g.EBE.CastlingRights,
				EnPassantTarget: g.EBE.EnPassantTarget,
			}

			addToCodebook(g.EBE.Board, move, g.EBE.Active)

			g.MakeMove(move)
		}
	}

	return nil
}
