package chess

import (
	"github.com/bysir-zl/sync-chess/core"
	"errors"
	"encoding/json"
	"fmt"
)

type Player struct {
	Cards     core.Cards `json:"Cards"`
	PengCards core.Cards
}

func (p *Player) Marshal() (bs []byte, err error) {
	bs, err = json.Marshal(p)
	return
}

func (p *Player) Unmarshal(bs []byte) (err error) {
	err = json.Unmarshal(bs, p)
	return
}

func (p *Player) CanActions(isRounder bool, card core.Card) core.ActionTypes {
	as := core.ActionTypes{}
	if isRounder {
		as = append(as, core.AT_Pass, core.AT_Play)
	} else {
		as = append(as, core.AT_Pass, core.AT_Peng)
	}

	return as
}

func (p *Player) RequestActionAuto(actions core.ActionTypes, lastCard core.Card) (playerAction *core.PlayerActionRequest) {
	for _, a := range actions {
		switch a {
		case core.AT_HuZiMo, core.AT_HuDian, core.AT_HuQiangGang:
			playerAction = &core.PlayerActionRequest{
				Types: a,
				Card:  lastCard,
			}
			return
		case core.AT_Play:
			// 自动打最后一张
			lastCard, _ := p.Cards.Last()
			playerAction = &core.PlayerActionRequest{
				Types: a,
				Card:  lastCard,
			}
			return
		}
	}

	playerAction = &core.PlayerActionRequest{
		Types: core.AT_Pass,
	}

	return
}

func (p *Player) DoAction(action *core.PlayerActionRequest, playerDe *core.Player) (err error) {
	card := action.Card

	switch action.Types {
	case core.AT_GangAn:
		err = p.GangAn(card)
	case core.AT_GangBu:
		err = p.GangBu(card)
	case core.AT_GangDian:
		err = p.GangDian(playerDe, card)
	case core.AT_Play:
		err = p.Play(card)
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
		err = p.GetCard(card)
	}

	return
}

func (p *Player) SetCards(cards core.Cards) {
	p.Cards = cards
}

func (p *Player) Play(card core.Card) (err error) {
	if !p.Cards.Delete(card) {
		err = errors.New("err card " + card.String())
	}
	return
}

// 摸牌
func (p *Player) GetCard(card core.Card) (err error) {
	p.Cards.Append(card)
	return
}

// 只能碰别人p的牌card
func (p *Player) Peng(player *core.Player, card core.Card) (err error) {
	return
}

// 亮倒并出牌
func (p *Player) LiangDao(cards core.Cards, card core.Card) (err error) {
	return
}

// 点杠
func (p *Player) GangDian(player *core.Player, card core.Card) (err error) {
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
func (p *Player) HuDian(player *core.Player, card core.Card) (err error) {
	return
}

// 抢杠胡
func (p *Player) HuQiangGang(player *core.Player, card core.Card) (err error) {
	return
}

func (p *Player) String() (s string) {
	s = fmt.Sprintf("Cards: %v", p.Cards)
	return
}

func NewPlayer() *Player {
	return &Player{}
}
