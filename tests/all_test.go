package tests

import (
	"testing"
	"github.com/bysir-zl/sync-chess/chess"
	"time"
)

// 验证 当有碰有胡时 胡优先
func TestPlayPengHu(t *testing.T) {
	p1 := chess.NewPlayer()
	p2 := chess.NewPlayer()
	p3 := chess.NewPlayer()
	p1.Name = "p1"
	p2.Name = "p2"
	p3.Name = "p3"
	m := chess.Manager{
		Players: chess.Players{
			p1, p2, p3,
		},
	}

	p1.CanActions = chess.ActionTypes{chess.AT_Play}
	go m.StartSupervise()

	time.Sleep(1 * time.Second)
	p1.Reader <- &chess.PlayerAction{
		ActionType: chess.AT_Play,
		Card:       100,
	}
	// 模拟玩家3可胡
	p3.CanActions = chess.ActionTypes{chess.AT_Hu, chess.AT_Pass}
	// 玩家2可碰
	p2.CanActions = chess.ActionTypes{chess.AT_Peng}

	// 这时候玩家2先点击碰
	p2.Reader <- &chess.PlayerAction{
		ActionType: chess.AT_Peng,
		Card:       100,
	}
	p2.CanActions = chess.ActionTypes{chess.AT_Play}

	way := 1
	switch way {
	case 1:
		// 玩家3点击胡
		// 期望结果是 玩家3胡了 游戏结束
		p3.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Hu,
			Card:       100,
		}
	case 2:
		// 另: 玩家3点pass
		// 期望结果是 玩家2碰了
		p3.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Pass,
			Card:       100,
		}
	}

	time.Sleep(1 * time.Hour)
}

// 验证自摸胡
func TestPlayZiMoHu(t *testing.T) {
	p1 := chess.NewPlayer()
	p2 := chess.NewPlayer()
	p3 := chess.NewPlayer()
	p1.Name = "p1"
	p2.Name = "p2"
	p3.Name = "p3"
	m := chess.Manager{
		Players: chess.Players{
			p1, p2, p3,
		},
	}

	p1.CanActions = chess.ActionTypes{chess.AT_Hu}
	go m.StartSupervise()

	time.Sleep(1 * time.Second)
	p1.Reader <- &chess.PlayerAction{
		ActionType: chess.AT_Hu,
		Card:       100,
	}

	time.Sleep(1 * time.Hour)
}

// 验证抢杠逻辑
func TestPlayQiangHang(t *testing.T) {
	p1 := chess.NewPlayer()
	p2 := chess.NewPlayer()
	p3 := chess.NewPlayer()
	p1.Name = "p1"
	p2.Name = "p2"
	p3.Name = "p3"
	m := chess.Manager{
		Players: chess.Players{
			p1, p2, p3,
		},
	}

	p1.CanActions = chess.ActionTypes{chess.AT_Gang}
	go m.StartSupervise()

	time.Sleep(1 * time.Second)
	p1.Reader <- &chess.PlayerAction{
		ActionType: chess.AT_Gang,
		Card:       100,
	}
	// 如果杠成功, 则会提示该p1出牌
	p1.CanActions = chess.ActionTypes{chess.AT_Play}

	p2.CanActions = chess.ActionTypes{chess.AT_Hu, chess.AT_Pass}
	p3.CanActions = chess.ActionTypes{chess.AT_Hu, chess.AT_Pass}

	way := 4

	switch way {
	case 1:
		// p1胡牌,p2过
		// 期望是只有p1胡牌, 游戏结束
		p2.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Hu,
			Card:       100,
		}

		p3.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Pass,
			Card:       100,
		}
	case 2:
		// p2胡牌,p1过
		// 期望是只有p2胡牌, 游戏结束
		p2.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Pass,
			Card:       100,
		}

		p3.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Hu,
			Card:       100,
		}
	case 3:
		// 都胡
		// 期望是两家都胡牌, 游戏结束
		p2.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Hu,
			Card:       100,
		}

		p3.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Hu,
			Card:       100,
		}
	case 4:
		// 都过
		// 期望p1 杠成功, 通知p1出牌
		p2.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Pass,
			Card:       100,
		}

		p3.Reader <- &chess.PlayerAction{
			ActionType: chess.AT_Pass,
			Card:       100,
		}
	}

	time.Sleep(1 * time.Hour)
}

// 测试错误请求
func TestErrPlay(t *testing.T) {
	p1 := chess.NewPlayer()
	p2 := chess.NewPlayer()
	p3 := chess.NewPlayer()
	p1.Name = "p1"
	p2.Name = "p2"
	p3.Name = "p3"
	m := chess.Manager{
		Players: chess.Players{
			p1, p2, p3,
		},
	}

	p1.CanActions = chess.ActionTypes{chess.AT_Play}
	go m.StartSupervise()

	time.Sleep(1 * time.Second)

	{
		// 不允许杠操作
		// 期望是能收到消息, 但不进入打牌逻辑
		p1.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Gang,
			Card:       1,
		})
	}

	p1.WriteAction(&chess.PlayerAction{
		ActionType: chess.AT_Play,
		Card:       100,
	})
	{
		// 正常操作之后会关闭输入
		// 期望是不能收到消息
		p1.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Gang,
			Card:       2,
		})
		p1.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Gang,
			Card:       3,
		})
	}

	{
		// 此时还没轮到p2
		// 期望是收不到消息
		p2.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Play,
			Card:       1,
		})
		p2.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Play,
			Card:       2,
		})
	}

	p2.CanActions = chess.ActionTypes{chess.AT_Peng}
	time.Sleep(1 * time.Second)
	{
		// 只能碰, 但是发送了play消息
		// 期望是收到消息, 但不进入打牌状态
		p2.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Play,
			Card:       100,
		})
	}
	p2.WriteAction(&chess.PlayerAction{
		ActionType: chess.AT_Peng,
		Card:       100,
	})
	p2.CanActions = chess.ActionTypes{chess.AT_Play}

	time.Sleep(1 * time.Hour)
}

// 验证 当有碰有胡时 胡优先 (错误命令)
func TestPlayPengHuErr(t *testing.T) {
	p1 := chess.NewPlayer()
	p2 := chess.NewPlayer()
	p3 := chess.NewPlayer()
	p1.Name = "p1"
	p2.Name = "p2"
	p3.Name = "p3"
	m := chess.Manager{
		Players: chess.Players{
			p1, p2, p3,
		},
	}

	p1.CanActions = chess.ActionTypes{chess.AT_Play}
	go m.StartSupervise()

	time.Sleep(1 * time.Second)
	p1.Reader <- &chess.PlayerAction{
		ActionType: chess.AT_Play,
		Card:       100,
	}

	// 模拟玩家3可胡
	p3.CanActions = chess.ActionTypes{chess.AT_Hu, chess.AT_Pass}
	// 玩家2可碰
	p2.CanActions = chess.ActionTypes{chess.AT_Peng}

	{
		// 又发送错误的出牌动作,由于一次只能有一个待处理命令,所以
		// 期望是 收不到消息
		p2.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Play,
			Card:       1,
		})
	}

	way := 2
	time.Sleep(1 * time.Millisecond)
	switch way {
	case 1:
		// 玩家3点击胡
		// 期望是 玩家3胡了 游戏结束
		p3.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Hu,
			Card:       100,
		})
	case 2:
		// 另: 玩家3点pass
		// 期望是 玩家2碰了,玩家2出牌
		p3.WriteAction(&chess.PlayerAction{
			ActionType: chess.AT_Pass,
			Card:       100,
		})
	}

	// 这时候玩家2碰
	time.Sleep(1 * time.Millisecond)
	p2.WriteAction(&chess.PlayerAction{
		ActionType: chess.AT_Peng,
		Card:       100,
	})



	p2.CanActions = chess.ActionTypes{chess.AT_Play}
	time.Sleep(1 * time.Millisecond)
	p2.WriteAction(&chess.PlayerAction{
		ActionType: chess.AT_Play,
		Card:       1,
	})

	time.Sleep(1 * time.Hour)
}