package video

import (
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/nfnt/resize"
)

const (
	zero = "0"
)

// NewSequence creates a new sequence struct
// screenwidth/height is the rectangle the video is being drawn in; used to scale
// totalframes is the total amount of frames in the video
// prefix is the prefix for the video, used in both the folder and filename ie. prefix video, video/video001.png
// filesystem is the filesystem containing the folder with the png sequence
func NewSequence(prefix string, totalFrames int, screenWidth int, screenHeight int, filesystem embed.FS) *SequenceNoAudio {
	numZeroes := int(math.Log10(float64(totalFrames)))
	frames := getAllImages(totalFrames, numZeroes, prefix, screenWidth, screenHeight, filesystem)

	return &SequenceNoAudio{
		totalFrames:        totalFrames,
		frames:             frames,
		currentFrameNumber: 1,
		lastFrameNumber:    0,
		partialFrame:       0.0,
		currentFrameImage:  frames[0],
		IsFinished:         false,
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
		sequence.IsFinished = true
		return
	}
	sequence.drawFrame(screen)
}

// ScaleImage scales an image to x by y
func ScaleImage(x int, y int, i image.Image) image.Image {
	return resize.Resize(uint(x), uint(y), i, resize.Lanczos3)
}

// generates all of the ebiten images needed for the video
func getAllImages(total int, numZeroes int, prefix string, x int, y int, filesystem embed.FS) []*ebiten.Image {
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
		filename = prefix + "/" + prefix + num + strconv.Itoa(i) + ".png"
		file, err := filesystem.Open(filename)
		if err != nil {
			log.Fatal(err)
		}

		img, _, err := image.Decode(file)
		img = ScaleImage(x, y, img)
		if err != nil {
			log.Fatal(err)
		}
		b[i] = ebiten.NewImageFromImage(img)
		file.Close()
		fmt.Println("file " + strconv.Itoa(i) + " done")
	}

	return b
}
