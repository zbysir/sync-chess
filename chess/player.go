package chess

import "fmt"

type Player struct {
	Id          string
	PlayerCards PlayerCards
}

func (p *Player) String() (s string) {
	s = fmt.Sprintf("playerId: %+v", *p)
	return
}
