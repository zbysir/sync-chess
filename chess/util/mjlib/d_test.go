package mjlib

import "testing"

func TestH(t *testing.T) {
	Init()
	MTableMgr.LoadTable()
	MTableMgr.LoadFengTable()

	cards := []int{
		2, 0, 0, 0, 0, 0, 0, 0, 0,
		1, 1, 1, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 2, 2, 2, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0,
	}
	ok := MHuLib.GetHuInfo(cards, 34, 1, 2)
	t.Log(ok)

}

// 211 ns/op
func BenchmarkHu(b *testing.B) {
	Init()
	MTableMgr.LoadTable()
	MTableMgr.LoadFengTable()

	for i := 0; i < b.N; i++ {
		cards := []int{
			1, 0, 1, 0, 0, 0, 0, 0, 0,
			1, 1, 1, 0, 3, 0, 0, 0, 0,
			0, 0, 0, 2, 2, 2, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0,
		}
		MHuLib.GetHuInfo(cards, 34, 34, 34)
		//b.Log(ok)
	}
}
