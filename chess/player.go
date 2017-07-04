package chess

import (
	"github.com/bysir-zl/sync-chess/core"
)

type Player struct {
	Name   string
	Reader chan *core.PlayerActionRequest
	cards  core.Cards
}

func (p *Player) CanActions(isRounder bool) core.ActionTypes {
	panic("implement me")
}

func (p *Player) RequestAction(types core.ActionTypes) (playerAction core.PlayerActionRequest) {
	panic("implement me")
}

func (p *Player) DoAction(action *core.PlayerActionRequest, playerDe Player) (response *core.PlayerActionResponse) {
	card := action.Card
	response = core.NewActionResponse()
	var err error

	switch action.Types {
	case core.AT_GangAn:
		err = p.GangAn(card)
	case core.AT_GangBu:
		err = p.GangBu(card)
	case core.AT_GangDian:
		err = p.GangDian(playerDe, card)
	case core.AT_Play:
		err = p.Play()
	case core.AT_Peng:
		err = p.Peng(playerDe, card)
	case core.AT_HuDian:
		err = p.HuDian(playerDe, card)
	case core.AT_HuZiMo:
		err = p.HuZiMo(card)
	case core.AT_HuQiangGang:
		err = p.HuQiangGang(playerDe, card)
	case core.AT_LiangDao:
		err = p.LiangDao(action.Cards, card)
	case core.AT_Get:
		e, card := p.GetCard()
		err = e
		response.Card = card
	}
	response.Err = err
	return
}

func (p *Player) DoActionAuto(action *core.PlayerActionRequest, playerDe core.Player) (response *core.PlayerActionResponse) {
	panic("implement me")
}

// 出牌
func (p *Player) Play() (err error) {
	return
}

// 摸牌
func (p *Player) GetCard() (err error, card core.Card) {
	return
}

// 只能碰别人p的牌card
func (p *Player) Peng(player Player, card core.Card) (err error) {
	return
}

// 亮倒并出牌
func (p *Player) LiangDao(cards core.Cards, card core.Card) (err error) {
	return
}

// 点杠
func (p *Player) GangDian(player Player, card core.Card) (err error) {
	return
}

// 补杠
func (p *Player) GangBu(card core.Card) (err error) {
	return
}

// 自杠
func (p *Player) GangAn(card core.Card) (err error) {
	return
}

// 自摸
func (p *Player) HuZiMo(card core.Card) (err error) {
	return
}

// 点炮
func (p *Player) HuDian(player Player, card core.Card) (err error) {
	return
}

// 抢杠胡
func (p *Player) HuQiangGang(player Player, card core.Card) (err error) {
	return
}

func (p *Player) String() (s string) {
	s = p.Name
	return
}

func NewPlayer() *Player {
	return &Player{
		Reader: make(chan *core.PlayerActionRequest, 1),
	}
}
