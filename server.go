package main

import (
	"sync"

	"github.com/Strum355/go-queue/queue"
)

func (s *servers) getCount() int {
	//not gonna mutex this because am i really gonna cry over
	//an inaccurate result for a single request?
	return s.Count
}

func (s *server) newVoiceInstance() {
	s.VoiceInst = &voiceInst{
		Queue:   queue.New(),
		Done:    make(chan error),
		RWMutex: new(sync.RWMutex),
	}
}

func (s server) nextSong() song {
	return s.VoiceInst.Queue.PopFront().(song)
}

func (s server) addSong(song song) {
	s.VoiceInst.Queue.PushBack(song)
}

func (s server) queueLength() int {
	s.VoiceInst.RLock()
	defer s.VoiceInst.RUnlock()
	return s.VoiceInst.Queue.Len()
}

func (s server) iterateQueue() []song {
	s.VoiceInst.RLock()
	defer s.VoiceInst.RUnlock()
	ret := make([]song, s.VoiceInst.Queue.Len())
	for i, val := range s.VoiceInst.Queue.List() {
		ret[i] = val.(song)
	}
	return ret
}
