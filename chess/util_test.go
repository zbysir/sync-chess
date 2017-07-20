package chess

import "testing"

func TestIsHu(t *testing.T) {
	cards := Cards{C_Tong[0], C_Tong[0], C_Tong[0],
				   C_Tong[1], C_Tong[1], C_Tong[1],
				   C_Tong[2], C_Tong[3], C_Tong[4],
				   C_Tong[6], C_Tong[6], C_Tong[6],
				   C_Tiao[8], C_Tiao[8],
	}
	can := IsHu(cards, 0)
	t.Log(can)
}
