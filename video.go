package video

import (
	"fmt"
	"image"
	_ "image/png"
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
func NewSequence(location string, prefix string, totalFrames int, screenWidth int, screenHeight int) *SequenceNoAudio {
	numZeroes := int(math.Log10(float64(totalFrames)))
	frames := getAllImages(location, totalFrames, numZeroes, prefix, screenWidth, screenHeight)

	return &SequenceNoAudio{
		location:           location,
		totalFrames:        totalFrames,
		frames:             frames,
		currentFrameNumber: 1,
		lastFrameNumber:    0,
		partialFrame:       0.0,
		currentFrameImage:  frames[0],
	}
}

// UpdateSequence updates the info in the png sequence so you can draw it with DrawSequence
func UpdateSequence(sequence *SequenceNoAudio, fps int, tps int) {
	if sequence.currentFrameNumber >= sequence.totalFrames {
		return
	}
	sequence.partialFrame += (float64(fps) / float64(tps))
	if sequence.partialFrame >= 1.0 {
		sequence.partialFrame = 0
		sequence.currentFrameNumber++
	}
}

// DrawSequence draws the current frame of the given sequence
func DrawSequence(sequence *SequenceNoAudio, screen *ebiten.Image) {
	if sequence.currentFrameNumber >= sequence.totalFrames {
		return
	}
	sequence.drawFrame(screen)
}

// ScaleImage scales an image to x by y
func ScaleImage(x int, y int, i image.Image) image.Image {
	return resize.Resize(uint(x), uint(y), i, resize.Lanczos3)
}

func getAllImages(location string, total int, numZeroes int, prefix string, x int, y int) []*ebiten.Image {
	b := make([]*ebiten.Image, total)
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
			os.Exit(1)
		}

		img, _, err := image.Decode(file)
		img = ScaleImage(x, y, img)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		b[i] = ebiten.NewImageFromImage(img)
	}

	return b
}
