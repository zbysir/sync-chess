package chess

import "errors"

var (
	ERR_BadActionTypeNeedRetry = errors.New("BadActionTypeNeedRetry")
	ERR_JoinRoomPlayerExist    = errors.New("JoinRoomPlayerExist") // 存在
)
