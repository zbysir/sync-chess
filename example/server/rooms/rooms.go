package rooms

import (
	"github.com/bysir-zl/sync-chess/chess"
	"github.com/bysir-zl/sync-chess/example/chess_i"
	"sync"
)

var RoomMap = map[string]*Room{}
var lock sync.RWMutex

func FindOrCreateRoom(roomId string) (room *Room, err error) {
	lock.Lock()
	defer lock.Unlock()
	m, ok := RoomMap[roomId]
	if !ok {
		cg := chess_i.NewCardGenerator()
		pl := chess_i.NewPlayerLeader()
		mh := chess_i.NewMessageHandler()
		m = &Room{chess.NewManager(roomId, cg, pl, mh)}

		RoomMap[m.Id] = m
	}
	room = m
	return
}
