package chess

import (
	"github.com/bysir-zl/sync-chess/chess/util/mjlib"
)

// cards 手牌, gui 鬼牌(任意牌)
func IsHu(cards Cards, gui ...Card) (bool) {
	t := genTableCards(cards)
	guiT := make([]int, len(gui))
	for i, c := range gui {
		guiT[i] = int(c - 1)
	}
	ok := mjlib.MHuLib.GetHuInfo(t, 34, guiT...)
	return ok
}

func genTableCards(cards Cards) ([]int) {
	t := make([]int, 34)
	for _, card := range cards {
		t[int(card-1)]++
	}
	return t
}

func init() {
	mjlib.Init()
	mjlib.MTableMgr.LoadTable()
	mjlib.MTableMgr.LoadFengTable()
}
