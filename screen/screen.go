package screen

import (
	"goboy2/ppu"
	"image"
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const winWidth, winHeight int32 = 160, 144
const initialScale int32 = 1

type KeyEvent struct {
	Pressed bool
	Key     sdl.Keycode
}

type Screen struct {
	stop   chan struct{}
	render chan *image.RGBA
	input  chan interface{}
}

func Main(mainFn func(s *Screen, input <-chan interface{}, exitChan <-chan struct{})) {
	runtime.LockOSThread()
	runtime.GOMAXPROCS(runtime.NumCPU())
	screen := &Screen{
		stop:   make(chan struct{}),
		render: make(chan *image.RGBA),
		input:  make(chan interface{}),
	}
	wnd, err := sdl.CreateWindow("GoBoy2",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		winWidth*initialScale,
		winHeight*initialScale,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)

	if err != nil {
		log.Fatal(err)
	}
	defer wnd.Destroy()
	wnd.SetMinimumSize(winWidth, winHeight)

	renderer, err := sdl.CreateRenderer(wnd, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Destroy()
	renderer.SetLogicalSize(winWidth, winHeight)
	drawImageOnRenderer(nil, 0, 0, renderer)
	defer close(screen.render)

	go mainFn(screen, screen.input, screen.stop)

	var texture *sdl.Texture
	var dx, dy int32

	for {
		select {
		case _, _ = <-screen.stop:
			return
		case img := <-screen.render:
			if img != nil {
				b := img.Bounds()
				renderer.SetLogicalSize(int32(b.Dx()), int32(b.Dy()))
				texture, dx, dy = imgToTex(img, renderer)
				ppu.FreeScreen(img)
			} else {
				texture = nil
				dx = 0
				dy = 0
			}

			drawImageOnRenderer(texture, dx, dy, renderer)
		default:
			switch ev := sdl.PollEvent(); e := ev.(type) {
			case *sdl.QuitEvent:
				close(screen.stop)
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYUP {
					if e.Keysym.Sym == sdl.K_ESCAPE {
						close(screen.stop)
					} else {
						screen.input <- KeyEvent{false, e.Keysym.Sym}
					}
				} else {
					screen.input <- KeyEvent{true, e.Keysym.Sym}
				}
			default:
				//drawImageOnRenderer(texture, dx, dy, renderer)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
}

func imgToTex(img *image.RGBA, renderer *sdl.Renderer) (tex *sdl.Texture, dx, dy int32) {
	bnds := img.Bounds()

	sdlImg, err := sdl.CreateRGBSurfaceFrom(
		unsafe.Pointer(&img.Pix[0]),
		int32(bnds.Dx()), int32(bnds.Dy()),
		32, img.Stride,
		0, 0, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer sdlImg.Free()
	tex, err = renderer.CreateTextureFromSurface(sdlImg)
	if err != nil {
		log.Fatal(err)
	}
	return tex, int32(bnds.Dx()), int32(bnds.Dy())
}

func drawImageOnRenderer(img *sdl.Texture, dx, dy int32, renderer *sdl.Renderer) {
	renderer.Clear()
	if img != nil {
		renderer.Copy(img,
			&sdl.Rect{W: dx, H: dy},
			&sdl.Rect{W: int32(winWidth), H: int32(winHeight)},
		)
	}
	renderer.Present()
}

func (s *Screen) Stop() {
	close(s.stop)
}

func (s *Screen) GetOutputChannel() chan<- *image.RGBA {
	return s.render
}
