package core

import "fmt"

// 麻将?

type Card uint16

var (
	C_Tong map[int]Card // 1-9筒
	C_Tiao map[int]Card // 1-9条
	C_Wan  map[int]Card // 1-9万
	C_Feng map[int]Card // 东南西北 从1开始计数
	C_Zi   map[int]Card // 中发白 从1开始计数
)

var (
	TongMulti []Card
	TiaoMulti []Card
	WanMulti  []Card
	FengMulti []Card
	ZiMulti   []Card
)

func (p Card) String() string {
	s := "ERR"
	if p < C_Tiao[0] {
		s = fmt.Sprint("筒", p-C_Tong[0]+1)
	} else if p < C_Wan[0] {
		s = fmt.Sprint("条", p-C_Tiao[0]+1)
	} else if p < C_Feng[0] {
		s = fmt.Sprint("万", p-C_Wan[0]+1)
	} else if p < C_Zi[0] {
		s = fmt.Sprint("风", p-C_Feng[0]+1)
	} else if p < C_Zi[4] {
		s = fmt.Sprint("字", p-C_Zi[0]+1)
	}
	return s
}

type Cards []Card

func (p *Cards) Append(card Card) {
	*p = append(*p, card)
}

func (p *Cards) Delete(card Card) bool {
	if index := p.Index(card); index != -1 {
		*p = append((*p)[:index], (*p)[index+1:]...)
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
	c := 1
	for i := 1; i <= 9; i++ {
		C_Tong[c] = Card(i)
		c++
	}
	C_Tiao = map[int]Card{}
	c = 1
	for i := 10; i <= 18; i++ {
		C_Tiao[c] = Card(i)
		c++
	}
	C_Wan = map[int]Card{}
	c = 1
	for i := 19; i <= 27; i++ {
		C_Wan[c] = Card(i)
		c++
	}
	C_Feng = map[int]Card{}
	c = 1
	for i := 28; i <= 31; i++ {
		C_Feng[c] = Card(i)
		c++
	}
	C_Zi = map[int]Card{}
	c = 1
	for i := 32; i <= 34; i++ {
		C_Zi[c] = Card(i)
		c++
	}

	//

	TongMulti := make([]Card, len(C_Tong)*4)
	for i, c := range C_Tong {
		TongMulti[i-1] = c
		TongMulti[i] = c
		TongMulti[i+1] = c
		TongMulti[i+2] = c
	}

	TiaoMulti := make([]Card, len(C_Tiao)*4)
	for i, c := range C_Tiao {
		TiaoMulti[i-1] = c
		TiaoMulti[i] = c
		TiaoMulti[i+1] = c
		TiaoMulti[i+2] = c
	}
	WanMulti := make([]Card, len(C_Wan)*4)
	for i, c := range C_Wan {
		WanMulti[i-1] = c
		WanMulti[i] = c
		WanMulti[i+1] = c
		WanMulti[i+2] = c
	}
	FengMulti := make([]Card, len(C_Feng)*4)
	for i, c := range C_Feng {
		FengMulti[i-1] = c
		FengMulti[i] = c
		FengMulti[i+1] = c
		FengMulti[i+2] = c
	}
	ZiMulti := make([]Card, len(C_Zi)*4)
	for i, c := range C_Zi {
		ZiMulti[i-1] = c
		ZiMulti[i] = c
		ZiMulti[i+1] = c
		ZiMulti[i+2] = c
	}
}
