package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
)

type PlayerLeader struct {
}

func (p *PlayerLeader) Banker(players chess.Players) (player *chess.Player) {
	return players[0]
}

func (p *PlayerLeader) Next(currPlayer *chess.Player, players chess.Players) (player *chess.Player) {
	return players.After(currPlayer)
}

func (p *PlayerLeader) PlayerCreator() (player chess.PlayerInterface) {
	return NewPlayer()
}

func NewPlayerLeader() *PlayerLeader {
	return &PlayerLeader{
	}
}
