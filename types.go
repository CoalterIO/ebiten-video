package video

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SequenceNoAudio represents a video struct without an audio file
type SequenceNoAudio struct {
	location           string
	prefix             string
	currentFrameImage  *ebiten.Image
	lastFrameNumber    int
	currentFrameNumber int
	totalFrames        int
	frames             <-chan *ebiten.Image
	partialFrame       float64
	IsFinished         bool
}

// SequenceWithAudio represents a video struct WITH audio
type SequenceWithAudio struct {
	sequence     SequenceNoAudio
	audioContext *audio.Context
	song         []byte
}

func (s *SequenceNoAudio) drawFrame(screen *ebiten.Image) {
	screen.DrawImage(s.currentFrameImage, &ebiten.DrawImageOptions{})
	if s.lastFrameNumber != s.currentFrameNumber {
		s.lastFrameNumber = s.currentFrameNumber
	}
}

// func (s *SequenceNoAudio) ResetSequence() {
// 	s.currentFrameImage = s.frames[0]
// 	s.currentFrameNumber = 1
// 	s.lastFrameNumber = 0
// 	s.partialFrame = 0.0
// 	s.IsFinished = false
// }
