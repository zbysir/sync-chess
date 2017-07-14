package core

import (
	"github.com/bysir-zl/bygo/log"
	"bytes"
	"encoding/json"
	"errors"
)

// 用于保存牌局, 实现down机恢复, 重放等功能
type Storage struct {
	manager *Manager
}

type SnapShoot struct {
	Players            Players
	RoundStartPlayerId string
	SurplusCards       Cards
}

type Step struct {
	PlayerId      string
	ActionRequest *PlayerActionRequest
}

// 保存每轮开始(出牌玩家开始出牌之前)快照
func (p *Storage) SnapShoot() {
	players := p.manager.Players
	roundStartPlayer := p.manager.RoundStartPlayer
	surplusCards := p.manager.CardGenerator.GetCardsSurplus()

	s := SnapShoot{
		Players:            players,
		RoundStartPlayerId: roundStartPlayer.GetId(),
		SurplusCards:       surplusCards,
	}
	sBs, err := s.M()
	if err != nil {
		panic(err)
	}
	bs := make([]byte, len(sBs)+1)
	bs[0] = 1
	copy(sBs, bs[1:])

	// todo 存bs

	log.Info("Storage SnapShoot ", players, roundStartPlayer, surplusCards)
}

// 恢复快照,并且读取待运行的操作
func (p *Storage) Recovery() (has bool) {
	// 拉取所有记录, 找到最近一次SnapShoot并恢复
	// 将Step记录转换为manager.AutoAction, 当有AutoAction时manager不在询问player而是自动action

	log.Info("Storage Recoveryed")
	return
}

// 清空这局存档
func (p *Storage) Clean() {
	//players := p.manager.Players
	//roundStartPlayer := p.manager.RoundStartPlayer
	//surplusCards := p.manager.CardGenerator.GetCards()

	log.Info("Storage Cleaned")
	return
}

// 保存玩家操作日志
func (p *Storage) Step(player Player, request *PlayerActionRequest) {
	s := Step{
		PlayerId:      player.GetId(),
		ActionRequest: request,
	}
	sBs, err := json.Marshal(&s)
	if err != nil {
		panic(err)
	}
	bs := make([]byte, len(sBs)+1)
	bs[0] = 2
	copy(sBs, bs[1:])

	// todo 存bs

	log.Info("Storage Step ", player, request)
}

var sp = []byte("@#$%$#@")
var spPlayer = []byte("^&*(*&^")

func (s *SnapShoot) M() (bs []byte, err error) {
	var buff bytes.Buffer
	playerBs := [][]byte{}
	for _, player := range s.Players {
		pbs, e := player.Marshal()
		if e != nil {
			err = e
			return
		}
		playerBs = append(playerBs, pbs)
	}
	// 写入玩家
	buff.Write(bytes.Join(playerBs, spPlayer))
	// 写入局头人
	buff.Write(sp)
	buff.Write([]byte(s.RoundStartPlayerId))
	// 写入剩余卡牌
	buff.Write(sp)
	cardsBs, err := json.Marshal(s.SurplusCards)
	if err != nil {
		return
	}
	buff.Write(cardsBs)

	bs = buff.Bytes()

	return
}

func (s *SnapShoot) UnM(bs []byte, PlayerCreator func() Player) (err error) {
	bsp := bytes.Split(bs, sp)
	if len(bsp) != 3 {
		err = errors.New("bad format")
		return
	}
	// 读取玩家
	if len(bsp[0]) == 0 {
		err = errors.New("bad format: player")
		return
	}
	playerBs := bytes.Split(bsp[0], spPlayer)
	lenPlayer := len(playerBs)
	players := make(Players, lenPlayer)
	for i, pbs := range playerBs {
		player := PlayerCreator()
		err = player.Unmarshal(pbs)
		if err != nil {
			return
		}
		players[i] = player
	}
	// 读取id
	s.RoundStartPlayerId = string(bsp[1])
	// 剩余卡牌
	err = json.Unmarshal(bsp[2], &s.SurplusCards)
	if err != nil {
		return
	}

	return
}

func NewStorage(manager *Manager) *Storage {
	return &Storage{manager: manager}
}
