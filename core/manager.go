package core

import (
	"context"
	"log"
	"time"
)

// 管理员管理整个打牌逻辑
// 命令该谁出牌
// 一个房间一个Manager
type Manager struct {
	Id               string        // 管理员唯一标示
	Players          Players       // 所有玩家
	RoundStartPlayer Player        // 每轮开始玩家
	CardGenerator    CardGenerator // 发牌器
	Storage
}

type WaitActionPlayer struct {
	Player     Player
	CanActions ActionTypes
}

var timeout = time.Second * 15

// 开始监督
func (p *Manager) StartSupervise() {
	if len(p.Players) == 0 {
		panic("players is nil")
	}
	ctx := context.Background()

	needLoadStorage := true

	// 设置第一个出牌者
	roundPlayer := p.Players[0]
	// 是否需要从Storage装载玩家动作, 用于down机恢复
	if needLoadStorage {
		// 读取存档
		p.Recovery()
	}

	// 胡牌才会结束
	for {
	startRound:
		for _, player := range p.Players {
			p.ClearPlayerMessage(player)
		}

		// 不需要读存档 说明不是第一次运行的现场恢复,则需要存档
		if !needLoadStorage {
			// 存档
			p.SnapShoot()
		} else {
			needLoadStorage = false
		}

		as := p.GetCanActions(roundPlayer, true, 0)
		p.NotifyNeedAction(roundPlayer, as)

	get:
		a, err := p.GetPlayerAction(ctx, roundPlayer, as)
		if err != nil {
			log.Printf("GetPlayerAction err %+v", err)
			goto get
		}
		log.Printf("GetPlayerAction %+v %+v", roundPlayer, a)

		switch a.Types {
		case AT_Play:
			// 出牌, 通知其他人
			card := a.Card
			waitActionPlayer := []WaitActionPlayer{}
			for _, player := range p.Players.Exclude(roundPlayer) {
				as := p.GetCanActions(player, false, card)
				if len(as) != 0 {
					p.NotifyNeedAction(player, as)
					// 胡牌优先选择
					if as.Contain(AT_HuDian) {
						waitActionPlayer = append([]WaitActionPlayer{{Player: player, CanActions: as}}, waitActionPlayer...)
					} else {
						waitActionPlayer = append(waitActionPlayer, WaitActionPlayer{Player: player, CanActions: as})
					}
				}
			}
			isHasHu := false
			for _, wap := range waitActionPlayer {
				player := wap.Player
				ctxTime, _ := context.WithTimeout(ctx, timeout)
			retry:
				a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions)
				// 超时未响应则执行自动打牌
				if err != nil {
					a = player.RequestActionAuto(wap.CanActions)
				}
				log.Printf("GetPlayerAction %+v %+v", player, a)
				switch a.Types {
				case AT_Pass:
					continue
				case AT_GangDian:
					// 胡牌和杠碰是互斥的
					// 胡过之后就不能杠碰
					if !isHasHu {
						// todo 摸牌
						roundPlayer = player
						goto startRound
					}
				case AT_Peng:
					if !isHasHu {
						roundPlayer = player
						goto startRound
					}
				case AT_HuDian:
					isHasHu = true
					rsp := player.DoAction(a, roundPlayer)
					if rsp.Err != nil {
						goto retry
					}
					p.Step(player, a)
					player.ResponseAction(rsp)
				}
			}
			if isHasHu {
				// 有人胡牌就结束,可能是多个人胡
				goto end
			}

			// 若没人动作 下家摸牌
			roundPlayer = p.Players.After(roundPlayer)
			// todo 摸牌
			goto startRound
		case AT_GangAn:

		case AT_GangBu:
			// 抢杠胡(补杠)
			card := a.Card
			// 判断其他玩家有不有胡
			waitActionPlayer := []WaitActionPlayer{}
			for _, player := range p.Players.Exclude(roundPlayer) {
				as := p.GetCanActions(player, false, card)
				if as.Contain(AT_HuDian) {
					as = ActionTypes{AT_HuQiangGang, AT_Pass}
					p.NotifyNeedAction(player, as)
					waitActionPlayer = append(waitActionPlayer, WaitActionPlayer{Player: player, CanActions: as})
				}
			}

			isHasHu := false
			// 有玩家可胡,就获取玩家操作
			for _, wap := range waitActionPlayer {
				player := wap.Player
				a, _ := p.GetPlayerAction(ctx, player, wap.CanActions)
				log.Printf("GetPlayerAction %+v %+v", player, a)
				switch a.Types {
				case AT_Pass:
					continue
				case AT_HuQiangGang:
					rsp := player.DoAction(a, roundPlayer)
					player.ResponseAction(rsp)
					isHasHu = true
				}
			}
			if isHasHu {
				// 有人胡牌就结束,可能是多个人胡
				goto end
			}

			// todo 摸牌
			goto startRound
		case AT_Peng:
			goto startRound
		case AT_HuZiMo:
			// 胡牌就结束
			goto end
		}
	}
end:

	log.Printf("end ")
}

// 响应玩家出牌动作
func (p *Manager) ResponsePlayer(player Player, response *PlayerActionResponse) {

}

func (p *Manager) NotifyNeedAction(player Player, actions []ActionType) {

	log.Printf("NeedAction %+v %+v", player, actions)

	player.Query(actions)
}

// isFirst 是否该他出牌
// card 能够吃的牌(其他人打来的)
func (p *Manager) GetCanActions(player Player, isFirst bool, card Card) (actions ActionTypes) {
	actions = player.GetCanActions(isFirst, card)
	return
}

// 阻塞获取玩家动作
func (p *Manager) GetPlayerAction(context context.Context, player Player, canActions ActionTypes) (action *PlayerActionRequest, err error) {
	for {
		select {
		case <-context.Done():
			err = context.Err()
			return
		case playerAction := <-player.Reader:
			log.Printf("[debug] GetPlayerAction: %+v %+v", player, playerAction)
			// 如果不在说明有误 再次获取
			if !canActions.Contain(playerAction.ActionType) {
				continue
			}
			action = playerAction
			return
		}
	}
}

// 开启接收玩家消息
func (p *Manager) ClearPlayerMessage(player Player) {
	// 清空缓存
	for {
		select {
		case <-player.Reader:
		default:
			return
		}
	}
}
