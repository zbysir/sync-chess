package tests

import (
	"testing"
	"time"
	"github.com/bysir-zl/sync-chess/example/chess_i"
	"github.com/bysir-zl/bygo/log"
	"runtime"
	"github.com/bysir-zl/sync-chess/chess"
)

// 验证 当有碰有胡时 胡优先
func TestPlayPengHu(t *testing.T) {

	cg := chess_i.NewCardGenerator()
	pl := chess_i.NewPlayerLeader()
	mh := chess_i.NewMessageHandler()
	m := chess.NewManager("1", cg, pl, mh)
	p1 := "p1"
	p2 := "p2"
	p3 := "p3"

	m.AddPlayer(p1)
	m.AddPlayer(p2)
	m.AddPlayer(p3)

	m.Start()

	time.Sleep(1 * time.Second)

	err := m.WritePlayerAction(p1, &chess.PlayerActionRequest{
		Types: chess.AT_Play,
		Card:  chess.C_Tong[3],
	})
	if err != nil {
		log.Error("test", err)
	}

	runtime.Gosched()
	time.Sleep(1 * time.Second)
	// 这时候玩家2先点击碰
	err = m.WritePlayerAction(p2, &chess.PlayerActionRequest{
		Types: chess.AT_Peng,
		Card:  100,
	})
	if err != nil {
		log.Error("test", err)
	}

	m.Wait()
}
