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
	frames             []*ebiten.Image
	partialFrame       float64
}

// SequenceWithAudio represents a video struct WITH audio
type SequenceWithAudio struct {
	sequence     SequenceNoAudio
	audioContext *audio.Context
	song         []byte
}

func (s *SequenceNoAudio) drawFrame(screen *ebiten.Image) {
	screen.DrawImage(s.frames[s.lastFrameNumber], &ebiten.DrawImageOptions{})
	if s.lastFrameNumber != s.currentFrameNumber {
		s.lastFrameNumber = s.currentFrameNumber
	}
}
