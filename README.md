# ebiten-video
a proof-of-concept video player for ebiten that DOESN'T require ffmpeg

# information
this library takes a png sequence (generated from something like Adobe After Effects/Premier) and outputs a video-like sequence to an ebiten screen

currently it only works with png sequences, and unless ebiten is changed in some way to allow videos, it will only ever work with png sequences

this library requires go 1.16 (for embed)

# instructions
-first, make sure you have your png sequence in a folder where the folder name is the same as the png prefix
ie. video/video001.png

-secondly, add the local path to the png sequence (ie. video/video001.png)
```go
const (
	prefix = "video"
)
```

-ALTERNATIVELY: embed the folder containing the png sequence as a filesystem
```go
//embed:go video/*
var filesystem embed.FS
```

-thirdly, initialize the sequence
```go
const (
	x = 1280 //screen width
	y = 720 //screen height
	prefix = "video"
	totalFrames = 100
)

//embed:go video/*
var filesystem embed.FS

func main() {
	ebiten.SetWindowSize(x, y)
	ebiten.SetWindowTitle("Video Test")
	ebiten.SetMaxTPS(60)
	ebiten.SetFullscreen(false)
	go initializeVideo()
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

func initializeVideo() {
	sequence, err = video.NewSequenceFromFolder(prefix, location, totalFrames, x, y)
	// Or, if using the embedded filesystem
	sequenceFS, err = video.NewSequenceFromFS(prefix, filesystem, totalFrames, x, y)
	// handle error...
}
```

-lastly, you can call UpdateSequence and DrawSequence in the Update and Draw function
```go
func (g *Game) Update() error {
	if sequence != nil {
		video.UpdateSequence(sequence, desiredVideoFps, yourTickRate)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if sequence != nil {
		video.DrawSequence(sequence, screen)
	} else {
		// now loading...
	}
}
```

you can run the initialization as a goroutine if you need to do something else while the video is being loaded

# TODO
sound

~~video scaling~~ done

