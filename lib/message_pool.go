package sebak

import "boscoin.io/sebak/lib/common"
import "boscoin.io/sebak/lib/error"

type MessagePool struct {
	Messages map[ /* `Message.GetHash()`*/ string]sebakcommon.Message
}

func NewMessagePool() *MessagePool {
	return &MessagePool{map[string]sebakcommon.Message{}}
}

func (mp *MessagePool) HasMessage(m sebakcommon.Message) bool {
	_, ok := mp.Messages[m.GetHash()]
	return ok
}

func (mp *MessagePool) IsEmpty() bool {
	return len(mp.Messages) == 0
}

func (mp *MessagePool) AddMessage(m sebakcommon.Message) (err error) {
	if mp.HasMessage(m) {
		err = sebakerror.ErrorNewButKnownMessage
	} else {
		mp.Messages[m.GetHash()] = m
	}
	return
}
