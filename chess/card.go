package chess

import (
	"fmt"
)

// 一共 27+4+3 = 34 种牌
type Card uint16

var (
	C_Tong map[int]Card // 1-9筒 从0开始计数 对应的是1筒 card=1
	C_Tiao map[int]Card // 1-9条 从0开始计数
	C_Wan  map[int]Card // 1-9万 从0开始计数
	C_Feng map[int]Card // 东南西北 从0开始计数
	C_Zi   map[int]Card // 中发白 从0开始计数
)

var (
	TongMulti Cards
	TiaoMulti Cards
	WanMulti  Cards
	FengMulti Cards
	ZiMulti   Cards
)

func (p Card) String() string {
	s := "ERR"
	if p == 0 {
		s = "0"
	} else if p < C_Tiao[0] {
		s = fmt.Sprint("筒", int(p-C_Tong[0]+1))
	} else if p < C_Wan[0] {
		s = fmt.Sprint("条", int(p-C_Tiao[0]+1))
	} else if p < C_Feng[0] {
		s = fmt.Sprint("万", int(p-C_Wan[0]+1))
	} else if p < C_Zi[0] {
		s = fmt.Sprint("风", int(p-C_Feng[0]+1))
	} else if p < C_Zi[4] {
		s = fmt.Sprint("字", int(p-C_Zi[0]+1))
	}
	return s
}

type Cards []Card

func (p *Cards) Len() int {
	return len(*p)
}

func (p *Cards) Less(i, j int) bool {
	return (*p)[i] < (*p)[j]
}

func (p *Cards) Swap(i, j int) {
	(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
}

func (p *Cards) Append(card Card) {
	*p = append(*p, card)
}

func (p *Cards) Delete(card Card) bool {
	if index := p.Index(card); index != -1 {
		cStar := append(Cards{}, (*p)[:index]...)
		*p = append(cStar, (*p)[index+1:]...)
		return true
	}
	return false
}

func (p *Cards) Last() (card Card, has bool) {
	if len(*p) == 0 {
		return
	}
	card = (*p)[len(*p)-1]
	has = true
	return
}

func (p *Cards) Index(card Card) int {
	for i, c := range *p {
		if c == card {
			return i
		}
	}
	return -1
}

func init() {
	C_Tong = map[int]Card{}
	c := 0
	for i := 1; i <= 9; i++ {
		C_Tong[c] = Card(i)
		c++
	}
	C_Tiao = map[int]Card{}
	c = 0
	for i := 10; i <= 18; i++ {
		C_Tiao[c] = Card(i)
		c++
	}
	C_Wan = map[int]Card{}
	c = 0
	for i := 19; i <= 27; i++ {
		C_Wan[c] = Card(i)
		c++
	}
	C_Feng = map[int]Card{}
	c = 0
	for i := 28; i <= 31; i++ {
		C_Feng[c] = Card(i)
		c++
	}
	C_Zi = map[int]Card{}
	c = 0
	for i := 32; i <= 34; i++ {
		C_Zi[c] = Card(i)
		c++
	}

	// -------------

	TongMulti = make([]Card, len(C_Tong)*4)
	for i, c := range C_Tong {
		TongMulti[i*4] = c
		TongMulti[i*4+1] = c
		TongMulti[i*4+2] = c
		TongMulti[i*4+3] = c
	}

	TiaoMulti = make([]Card, len(C_Tiao)*4)
	for i, c := range C_Tiao {
		TiaoMulti[i*4] = c
		TiaoMulti[i*4+1] = c
		TiaoMulti[i*4+2] = c
		TiaoMulti[i*4+3] = c
	}
	WanMulti = make([]Card, len(C_Wan)*4)
	for i, c := range C_Wan {
		WanMulti[i*4] = c
		WanMulti[i*4+1] = c
		WanMulti[i*4+2] = c
		WanMulti[i*4+3] = c
	}
	FengMulti = make([]Card, len(C_Feng)*4)
	for i, c := range C_Feng {
		FengMulti[i*4] = c
		FengMulti[i*4+1] = c
		FengMulti[i*4+2] = c
		FengMulti[i*4+3] = c
	}
	ZiMulti = make([]Card, len(C_Zi)*4)
	for i, c := range C_Zi {
		ZiMulti[i*4] = c
		ZiMulti[i*4+1] = c
		ZiMulti[i*4+2] = c
		ZiMulti[i*4+3] = c
	}
}
