package core

import (
	"context"
	"log"
	"time"
	"errors"
)

// 管理员管理整个打牌逻辑
// 命令该谁出牌
// 一个房间一个Manager
type Manager struct {
	Id               string        // 管理员唯一标示
	Players          Players       // 所有玩家
	RoundStartPlayer Player        // 每轮开始者
	CardGenerator    CardGenerator // 发牌器
	PlayerLeader     PlayerLeader  // 玩家领导
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

	// 读取存档, 用于down机恢复
	hasStorage := p.Recovery()
	if !hasStorage {
		// 没有存档 则是新的一局,初始化牌桌
		p.startGame()
	}

	// 初始化

	// 胡牌才会结束
	for {
	startRound:

	// 不需要读存档 说明不是第一次运行的现场恢复,则需要存档
		if !hasStorage {
			// 存档
			p.SnapShoot()
		} else {
			hasStorage = false
		}
		ctxTime, _ := context.WithTimeout(ctx, timeout)

	getRoundPlayerAction:
		as := p.GetCanActions(p.RoundStartPlayer, true, 0)
		p.NotifyNeedAction(p.RoundStartPlayer, as)
		a, err := p.GetPlayerAction(ctxTime, p.RoundStartPlayer, as)
		// 超时未响应则执行自动打牌
		if err != nil {
			a = p.RoundStartPlayer.RequestActionAuto(as, 0)
		}
		log.Printf("GetPlayerAction %+v %+v", p.RoundStartPlayer, a)

		switch a.Types {
		case AT_Play:
			// 出牌
			rsp := p.RoundStartPlayer.DoAction(a, p.RoundStartPlayer)
			if rsp.Err != nil {
				goto getRoundPlayerAction
			}
			p.Step(p.RoundStartPlayer, a)
			p.RoundStartPlayer.ResponseAction(rsp)
			p.Players.NotifyOtherPlayerAction(p.RoundStartPlayer, a)

			card := a.Card
			waitActionPlayer := []WaitActionPlayer{}
			for _, player := range p.Players.Exclude(p.RoundStartPlayer) {
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

			getOtherPlayerAction:
				a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions)
				// 超时未响应则执行自动打牌
				if err != nil {
					a = player.RequestActionAuto(wap.CanActions, card)
				}
				log.Printf("GetPlayerAction %+v %+v", player, a)
				switch a.Types {
				case AT_Pass:
					passRsp := &PlayerActionResponse{types: AT_Pass}
					player.ResponseAction(passRsp)
					continue
				case AT_GangDian:
					// 胡牌和杠碰是互斥的
					// 胡过之后就不能杠碰
					if !isHasHu {
						rsp := player.DoAction(a, p.RoundStartPlayer)
						if rsp.Err != nil {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getOtherPlayerAction
						}
						player.ResponseAction(rsp)
						p.Players.NotifyOtherPlayerAction(player, a)

						p.RoundStartPlayer = player
						err := p.GetCard(p.RoundStartPlayer)
						if err != nil {
							// 没牌了,直接结束
							goto end
						}
						goto startRound
					}
				case AT_Peng:
					if !isHasHu {
						rsp := player.DoAction(a, p.RoundStartPlayer)
						if rsp.Err != nil {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getOtherPlayerAction
						}
						player.ResponseAction(rsp)
						p.Players.NotifyOtherPlayerAction(player, a)

						p.RoundStartPlayer = player
						goto startRound
					}
				case AT_HuDian:
					isHasHu = true
					rsp := player.DoAction(a, p.RoundStartPlayer)
					if rsp.Err != nil {
						p.NotifyNeedAction(player, wap.CanActions)
						goto getOtherPlayerAction
					}
					player.ResponseAction(rsp)
					p.Players.NotifyOtherPlayerAction(player, a)

					p.Step(player, a)
				}
			}
			if isHasHu {
				// 有人胡牌就结束,可能是多个人胡
				goto end
			}

			// 若没人动作 下家摸牌
			p.RoundStartPlayer = p.PlayerLeader.Next(p.RoundStartPlayer, p.Players)
			err := p.GetCard(p.RoundStartPlayer)
			if err != nil {
				// 没牌了,直接结束
				goto end
			}

			goto startRound
		case AT_GangAn:
			rsp := p.RoundStartPlayer.DoAction(a, p.RoundStartPlayer)
			if rsp.Err != nil {
				goto getRoundPlayerAction
			}
			p.Step(p.RoundStartPlayer, a)
			p.RoundStartPlayer.ResponseAction(rsp)
			p.Players.NotifyOtherPlayerAction(p.RoundStartPlayer, a)

			// 摸牌
			err := p.GetCard(p.RoundStartPlayer)
			if err != nil {
				// 没牌了,直接结束
				goto end
			}
		case AT_GangBu:
			// 抢杠胡(补杠)
			card := a.Card
			// 判断其他玩家有不有胡
			waitActionPlayer := []WaitActionPlayer{}
			for _, player := range p.Players.Exclude(p.RoundStartPlayer) {
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
				ctxTime, _ := context.WithTimeout(ctx, timeout)
				a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions)
				if err != nil {
					a = player.RequestActionAuto(wap.CanActions, card)
				}
				log.Printf("GetPlayerAction %+v %+v", player, a)
				switch a.Types {
				case AT_Pass:
					passRsp := &PlayerActionResponse{types: AT_Pass}
					player.ResponseAction(passRsp)
					continue
				case AT_HuQiangGang:
					rsp := player.DoAction(a, p.RoundStartPlayer)
					if rsp.Err != nil {
						// 抢杠胡都有错误? 那就当pass了吧
						continue
					}

					player.ResponseAction(rsp)
					p.Players.NotifyOtherPlayerAction(player, a)
					isHasHu = true
				}
			}
			if isHasHu {
				// 有人胡牌就结束,可能是多个人胡
				goto end
			} else {
				// 补杠逻辑
				rsp := p.RoundStartPlayer.DoAction(a, p.RoundStartPlayer)
				if rsp.Err != nil {
					goto getRoundPlayerAction
				}
				p.Step(p.RoundStartPlayer, a)
				p.RoundStartPlayer.ResponseAction(rsp)
				p.Players.NotifyOtherPlayerAction(p.RoundStartPlayer, a)

				// 摸牌
				err := p.GetCard(p.RoundStartPlayer)
				if err != nil {
					// 没牌了,直接结束
					goto end
				}

				goto startRound
			}

		case AT_HuZiMo:
			rsp := p.RoundStartPlayer.DoAction(a, p.RoundStartPlayer)
			if rsp.Err != nil {
				goto getRoundPlayerAction
			}
			p.Step(p.RoundStartPlayer, a)
			p.RoundStartPlayer.ResponseAction(rsp)
			p.Players.NotifyOtherPlayerAction(p.RoundStartPlayer, a)

			// 胡牌就结束
			goto end
		}
	}

end:
	p.Clean()

	log.Printf("end ")
}

func (p *Manager) GetCard(player Player) (err error) {
	card, ok := p.CardGenerator.GetCard()
	if !ok {
		err = errors.New("not cards")
		return
	}
	addCardAction := &PlayerActionRequest{
		Card:  card,
		Types: AT_Get,
	}
	r := player.DoAction(addCardAction, player)
	if r.Err != nil {
		err = r.Err
	}
	player.ResponseAction(r)
	p.Players.NotifyOtherPlayerAction(player, addCardAction)
	return
}

func (p *Manager) startGame() {
	// 选庄家
	p.RoundStartPlayer = p.Players[0]
	//cards := append(TongMulti, TiaoMulti...)
	//cards = append(cards, WanMulti...)
	p.CardGenerator.Reset()
	p.CardGenerator.Shuffle()
}

func (p *Manager) NotifyNeedAction(player Player, actions ActionTypes) {
	// 保存玩家需要的动作, 用于重连时重发
	player.SetValue("NeedAction", actions)

	log.Printf("NeedAction %+v %+v", player, actions)

	player.NotifyNeedAction(actions)
}

// isFirst 是否该他出牌
// card 能够吃的牌(其他人打来的)
func (p *Manager) GetCanActions(player Player, isFirst bool, card Card) (actions ActionTypes) {
	actions = player.CanActions(isFirst, card)
	return
}

// 阻塞获取玩家动作
func (p *Manager) GetPlayerAction(context context.Context, player Player, canActions ActionTypes) (action *PlayerActionRequest, err error) {
	action, err = player.WaitAction(context)
	// 错误的动作, 重新获取
	if !canActions.Contain(action.Types) {
		err = errors.New("except action")
	}
	return
}
