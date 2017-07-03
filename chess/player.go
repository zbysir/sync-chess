package chess

import "github.com/bysir-zl/sync-chess/core"

type Player struct {
	Name   string
	Reader chan *core.PlayerAction
}

func (p *Player) Peng(player core.Player, card core.Card) (err error) {
	panic("implement me")
}

func (p *Player) GangDian(player core.Player, card core.Card) {
	panic("implement me")
}

func (p *Player) GangBu(card core.Card) {
	panic("implement me")
}

func (p *Player) GangZi(card core.Card) {
	panic("implement me")
}

func (p *Player) HuZiMo(card core.Card) {
	panic("implement me")
}

func (p *Player) HuDian(player core.Player, c core.Card) {
	panic("implement me")
}

func (p *Player) HuQiangGang(player core.Player, c core.Card) {
	panic("implement me")
}

func (p *Player) CanActions(isRounder bool) core.ActionTypes {
	panic("implement me")
}

func (p *Player) RequestAction(actions core.ActionTypes) (playerAction core.PlayerAction) {
	panic("implement me")
}

func (p *Player) String() (s string) {
	s = p.Name
	return
}

func NewPlayer() *Player {
	return &Player{
		Reader: make(chan *core.PlayerAction, 1),
	}
}
