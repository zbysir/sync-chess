package chess

import "github.com/bysir-zl/sync-chess/core"

type PlayerLeader struct {
}

func (p *PlayerLeader) Banker(players core.Players) (player *core.Player) {
	return players[0]
}

func (p *PlayerLeader) Next(currPlayer *core.Player, players core.Players) (player *core.Player) {
	return players.After(currPlayer)
}

func (p *PlayerLeader) PlayerCreator() (player core.PlayerInterface) {
	return NewPlayer()
}


func NewPlayerLeader() *PlayerLeader {
	return &PlayerLeader{
	}
}
