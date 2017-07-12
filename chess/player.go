package chess

import (
	"github.com/bysir-zl/sync-chess/core"
	"context"
	"errors"
	"github.com/bysir-zl/bygo/log"
	"fmt"
)

type Player struct {
	Name          string
	Reader        chan *core.PlayerActionRequest
	cards         core.Cards
	isOpenReceive bool
	data          map[string]interface{}
}

func (p *Player) CanActions(isRounder bool, card core.Card) core.ActionTypes {
	as := core.ActionTypes{}
	if isRounder {
		as = append(as, core.AT_Pass, core.AT_Play)
	} else {
		as = append(as, core.AT_Pass)
	}

	return as
}

// 通知玩家动作
func (p *Player) NotifyNeedAction(actions core.ActionTypes) {
	p.isOpenReceive = true
	log.Info("NeedAction", "%+v %+v", p, actions)
	return
}

// 获取玩家动作
func (p *Player) WaitAction(ctx context.Context) (playerAction *core.PlayerActionRequest, err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
		p.isOpenReceive = false
		return
	case playerAction = <-p.Reader:
		p.isOpenReceive = false
		return
	}
	return
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
			lastCard, _ := p.cards.Last()
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

func (p *Player) ResponseAction(response *core.PlayerActionResponse) () {
	log.Info("ResponseAction", "=>", p, response)
}

func (p *Player) NotifyFromOtherPlayerAction(notice *core.PlayerActionNotice) () {
	// 发送消息给自己
	log.Info("NotifyFromOtherPlayerAction", "=>", p, notice)
}

func (p *Player) SetValue(key string, value interface{}) {
	p.data[key] = value
}

func (p *Player) GetValue(key string) (value interface{}, ok bool) {
	value, ok = p.data[key]
	return
}

func (p *Player) DoAction(action *core.PlayerActionRequest, playerDe core.Player) (response *core.PlayerActionResponse) {
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
		err = p.Play(card)
		response.Card = card
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
		response.Card = card
	}
	response.Err = err

	return
}

func (p *Player) SetCards(cards core.Cards) {
	p.cards = cards
}

func (p *Player) WriteAction(action *core.PlayerActionRequest) (err error) {
	if !p.isOpenReceive {
		err = errors.New("not open receive")
		return
	}

	select {
	case p.Reader <- action:
	default:
		err = errors.New("channel is full")
		return
	}
	return
}

// 出牌
func (p *Player) Play(card core.Card) (err error) {
	if !p.cards.Delete(card) {
		err = errors.New("err card " + card.String())
	}
	return
}

// 摸牌
func (p *Player) GetCard(card core.Card) (err error) {
	p.cards.Append(card)
	return
}

// 只能碰别人p的牌card
func (p *Player) Peng(player core.Player, card core.Card) (err error) {
	return
}

// 亮倒并出牌
func (p *Player) LiangDao(cards core.Cards, card core.Card) (err error) {
	return
}

// 点杠
func (p *Player) GangDian(player core.Player, card core.Card) (err error) {
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
func (p *Player) HuDian(player core.Player, card core.Card) (err error) {
	return
}

// 抢杠胡
func (p *Player) HuQiangGang(player core.Player, card core.Card) (err error) {
	return
}

func (p *Player) String() (s string) {
	s = fmt.Sprintf("Name: %s, Cards: %v", p.Name, p.cards)
	return
}

func NewPlayer() *Player {
	return &Player{
		Reader: make(chan *core.PlayerActionRequest, 1),
		data:   map[string]interface{}{},
	}
}
