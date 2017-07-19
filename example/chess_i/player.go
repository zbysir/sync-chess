package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
	"errors"
	"encoding/json"
	"fmt"
	"github.com/bysir-zl/bygo/log"
)

type Player struct {
	Id          string
	Cards       chess.Cards // 手上的牌
	PlayedCards chess.Cards // 打出的牌
	PengCards   chess.Cards // 碰的牌
	GangCards   Gangs       // 杠的牌
	Hu          Hu

	manager *chess.Manager
}

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
		if p.CanGangAn() {
			as = append(as, chess.AT_GangAn)
		}
		if p.CanGangBu() {
			as = append(as, chess.AT_GangBu)
		}
	} else {
		as = append(as, chess.AT_Pass)
		if p.CanPeng(card) {
			as = append(as, chess.AT_Peng)
		}
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
		err = p.GangBu()
	case chess.AT_GangDian:
		err = p.GangDian(playerDe)
	case chess.AT_Play:
		err = p.Play(card)
	case chess.AT_Peng:
		err = p.Peng(playerDe)
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

	if err == nil {
		// 通知玩家动作
		NotifyActionResponse(p.Id, action)
		NotifyFromOtherPlayerAction(p.Id, p.manager.Players.Ids(), action)
	}

	return
}

func (p *Player) SetCards(cards chess.Cards) {
	p.Cards = cards
}

func (p *Player) Play(card chess.Card) (err error) {
	if !p.Cards.Delete(card) {
		err = errors.New("err card " + card.String())
		return
	}
	p.PlayedCards.Append(card)
	return
}

// 动作检查 --------------

func (p *Player) CanPeng(card chess.Card) (ok bool) {
	if p.PengCards.Index(card) != -1 {
		return
	}

	count := 0
	for _, c := range p.Cards {
		if c == card {
			count++
		}
	}
	log.Info("CanPeng", count)
	// 手上有3张也是能碰的 :-D
	ok = count >= 2
	return
}

// 暗杠
func (p *Player) CanGangAn() (ok bool) {
	cardNum := map[chess.Card]int{}
	for _, c := range p.Cards {
		cardNum[c]++
	}

	for _, num := range cardNum {
		if num == 4 {
			ok = true
			return
		}
	}

	return
}

// 补杠
func (p *Player) CanGangBu() (ok bool) {
	card, ok := p.Cards.Last()
	if !ok {
		return
	}
	ok = p.PengCards.Index(card) != -1
	return
}

// 直杠
func (p *Player) CanGangZhi(card chess.Card) (ok bool) {
	count := 0
	for _, c := range p.Cards {
		if c == card {
			count++
		}
	}

	ok = count == 3
	return
}

// 摸牌
func (p *Player) GetCard(card chess.Card) (err error) {
	p.Cards.Append(card)
	return
}

// 只能碰别人player的最后打的牌
func (p *Player) Peng(player chess.Player) (err error) {
	playerDe := player.(*Player)
	card, ok := playerDe.PlayedCards.Last()
	if !ok {
		err = errors.New("sb")
		return
	}
	count := 0
	for _, c := range p.Cards {
		if c == card {
			count++
		}
	}
	if count < 2 {
		err = errors.New("sb")
		return
	}
	// 删除两张牌
	p.Cards.Delete(card)
	p.Cards.Delete(card)

	p.PengCards.Append(card)

	return
}

// 亮倒并出牌
func (p *Player) LiangDao(cards chess.Cards, card chess.Card) (err error) {
	return
}

// 点杠
func (p *Player) GangDian(player chess.Player) (err error) {
	playerDe := player.(*Player)
	card, ok := playerDe.PlayedCards.Last()
	if !ok {
		err = errors.New("sb")
		return
	}

	count := 0
	for _, c := range p.Cards {
		if c == card {
			count++
		}
	}
	if count != 3 {
		err = errors.New("sb")
		return
	}

	// 删除三张牌
	p.Cards.Delete(card)
	p.Cards.Delete(card)
	p.Cards.Delete(card)

	p.GangCards.Add(&Gang{
		Card:  card,
		Giver: []string{player.GetId()},
		Score: 1,
		Types: GT_Dian,
	})
	return
}

// 补杠
func (p *Player) GangBu() (err error) {
	card, ok := p.Cards.Last()
	if !ok {
		err = errors.New("sb")
		return
	}
	if !p.PengCards.Delete(card) {
		err = errors.New("sb")
		return
	}

	p.GangCards.Add(&Gang{
		Card:  card,
		Giver: p.manager.Players.Exclude(p).Ids(),
		Score: 1,
		Types: GT_Bu,
	})
	return
}

// 暗杠
func (p *Player) GangAn(card chess.Card) (err error) {
	count := 0
	for _, c := range p.Cards {
		if c == card {
			count++
		}
	}
	if count != 4 {
		err = errors.New("sb")
		return
	}

	// 删除四张牌
	p.Cards.Delete(card)
	p.Cards.Delete(card)
	p.Cards.Delete(card)
	p.Cards.Delete(card)

	p.GangCards.Add(&Gang{
		Card:  card,
		Giver: p.manager.Players.Exclude(p).Ids(),
		Score: 2,
		Types: GT_An,
	})

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

// 能告知自己的信息
func (p *Player) InfoSelf() interface{} {
	return map[string]interface{}{
		"Id":          p.Id,
		"Cards":       p.Cards,
		"PlayedCards": p.PlayedCards,
		"Pengs":       p.PengCards,
		"Gangs":       p.GangCards,
	}
}

// 能告知别人的信息
func (p *Player) InfoOther() interface{} {
	return map[string]interface{}{
		"Id":          p.Id,
		"CardsLen":    len(p.Cards),
		"PlayedCards": p.PlayedCards,
		"Pengs":       p.PengCards,
		"Gangs":       p.GangCards,
	}
}

func NewPlayer(id string, manager *chess.Manager) *Player {
	return &Player{
		Id:          id, manager: manager,
		Cards:       chess.Cards{},
		PengCards:   chess.Cards{},
		GangCards:   Gangs{},
		PlayedCards: chess.Cards{},
	}
}
