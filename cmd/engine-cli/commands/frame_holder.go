package commands

import (
	"sync"

	"github.com/battlesnakeio/engine/controller/pb"
)

type FrameHolder struct {
	sync.RWMutex
	frames []*pb.GameFrame
}

func (fh *FrameHolder) Append(frame *pb.GameFrame) {
	fh.Lock()
	defer fh.Unlock()

	fh.frames = append(fh.frames, frame)
}

func (fh *FrameHolder) Get(index int) *pb.GameFrame {
	fh.RLock()
	defer fh.RUnlock()

	if index < 0 || index >= len(fh.frames) {
		return nil
	}

	return fh.frames[index]
}

func (fh *FrameHolder) Count() int {
	fh.RLock()
	defer fh.RUnlock()

	return len(fh.frames)
}
