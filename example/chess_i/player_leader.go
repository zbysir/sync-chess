package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
)

type PlayerLeader struct {
	Manager *chess.Manager
}

func (p *PlayerLeader) Banker(players chess.Players) (player chess.Player) {
	return players[0]
}

func (p *PlayerLeader) Next(currPlayer chess.Player, players chess.Players) (player chess.Player) {
	return players.After(currPlayer)
}

func (p *PlayerLeader) PlayerCreator(id string) (player chess.Player) {
	return NewPlayer(id,p.Manager)
}

func NewPlayerLeader() *PlayerLeader {
	return &PlayerLeader{
	}
}
