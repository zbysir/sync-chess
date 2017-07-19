package chess_i

import (
	"github.com/bysir-zl/sync-chess/chess"
	"math/rand"
	"github.com/bysir-zl/sync-chess/example/conf"
)

type CardGenerator struct {
	cards chess.Cards
}

func (p *CardGenerator) Reset() {
	cards := append(chess.TongMulti, chess.TiaoMulti...)
	cards = append(cards, chess.WanMulti...)
	p.cards = cards
}

func (p *CardGenerator) Shuffle() {
	if conf.StdApp.RunModel == "dev" {
		return
	}

	l := len(p.cards)
	for i := range p.cards {
		j := rand.Intn(l)
		p.cards[i], p.cards[j] = p.cards[j], p.cards[i]
	}
}

func (p *CardGenerator) GetCardsSurplus() (cards chess.Cards) {
	return p.cards
}

func (p *CardGenerator) SetCardsSurplus(cards chess.Cards) {
	p.cards = cards
}

func (p *CardGenerator) GetCards(length int) (cards chess.Cards, ok bool) {
	if len(p.cards) < length {
		return
	}
	cards = make(chess.Cards, length)
	copy(cards, p.cards[:length])
	p.cards = p.cards[length:]
	ok = true
	return
}

func NewCardGenerator() *CardGenerator {
	return &CardGenerator{
		cards: chess.Cards{},
	}
}
