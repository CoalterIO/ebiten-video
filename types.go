package video

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

// DirectoryDoesNotExistError does what it says on the box
type DirectoryDoesNotExistError struct {
	Path string
}

func (e *DirectoryDoesNotExistError) Error() string {
	return fmt.Sprintf("Error: directory %v does not exist", e.Path)
}

type imageWithOptions struct {
	i *ebiten.Image
	o *ebiten.DrawImageOptions
}

// SequenceNoAudio represents a video struct without an audio file
type SequenceNoAudio struct {
	location           string
	prefix             string
	currentFrameImage  *imageWithOptions
	currentFrameNumber int
	totalFrames        int
	frames             <-chan *imageWithOptions
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
	screen.DrawImage(s.currentFrameImage.i, s.currentFrameImage.o)
}

// func (s *SequenceNoAudio) ResetSequence() {
// 	s.currentFrameImage = s.frames[0]
// 	s.currentFrameNumber = 1
// 	s.lastFrameNumber = 0
// 	s.partialFrame = 0.0
// 	s.IsFinished = false
// }
