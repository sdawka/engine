package commands

import (
	"sync"

	"github.com/battlesnakeio/engine/controller/pb"
)

type frameHolder struct {
	sync.RWMutex
	frames []*pb.GameFrame
}

func (fh *frameHolder) append(frame *pb.GameFrame) {
	fh.Lock()
	defer fh.Unlock()

	fh.frames = append(fh.frames, frame)
}

func (fh *frameHolder) get(index int) *pb.GameFrame {
	fh.RLock()
	defer fh.RUnlock()

	if index < 0 || index >= len(fh.frames) {
		return nil
	}

	return fh.frames[index]
}

func (fh *frameHolder) count() int {
	fh.RLock()
	defer fh.RUnlock()

	return len(fh.frames)
}
