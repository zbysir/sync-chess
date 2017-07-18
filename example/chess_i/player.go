package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
	"errors"
	"encoding/json"
	"fmt"
)

type Player struct {
	Id    string
	Cards chess.Cards         // 手上的牌
	Pong  chess.Cards         // 碰的牌
	Gang  map[chess.Card]Gang // 杠的牌
}

type Gang struct {
	Score    int32          // 分数, 杠需要记录扣分的人. 杠上杠的情况分数不一样
	Receiver chess.Player   // 接收者
	Giver    []chess.Player // 给予者
	Types    GangType
}

type GangType int8

const (
	GT_Bu   GangType = iota + 1 // 补杠/扒杠
	GT_An                       // 暗杠/自杠
	GT_Dian                     // 点杠
)

func (p *Player) GetId() (string) {
	return p.Id
}

func (p *Player) Marshal() (bs []byte, err error) {
	bs, err = json.Marshal(p)
	return
}

func (p *Player) Unmarshal(bs []byte) (err error) {
	err = json.Unmarshal(bs, p)
	return
}

func (p *Player) CanActions(isRounder bool, card chess.Card) chess.ActionTypes {
	as := chess.ActionTypes{}
	if isRounder {
		as = append(as, chess.AT_Play)
	} else {
		as = append(as, chess.AT_Pass, chess.AT_Peng)
	}

	return as
}

func (p *Player) RequestActionAuto(actions chess.ActionTypes, lastCard chess.Card) (playerAction *chess.PlayerActionRequest) {
	for _, a := range actions {
		switch a {
		case chess.AT_HuZiMo, chess.AT_HuDian, chess.AT_HuQiangGang:
			playerAction = &chess.PlayerActionRequest{
				Types: a,
				Card:  lastCard,
			}
			return
		case chess.AT_Play:
			// 自动打最后一张
			lastCard, _ := p.Cards.Last()
			playerAction = &chess.PlayerActionRequest{
				Types: a,
				Card:  lastCard,
			}
			return
		}
	}

	playerAction = &chess.PlayerActionRequest{
		Types: chess.AT_Pass,
	}

	return
}

func (p *Player) DoAction(action *chess.PlayerActionRequest, playerDe chess.Player) (err error) {
	card := action.Card

	switch action.Types {
	case chess.AT_GangAn:
		err = p.GangAn(card)
	case chess.AT_GangBu:
		err = p.GangBu(card)
	case chess.AT_GangDian:
		err = p.GangDian(playerDe, card)
	case chess.AT_Play:
		err = p.Play(card)
	case chess.AT_Peng:
		err = p.Peng(playerDe, card)
	case chess.AT_HuDian:
		err = p.HuDian(playerDe, card)
	case chess.AT_HuZiMo:
		err = p.HuZiMo(card)
	case chess.AT_HuQiangGang:
		err = p.HuQiangGang(playerDe, card)
	case chess.AT_LiangDao:
		err = p.LiangDao(action.Cards, card)
	case chess.AT_Get:
		err = p.GetCard(card)
	}

	if err != nil {
		// todo 通知玩家动作
		//NotifyActionResponse()
	}

	return
}

func (p *Player) SetCards(cards chess.Cards) {
	p.Cards = cards
}

func (p *Player) Play(card chess.Card) (err error) {
	if !p.Cards.Delete(card) {
		err = errors.New("err card " + card.String())
	}
	return
}

// 摸牌
func (p *Player) GetCard(card chess.Card) (err error) {
	p.Cards.Append(card)
	return
}

// 只能碰别人p的牌card
func (p *Player) Peng(player chess.Player, card chess.Card) (err error) {
	return
}

// 亮倒并出牌
func (p *Player) LiangDao(cards chess.Cards, card chess.Card) (err error) {
	return
}

// 点杠
func (p *Player) GangDian(player chess.Player, card chess.Card) (err error) {
	return
}

// 补杠
func (p *Player) GangBu(card chess.Card) (err error) {
	return
}

// 自杠
func (p *Player) GangAn(card chess.Card) (err error) {
	return
}

// 自摸
func (p *Player) HuZiMo(card chess.Card) (err error) {
	return
}

// 点炮
func (p *Player) HuDian(player chess.Player, card chess.Card) (err error) {
	return
}

// 抢杠胡
func (p *Player) HuQiangGang(player chess.Player, card chess.Card) (err error) {
	return
}

func (p *Player) String() (s string) {
	s = fmt.Sprintf("Cards: %v", p.Cards)
	return
}

func NewPlayer() *Player {
	return &Player{}
}
