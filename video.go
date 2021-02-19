package video

import (
	"fmt"
	"image"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nfnt/resize"
)

const (
	zero = "0"
)

// NewSequence creates a new sequence struct
func NewSequence(location string, prefix string, totalFrames int) *SequenceNoAudio {
	numZeroes := int(math.Log10(float64(totalFrames)))

	return &SequenceNoAudio{
		location:           location,
		totalFrames:        totalFrames,
		frames:             getAllImages(location, totalFrames, numZeroes, prefix),
		currentFrameNumber: 0,
		partialFrame:       0.0,
		currentFrameImage:  &ebiten.Image{},
	}
}

// UpdateSequence updates the info in the png sequence so you can draw it with DrawSequence
func UpdateSequence(sequence *SequenceNoAudio, fps int, tps int) {
	sequence.partialFrame += (float64(tps) / float64(fps))
	if math.Floor(sequence.partialFrame) > float64(sequence.currentFrameNumber) {
		sequence.currentFrameNumber++
	}
}

// DrawSequence draws the current frame of the given sequence
func DrawSequence(sequence *SequenceNoAudio, screen *ebiten.Image) {
	sequence.drawFrame(screen)
}

func scaleImage(r image.Rectangle, i image.Image) image.Image {
	return resize.Resize(uint(r.Dx()), uint(r.Dy()), i, resize.Lanczos3)
}

func getAllImages(location string, total int, numZeroes int, prefix string) []image.Image {
	b := make([]image.Image, total)
	numZeroes = int(math.Floor(float64(numZeroes)) + 1)
	var filename, num string
	var z int

	for i := 0; i < total; i++ {
		if i != 0 {
			z = numZeroes - int(math.Floor(float64(math.Log10(float64(i)))+1))
		} else {
			z = numZeroes - 1
		}
		num = strings.Repeat(zero, z)
		filename = location + "/" + prefix + num + strconv.Itoa(i) + ".png"
		file, err := os.Open(filename)
		if err != nil {
			fmt.Println(err)
			fmt.Println("got to line 67")
			os.Exit(1)
		}

		b[i], _, err = image.Decode(file)
		if err != nil {
			fmt.Println(err)
			fmt.Println("got to line 74")
			os.Exit(1)
		}
	}
	return b
}
