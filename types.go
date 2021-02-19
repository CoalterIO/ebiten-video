package video

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

// SequenceNoAudio represents a video struct without an audio file
type SequenceNoAudio struct {
	location           string
	prefix             string
	currentFrameImage  *ebiten.Image
	currentFrameNumber int
	totalFrames        int
	frames             []image.Image
	partialFrame       float64
}

// SequenceWithAudio represents a video struct WITH audio
type SequenceWithAudio struct {
	sequence     SequenceNoAudio
	audioContext *audio.Context
	song         []byte
}

func (s *SequenceNoAudio) drawFrame(screen *ebiten.Image) {
	if s.currentFrameNumber > s.totalFrames-1 {
		return
	}
	i := s.frames[s.currentFrameNumber]

	r1 := screen.Bounds()
	r2 := i.Bounds()
	if r1.Size() != r2.Size() {
		i = scaleImage(r1, i)
	}

	s.currentFrameImage = ebiten.NewImageFromImage(i)
	screen.DrawImage(s.currentFrameImage, &ebiten.DrawImageOptions{})
}
