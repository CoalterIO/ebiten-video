package video

import (
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	zero            = "0"
	frameBufferSize = 1024
)

// NewSequenceFromFolder creates a new sequence struct from a folder name
// screenwidth/height is the rectangle the video is being drawn in; used to scale
// totalframes is the total amount of frames in the video
// prefix is the prefix for the video, used in the filename ie. prefix video, .../video001.png
// location is the path to the folder containing the png sequence
func NewSequenceFromFolder(prefix string, location string, totalFrames int, screenWidth int, screenHeight int) (*SequenceNoAudio, error) {
	if x, err := exists(location); err != nil {
		return nil, err
	} else if !x {
		return nil, &DirectoryDoesNotExistError{Path: location}
	}
	numZeroes := int(math.Log10(float64(totalFrames)))
	frameBuffer := getAllImagesFromFolder(totalFrames, numZeroes, prefix, location, screenWidth, screenHeight)

	return &SequenceNoAudio{
		totalFrames:        totalFrames,
		currentFrameNumber: 1,
		partialFrame:       0.0,
		currentFrameImage:  <-frameBuffer,
		IsFinished:         false,
		frames:             frameBuffer,
	}, nil
}

// NewSequenceFromFS creates a new sequence struct from an embedded filesystem
// screenwidth/height is the rectangle the video is being drawn in; used to scale
// totalframes is the total amount of frames in the video
// prefix is the prefix for the video, used in the filename ie. prefix video, .../video001.png
// filesystem is the embedded embed.FS
func NewSequenceFromFS(prefix string, filesystem embed.FS, totalFrames int, screenWidth int, screenHeight int) (*SequenceNoAudio, error) {
	numZeroes := int(math.Log10(float64(totalFrames)))
	frameBuffer := getAllImagesFromFS(totalFrames, numZeroes, prefix, filesystem, screenWidth, screenHeight)

	return &SequenceNoAudio{
		totalFrames:        totalFrames,
		currentFrameNumber: 1,
		partialFrame:       0.0,
		currentFrameImage:  <-frameBuffer,
		IsFinished:         false,
		frames:             frameBuffer,
	}, nil
}

// UpdateSequence updates the info in the png sequence so you can draw it with DrawSequence
func UpdateSequence(sequence *SequenceNoAudio, fps int, tps int) {
	if sequence.currentFrameNumber >= sequence.totalFrames {
		return
	}
	sequence.partialFrame += (float64(fps) / float64(tps))
	if sequence.partialFrame >= 1.0 {
		sequence.currentFrameImage = <-sequence.frames
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
	if sequence.currentFrameImage != nil {
		sequence.drawFrame(screen)
	}
}

func scaleImage(xin, xout, yin, yout int) *ebiten.DrawImageOptions {
	var (
		xScale = float64(xout) / float64(xin)
		yScale = float64(yout) / float64(yin)
		o      = &ebiten.DrawImageOptions{}
	)
	o.GeoM.Scale(xScale, yScale)
	return o
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// generates all of the ebiten images needed for the video
func getAllImagesFromFolder(total int, numZeroes int, prefix string, location string, x int, y int) <-chan *imageWithOptions {
	frameBuffer := make(chan *imageWithOptions, frameBufferSize)
	numZeroes = int(math.Floor(float64(numZeroes)) + 1)
	var filename, num string
	var (
		z   int
		err error
		img *ebiten.Image
		o   *ebiten.DrawImageOptions
	)

	go func() {
		for i := 0; i < total; i++ {
			if i != 0 {
				z = numZeroes - int(math.Floor(float64(math.Log10(float64(i)))+1))
			} else {
				z = numZeroes - 1
			}
			num = strings.Repeat(zero, z)
			filename = location + "/" + prefix + num + strconv.Itoa(i) + ".png"

			img, _, err = ebitenutil.NewImageFromFile(filename)
			if err != nil {
				log.Fatal(err)
			}
			if img.Bounds().Dx() != x || img.Bounds().Dy() != y {
				o = scaleImage(img.Bounds().Dx(), x, img.Bounds().Dy(), y)
			} else {
				o = &ebiten.DrawImageOptions{}
			}
			s := &imageWithOptions{
				i: img,
				o: o,
			}
			frameBuffer <- s
			fmt.Println("file " + strconv.Itoa(i) + " done")
		}

		go func() {
			for {
				if len(frameBuffer) <= 0 {
					close(frameBuffer)
					break
				}
			}
		}()
	}()

	return frameBuffer
}

func getAllImagesFromFS(total int, numZeroes int, prefix string, filesystem embed.FS, x int, y int) <-chan *imageWithOptions {
	frameBuffer := make(chan *imageWithOptions, frameBufferSize)
	numZeroes = int(math.Floor(float64(numZeroes)) + 1)
	var filename, num string
	var (
		z    int
		err  error
		i    image.Image
		img  *ebiten.Image
		o    *ebiten.DrawImageOptions
		file fs.File
	)

	go func() {
		for j := 0; j < total; j++ {
			if j != 0 {
				z = numZeroes - int(math.Floor(float64(math.Log10(float64(j)))+1))
			} else {
				z = numZeroes - 1
			}
			num = strings.Repeat(zero, z)

			filename = prefix + "/" + prefix + num + strconv.Itoa(j) + ".png"
			file, err = filesystem.Open(filename)
			defer file.Close()
			if err != nil {
				log.Fatal(err)
			}
			i, _, err = image.Decode(file)
			if err != nil {
				log.Fatal(err)
			}

			img = ebiten.NewImageFromImage(i)
			if img.Bounds().Dx() != x || img.Bounds().Dy() != y {
				o = scaleImage(img.Bounds().Dx(), x, img.Bounds().Dy(), y)
			} else {
				o = &ebiten.DrawImageOptions{}
			}
			s := &imageWithOptions{
				i: img,
				o: o,
			}
			frameBuffer <- s
			fmt.Println("file " + strconv.Itoa(j) + " done")
		}

		go func() {
			for {
				if len(frameBuffer) <= 0 {
					close(frameBuffer)
					break
				}
			}
		}()
	}()

	return frameBuffer
}
