package main

import (
	"sync"
)

func (s *servers) getCount() int {
	//not gonna mutex this because am i really gonna cry over
	//an inaccurate result for a single request?
	return s.Count
}

func (s *server) newVoiceInstance() {
	s.VoiceInst = &voiceInst {
		Queue: make([]song, 0),
		Done: make(chan error),
		Mutex: new(sync.Mutex),
	}
}