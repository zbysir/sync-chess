package chess

import "github.com/bysir-zl/sync-chess/core"

type CardGenerator struct {
	cards core.Cards
}

func (p *CardGenerator) Reset() {
	cards := append(core.TongMulti, core.TiaoMulti...)
	cards = append(cards, core.WanMulti...)
	p.cards = cards
}

func (p *CardGenerator) Shuffle() {
	// todo shuffle
	// dev mode can do not this
}

func (p *CardGenerator) GetCardsSurplus() (cards core.Cards) {
	return p.cards
}

func (p *CardGenerator) GetCards(length int) (cards core.Cards, ok bool) {
	if len(p.cards) < length {
		return
	}
	cards = p.cards[:length]
	p.cards = p.cards[length:]
	ok = true
	return
}

func NewCardGenerator() *CardGenerator {
	return &CardGenerator{
		cards: core.Cards{},
	}
}
