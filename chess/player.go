package chess

import "fmt"

type PlayerInterface interface {
	// 设置手牌
	SetCards(cards Cards)

	// 获能进行的动作,应该根据手上的牌判断返回
	// isRounder是否该你出牌
	// card 别人打的牌
	CanActions(isRounder bool, card Card) ActionTypes

	// 当玩家超时, 或者亮倒时, 执行自动打牌
	// 应当从actions中选取一个动作
	// card为要吃的牌, 比如胡哪张牌
	RequestActionAuto(actions ActionTypes, card Card) (playerAction *PlayerActionRequest)

	// 所有打牌动作,当用户请求时应该handle
	// playerDe 被操作者(点炮,点杠,碰等)
	DoAction(action *PlayerActionRequest, playerDe *Player) (err error)

	// 序列化, 保存玩家状态(手上的牌,碰,杠,胡)
	Marshal() (bs []byte, err error)
	Unmarshal(bs []byte) (err error)
}

type Player struct {
	Id      string
	PlayerI PlayerInterface
}

func (p *Player) String() (s string) {
	s = fmt.Sprintf("playerId: %+v", *p)
	return
}

type ActionFrom int32

const (
	AF_Auto    ActionFrom = iota + 1 // 自动打牌
	AF_Player                        // 来至玩家
	AF_Storage                       // 来至存档
)

func (p ActionFrom) String() string {
	s := ""
	switch p {
	case AF_Auto:
		s = "Auto"
	case AF_Player:
		s = "Player"
	case AF_Storage:
		s = "Storage"
	}
	return s
}

// 玩家动作请求
type PlayerActionRequest struct {
	Types      ActionType
	Cards      Cards // 动作哪几张牌, 比如亮倒隐藏刻子时有用
	Card       Card  // 动作哪张牌
	ActionFrom ActionFrom
}

// 玩家动作相应
type PlayerActionResponse struct {
	Types      ActionType
	Card       Card // 动作哪张牌
	ActionFrom ActionFrom
}

// 通知玩家来自其他玩家动作
type PlayerActionNotice struct {
	PlayerFrom *Player
	Types      ActionType
	Card       Card // 动作哪张牌,有部分情况会为空 如摸牌
}

func NewPlayer(id string, playerCreator func() PlayerInterface) *Player {
	return &Player{
		Id:      id,
		PlayerI: playerCreator(),
	}
}
