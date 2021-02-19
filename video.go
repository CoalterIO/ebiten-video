package main

import (
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/giorgisio/goav/avcodec"
	"github.com/giorgisio/goav/avformat"
	"github.com/giorgisio/goav/avutil"
	"github.com/giorgisio/goav/swscale"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	FrameBufferSize = 1024
)

var (
	WindowWidth  int
	WindowHeight int
	view         *ebiten.Image
	frameBuffer  <-chan *ebiten.Image
)

func getFrameRGBA(frame *avutil.Frame, width, height int) *ebiten.Image {
	pix := []byte{}

	for y := 0; y < height; y++ {
		data0 := avutil.Data(frame)[0]
		buf := make([]byte, width*4)
		startPos := uintptr(unsafe.Pointer(data0)) +
			uintptr(y)*uintptr(avutil.Linesize(frame)[0])

		for i := 0; i < width*4; i++ {
			element := *(*uint8)(unsafe.Pointer(startPos + uintptr(i)))
			buf[i] = element
		}

		pix = append(pix, buf...)
	}
	s := ebiten.NewImage(width, height)
	s.ReplacePixels(pix)
	return s
}

func readVideoFrames(videoPath string) <-chan *ebiten.Image {
	// Create a frame buffer.
	frameBuffer := make(chan *ebiten.Image, FrameBufferSize)

	go func() {
		// Open a video file.
		pFormatContext := avformat.AvformatAllocContext()

		if avformat.AvformatOpenInput(&pFormatContext, videoPath, nil, nil) != 0 {
			fmt.Printf("Unable to open file %s\n", videoPath)
			os.Exit(1)
		}

		// Retrieve the stream information.
		if pFormatContext.AvformatFindStreamInfo(nil) < 0 {
			fmt.Println("Couldn't find stream information")
			os.Exit(1)
		}

		// Dump information about the video to stderr.
		pFormatContext.AvDumpFormat(0, videoPath, 0)

		// Find the first video stream
		for i := 0; i < int(pFormatContext.NbStreams()); i++ {
			switch pFormatContext.Streams()[i].
				CodecParameters().AvCodecGetType() {
			case avformat.AVMEDIA_TYPE_VIDEO:

				// Get a pointer to the codec context for the video stream
				pCodecCtxOrig := pFormatContext.Streams()[i].Codec()
				// Find the decoder for the video stream
				pCodec := avcodec.AvcodecFindDecoder(avcodec.
					CodecId(pCodecCtxOrig.GetCodecId()))

				if pCodec == nil {
					fmt.Println("Unsupported codec!")
					os.Exit(1)
				}

				// Copy context
				pCodecCtx := pCodec.AvcodecAllocContext3()

				if pCodecCtx.AvcodecCopyContext((*avcodec.
					Context)(unsafe.Pointer(pCodecCtxOrig))) != 0 {
					fmt.Println("Couldn't copy codec context")
					os.Exit(1)
				}

				// Open codec
				if pCodecCtx.AvcodecOpen2(pCodec, nil) < 0 {
					fmt.Println("Could not open codec")
					os.Exit(1)
				}

				// Allocate video frame
				pFrame := avutil.AvFrameAlloc()

				// Allocate an AVFrame structure
				pFrameRGB := avutil.AvFrameAlloc()

				if pFrameRGB == nil {
					fmt.Println("Unable to allocate RGB Frame")
					os.Exit(1)
				}

				// Determine required buffer size and allocate buffer
				numBytes := uintptr(avcodec.AvpictureGetSize(avcodec.PixelFormat(avcodec.AV_PIX_FMT_RGBA), pCodecCtx.Width(), pCodecCtx.Height()))
				buffer := avutil.AvMalloc(numBytes)

				// Assign appropriate parts of buffer to image planes in pFrameRGB
				// Note that pFrameRGB is an AVFrame, but AVFrame is a superset
				// of AVPicture
				avp := (*avcodec.Picture)(unsafe.Pointer(pFrameRGB))
				avp.AvpictureFill((*uint8)(buffer), avcodec.PixelFormat(avcodec.AV_PIX_FMT_RGBA), pCodecCtx.Width(), pCodecCtx.Height())

				// initialize SWS context for software scaling
				swsCtx := swscale.SwsGetcontext(
					pCodecCtx.Width(),
					pCodecCtx.Height(),
					(swscale.PixelFormat)(pCodecCtx.PixFmt()),
					pCodecCtx.Width(),
					pCodecCtx.Height(),
					swscale.PixelFormat(avcodec.AV_PIX_FMT_RGBA),
					2, // SWS.BILINEAR
					nil,
					nil,
					nil,
				)

				// Read frames and save first five frames to disk
				packet := avcodec.AvPacketAlloc()

				for pFormatContext.AvReadFrame(packet) >= 0 {
					// Is this a packet from the video stream?
					if packet.StreamIndex() == i {
						// Decode video frame
						response := pCodecCtx.AvcodecSendPacket(packet)

						if response < 0 {
							fmt.Printf("Error while sending a packet to the decoder: %s\n",
								avutil.ErrorFromCode(response))
						}

						for response >= 0 {
							response = pCodecCtx.AvcodecReceiveFrame(
								(*avcodec.Frame)(unsafe.Pointer(pFrame)))

							if response == avutil.AvErrorEAGAIN ||
								response == avutil.AvErrorEOF {
								break
							} else if response < 0 {
								//fmt.Printf("Error while receiving a frame from the decoder: %s\n",
								//avutil.ErrorFromCode(response))

								//return
							}

							// Convert the image from its native format to RGB
							swscale.SwsScale2(swsCtx, avutil.Data(pFrame),
								avutil.Linesize(pFrame), 0, pCodecCtx.Height(),
								avutil.Data(pFrameRGB), avutil.Linesize(pFrameRGB))

							// Save the frame to the frame buffer.
							frame := getFrameRGBA(pFrameRGB, pCodecCtx.Width(), pCodecCtx.Height())
							frameBuffer <- frame
						}
					}

					// Free the packet that was allocated by av_read_frame
					packet.AvFreePacket()
				}

				go func() {
					for {
						if len(frameBuffer) <= 0 {
							close(frameBuffer)
							break
						}
					}
				}()

				// Free the RGB image
				avutil.AvFree(buffer)
				avutil.AvFrameFree(pFrameRGB)

				// Free the YUV frame
				avutil.AvFrameFree(pFrame)

				// Close the codecs
				pCodecCtx.AvcodecClose()
				(*avcodec.Context)(unsafe.Pointer(pCodecCtxOrig)).AvcodecClose()

				// Close the video file
				pFormatContext.AvformatCloseInput()

				// Stop after saving frames of first video straem
				break

			default:
				fmt.Println("Didn't find a video stream")
				os.Exit(1)
			}
		}
	}()

	return frameBuffer
}

type Game struct{}

func (g *Game) Update() error {
	select {
	case frame, ok := <-frameBuffer:
		if !ok {
			os.Exit(0)
		}

		if frame != nil {
			view = frame
		}

	default:
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(view, &ebiten.DrawImageOptions{})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return WindowWidth, WindowHeight
}

func main() {
	WindowWidth = 1280
	WindowHeight = 720

	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetMaxTPS(60)

	view = ebiten.NewImage(WindowWidth, WindowHeight)
	frameBuffer = readVideoFrames(os.Args[1])
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
