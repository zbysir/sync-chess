package chess

import (
	"context"
	"time"
	"errors"
	"github.com/bysir-zl/bygo/log"
)

// 管理员管理整个打牌逻辑
// 命令该谁出牌
// 一个房间一个Manager
type Manager struct {
	Id                   string                               // 管理员唯一标示
	storage              *Storage                             // 存档器
	cardGenerator        CardGenerator                        // 发牌器
	playerLeader         PlayerLeader                         // 玩家领导
	isInitFromStorage    bool                                 // 是否是从存档恢复
	isStarted            bool                                 // 是否开始了游戏
	playerActionRequestC map[string]chan *PlayerActionRequest // 玩家动作请求队列
	isNeedStopGame       bool                                 // 是否在这小回合结束后结束游戏
	ctxCancel            func()                               // ctx关闭func
	ctx                  context.Context                      //

	RoundStartPlayer     *Player                // 每轮开始者
	Players              Players                // 所有玩家
	LastPlayerNeedAction map[string]ActionTypes // 最后一次需要玩家的动作, 用于玩家重连重新发送请求
	MessageHandler       MessageHandler         // 消息通知
}

type WaitActionPlayer struct {
	Player     *Player
	CanActions ActionTypes
}

var timeout = time.Second * 10000

// 开始监督(游戏进行中)
func (p *Manager) Start() {
	if p.isStarted {
		return
	}
	p.isStarted = true

	if len(p.Players) == 0 {
		panic("players is nil")
	}
	ctx, cancelFun := context.WithCancel(context.Background())
	p.ctx = ctx
	p.ctxCancel = cancelFun

	if !p.isInitFromStorage {
		p.startGame()
	}

	go func() {
	startRound:

	// 不需要读存档 说明不是第一次运行的现场恢复,则需要存档
		if !p.isInitFromStorage {
			// 存档
			p.storage.SnapShoot()
		} else {
			p.isInitFromStorage = false
		}

		as := p.GetCanActions(p.RoundStartPlayer, true, 0)
		p.NotifyNeedAction(p.RoundStartPlayer, as)
		ctxTime, _ := context.WithTimeout(ctx, timeout)
	getRoundPlayerAction:
		a, err := p.GetPlayerAction(ctxTime, p.RoundStartPlayer, as, 0)
		if err != nil {
			p.NotifyNeedAction(p.RoundStartPlayer, as)
			goto getRoundPlayerAction
		}
		log.Info("GetPlayerAction", "%+v %+v", p.RoundStartPlayer, a)

		switch a.Types {
		case AT_Play:
			// 出牌
			err := p.RoundStartPlayer.PlayerCards.DoAction(a, p.RoundStartPlayer)
			if err != nil {
				p.NotifyNeedAction(p.RoundStartPlayer, as)
				goto getRoundPlayerAction
			}
			p.DoActionAfter(p.RoundStartPlayer, a)

			card := a.Card
			waitActionPlayer := []WaitActionPlayer{}
			for _, player := range p.Players.Exclude(p.RoundStartPlayer) {
				as := p.GetCanActions(player, false, card)
				if len(as) == 0 {
					continue
				}
				// 只有过就不通知用户了
				if len(as) == 1 && as[0] == AT_Pass {
					continue
				}
				p.NotifyNeedAction(player, as)
				// 胡牌优先选择
				if as.Contain(AT_HuDian) {
					waitActionPlayer = append([]WaitActionPlayer{{Player: player, CanActions: as}}, waitActionPlayer...)
				} else {
					waitActionPlayer = append(waitActionPlayer, WaitActionPlayer{Player: player, CanActions: as})
				}
			}
			isHasHu := false
			for _, wap := range waitActionPlayer {
				player := wap.Player

				ctxTime, _ := context.WithTimeout(ctx, timeout)
			getOtherPlayerAction:
				a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions, card)
				// 超时未响应则执行自动打牌
				if err != nil {
					p.NotifyNeedAction(player, wap.CanActions)
					goto getOtherPlayerAction
				}

				log.Info("GetPlayerAction", "%+v %+v", player, a)
				switch a.Types {
				case AT_Pass:
					p.DoActionAfter(player, a)
					continue
				case AT_GangDian:
					// 胡牌和杠碰是互斥的
					// 胡过之后就不能杠碰
					if !isHasHu {
						err := player.PlayerCards.DoAction(a, p.RoundStartPlayer)
						if err != nil {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getOtherPlayerAction
						}
						p.DoActionAfter(player, a)

						p.RoundStartPlayer = player
						err = p.GetCard(p.RoundStartPlayer)
						if err != nil {
							// 没牌了,直接结束
							goto end
						}
						goto startRound
					}
				case AT_Peng:
					if !isHasHu {
						err = player.PlayerCards.DoAction(a, p.RoundStartPlayer)
						if err != nil {
							p.NotifyNeedAction(player, wap.CanActions)
							goto getOtherPlayerAction
						}
						p.DoActionAfter(player, a)
						p.RoundStartPlayer = player
						goto startRound
					}
				case AT_HuDian:
					isHasHu = true
					err := player.PlayerCards.DoAction(a, p.RoundStartPlayer)
					if err != nil {
						p.NotifyNeedAction(player, wap.CanActions)
						goto getOtherPlayerAction
					}
					p.DoActionAfter(player, a)
				}
			}

			// 打牌后检查是否需要结束游戏
			if p.isNeedStopGame {
				goto end
			}

			// 若没人动作 下家摸牌
			p.RoundStartPlayer = p.playerLeader.Next(p.RoundStartPlayer, p.Players)
			err = p.GetCard(p.RoundStartPlayer)
			if err != nil {
				// 没牌了,直接结束
				goto end
			}

			goto startRound
		case AT_GangAn:
			err := p.RoundStartPlayer.PlayerCards.DoAction(a, p.RoundStartPlayer)
			if err != nil {
				p.NotifyNeedAction(p.RoundStartPlayer, as)
				goto getRoundPlayerAction
			}
			p.DoActionAfter(p.RoundStartPlayer, a)

			// 摸牌
			err = p.GetCard(p.RoundStartPlayer)
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
			getBuGangAction:
				a, err := p.GetPlayerAction(ctxTime, player, wap.CanActions, card)
				if err != nil {
					p.NotifyNeedAction(player, wap.CanActions)
					goto getBuGangAction
				}
				log.Info("GetPlayerAction", "%+v %+v", player, a)
				switch a.Types {
				case AT_Pass:
					p.DoActionAfter(player, a)
					continue
				case AT_HuQiangGang:
					err := player.PlayerCards.DoAction(a, p.RoundStartPlayer)
					if err != nil {
						// 抢杠胡都有错误? 那就当pass了吧
						continue
					}
					p.DoActionAfter(player, a)

					isHasHu = true
				}
			}

			if isHasHu {
				// 有人胡牌就检查是否需要结束游戏
				if p.isNeedStopGame {
					goto end
				}
			} else {
				// 补杠逻辑
				err := p.RoundStartPlayer.PlayerCards.DoAction(a, p.RoundStartPlayer)
				if err != nil {
					p.NotifyNeedAction(p.RoundStartPlayer, as)
					goto getRoundPlayerAction
				}
				p.DoActionAfter(p.RoundStartPlayer, a)

				// 摸牌
				err = p.GetCard(p.RoundStartPlayer)
				if err != nil {
					// 没牌了,直接结束
					goto end
				}
				log.Debug("goto startRound")
				goto startRound
			}
		case AT_HuZiMo:
			err := p.RoundStartPlayer.PlayerCards.DoAction(a, p.RoundStartPlayer)
			if err != nil {
				p.NotifyNeedAction(p.RoundStartPlayer, as)
				goto getRoundPlayerAction
			}
			p.DoActionAfter(p.RoundStartPlayer, a)

			// 自摸胡牌检查是否需要结束游戏
			if p.isNeedStopGame {
				goto end
			}
		}

	end:
		p.clear()
		log.Info("end ")
		return
	}()
}

// 让游戏结束
func (p *Manager) Stop() (err error) {
	if !p.isStarted {
		err = errors.New("not started")
		return
	}
	p.ctxCancel()
	p.isNeedStopGame = true
	return
}

// 清理
func (p *Manager) clear() (err error) {
	p.ctxCancel()
	p.storage.Clean()
	p.isStarted = false
	p.isInitFromStorage = false
	p.Players = Players{}
	p.cardGenerator.Reset()
	p.playerActionRequestC = map[string]chan *PlayerActionRequest{}
	p.isNeedStopGame = false
	p.RoundStartPlayer = nil
	p.LastPlayerNeedAction = map[string]ActionTypes{}
	return
}

func (p *Manager) Wait() {
	<-p.ctx.Done()
}

// 玩家成功动作后记录,并响应玩家
func (p *Manager) DoActionAfter(player *Player, action *PlayerActionRequest, ) {
	if action.ActionFrom != AF_Storage {
		p.storage.Step(player, action)
		rsp := &PlayerActionResponse{
			ActionFrom: action.ActionFrom,
			Card:       action.Card,
			Types:      action.Types,
		}
		notice := &PlayerActionNotice{
			Types:      action.Types,
			Card:       action.Card,
			PlayerFrom: player,
		}
		p.MessageHandler.NotifyActionResponse(player.Id, rsp)
		p.NotifyOtherPlayerAction(player, notice)
	}
}

// 摸牌
func (p *Manager) GetCard(player *Player) (err error) {
	cards, ok := p.cardGenerator.GetCards(1)
	if !ok {
		err = errors.New("not cards")
		return
	}
	card := cards[0]
	action := &PlayerActionRequest{
		Card:  card,
		Types: AT_Get,
	}
	err = player.PlayerCards.DoAction(action, player)
	if err != nil {
		return
	}
	rsp := &PlayerActionResponse{
		ActionFrom: action.ActionFrom,
		Card:       action.Card,
		Types:      action.Types,
	}
	notice := &PlayerActionNotice{
		Types:      action.Types,
		Card:       action.Card,
		PlayerFrom: player,
	}
	p.MessageHandler.NotifyActionResponse(player.Id, rsp)
	p.NotifyOtherPlayerAction(player, notice)
	return
}

// 通知其他人消息
func (p *Manager) NotifyOtherPlayerAction(currPlayer *Player, notice *PlayerActionNotice) {
	// pass 消息不需要告知其他人
	if notice.Types == AT_Pass {
		return
	}

	if notice.Types == AT_Get {
		notice.Card = 0
	}

	otherPlayer := p.Players.Exclude(currPlayer)
	for _, player := range otherPlayer {
		p.MessageHandler.NotifyFromOtherPlayerAction(player.Id, notice)
	}

	return
}

// 写入玩家动作
func (p *Manager) WritePlayerAction(playerId string, action *PlayerActionRequest) (err error) {
	_, index := p.Players.Find(playerId)
	if index == -1 {
		err = errors.New("not find player id is " + playerId)
		return
	}

	if _, ok := p.playerActionRequestC[playerId]; !ok {
		err = errors.New("player " + playerId + " not open receive")
		return
	}

	select {
	case p.playerActionRequestC[playerId] <- action:
	default:
		err = errors.New("player " + playerId + " not open receive")
	}
	return
}

func (p *Manager) startGame() {
	// 选庄家
	p.RoundStartPlayer = p.Players[0]
	p.cardGenerator.Reset()
	p.cardGenerator.Shuffle()

	// 发牌
	for _, player := range p.Players {
		if player == p.RoundStartPlayer {
			cards, _ := p.cardGenerator.GetCards(14)
			player.PlayerCards.SetCards(cards)
		} else {
			cards, _ := p.cardGenerator.GetCards(13)
			player.PlayerCards.SetCards(cards)
		}
	}
}

func (p *Manager) NotifyNeedAction(player *Player, actions ActionTypes) {
	// 有动作存档 就不发送通知,而是直接读取
	playerId := player.Id
	if p.storage.HasStep(playerId) {
		return
	}

	// 开启通道
	if c, ok := p.playerActionRequestC[playerId]; ok {
		<-c
	} else {
		p.playerActionRequestC[playerId] = make(chan *PlayerActionRequest, 1)
	}
	// 保存玩家需要的动作, 用于重连时重发
	p.LastPlayerNeedAction[playerId] = actions
	p.MessageHandler.NotifyNeedAction(playerId, actions)
}

// isFirst 是否该他出牌
// card 能够吃的牌(其他人打来的)
func (p *Manager) GetCanActions(player *Player, isFirst bool, card Card) (actions ActionTypes) {
	// 有动作存档 就不需要或者CanAction
	if p.storage.HasStep(player.Id) {
		return
	}
	actions = player.PlayerCards.CanActions(isFirst, card)
	return
}

// 阻塞获取玩家动作
func (p *Manager) GetPlayerAction(ctx context.Context, player *Player, canActions ActionTypes, card Card) (action *PlayerActionRequest, err error) {
	playerId := player.Id
	// 有动作存档 直接读取
	if a, ok := p.storage.PopStep(playerId); ok {
		action = a
		action.ActionFrom = AF_Storage
		return
	}
	select {
	case action = <-p.playerActionRequestC[playerId]:
		p.playerActionRequestC[playerId] <- nil
		delete(p.LastPlayerNeedAction, playerId)

		// 错误的动作, 重新获取
		if !canActions.Contain(action.Types) {
			err = ERR_BadActionTypeNeedRetry
			return
		}
		action.ActionFrom = AF_Player
	case <-ctx.Done():
		// 写入一个空 占满通道让他关闭接收消息
		p.playerActionRequestC[playerId] <- nil
		delete(p.LastPlayerNeedAction, playerId)

		// 超时就自动打牌
		action = p.GetPlayerActionAuto(player, canActions, card)
	}

	return
}

// 获取玩家自动动作
func (p *Manager) GetPlayerActionAuto(player *Player, canActions ActionTypes, card Card) (action *PlayerActionRequest) {
	action = player.PlayerCards.RequestActionAuto(canActions, card)
	action.ActionFrom = AF_Auto
	return
}

// 添加玩家
func (p *Manager) AddPlayer(playerId string) (err error) {
	_, index := p.Players.Find(playerId)
	if index != -1 {
		err = ERR_JoinRoomPlayerExist
		return
	}

	// 如果游戏已经开始, 则不能新加玩家
	if p.isStarted {
		err = errors.New("game is started")
		return
	}

	player := &Player{
		Id:          playerId,
		PlayerCards: p.playerLeader.PlayerCardsCreator(),
	}
	if !p.Players.Add(player) {
		err = errors.New("add player error")
		return
	}
	return
}

// 删除玩家
func (p *Manager) RemovePlayer(playerId string) (err error) {
	// 如果游戏已经开始, 则不能删除玩家
	if p.isStarted {
		err = errors.New("game is started")
		return
	}
	_, index := p.Players.Find(playerId)
	if index == -1 {
		err = errors.New("this player is not in room")
		return
	}

	if !p.Players.RemoveById(playerId) {
		err = errors.New("remove player error")
		return
	}
	return
}

func NewManager(id string, cardGenerator CardGenerator, playerLeader PlayerLeader, messageHandler MessageHandler) *Manager {
	m := &Manager{
		Id:                   id,
		Players:              Players{},
		cardGenerator:        cardGenerator,
		playerLeader:         playerLeader,
		MessageHandler:       messageHandler,
		LastPlayerNeedAction: map[string]ActionTypes{},
		playerActionRequestC: map[string]chan *PlayerActionRequest{},
	}
	m.storage = NewStorage(m)
	// 尝试读档
	m.isInitFromStorage = m.storage.Recovery()
	if m.isInitFromStorage {
		m.Start()
	}
	return m
}
