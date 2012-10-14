package glhelpers

import (
	"log"
	"runtime"
	"testing"

	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
)

var main_thread_work = make(chan func())
var main_thread_work_done = make(chan bool)

func OnTheMainThread(setup, after func()) {
	main_thread_work <- setup
	main_thread_work <- after
	<-main_thread_work_done
}

func init() {
	go func() {
		runtime.LockOSThread()

		if err := glfw.Init(); err != nil {
			log.Panic("glfw Error:", err)
		}

		w, h := 100, 100
		err := glfw.OpenWindow(w, h, 0, 0, 0, 0, 0, 0, glfw.Windowed)
		if err != nil {
			log.Panic("Error:", err)
		}

		if gl.Init() != 0 {
			log.Panic("gl error")
		}

		glfw.SetWindowSizeCallback(Reshape)
		glfw.SwapBuffers()

		for {
			(<-main_thread_work)()
			glfw.SwapBuffers()
			(<-main_thread_work)()
			main_thread_work_done <- true
		}

	}()
}

func SetWindowSize(width, height int) {
	glfw.SetWindowSize(width, height)
	// Need to wait for the reshape event, otherwise it happens at an arbitrary
	// point in the future (some unknown number of SwapBuffers())
	glfw.WaitEvents()
}

func Reshape(width, height int) {
	gl.Viewport(0, 0, width, height)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(-1, 1, -1, 1, -1, 1)
	//gl.Ortho(-2.1, 6.1, -2.25*2, 2.1*2, -1, 1) // Y debug

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

// Draw a test pattern
func TestWindowCoords(t *testing.T) {
	OnTheMainThread(func() {
		SetWindowSize(40, 40)

		w, h := GetViewportWH()
		With(WindowCoords{}, func() {
			// So that we draw in the middle of the pixel
			gl.Translated(0.5, 0.5, 0)

			stride := 1
			internal_n := 4
			for b := 0; b < w/2-internal_n*stride; b += stride {
				if b/stride%2 == 0 {
					gl.Color4f(1, 1, 1, 1)
				} else {
					gl.Color4f(1, 0, 0, 1)
				}
				With(Primitive{gl.LINE_LOOP}, func() {
					gl.Vertex2i(b, b)
					gl.Vertex2i(w-b, b)
					gl.Vertex2i(w-b, h-b)
					gl.Vertex2i(b, h-b)
				})
			}

			// Central white, green, blue checked pattern
			gl.PointSize(2)
			With(Primitive{gl.POINTS}, func() {
				gl.Color4f(1, 1, 1, 1)
				gl.Vertex2i(w/2-2, h/2-2)
				gl.Vertex2i(w/2+2, h/2+2)

				gl.Color4f(0, 1, 0, 1)
				gl.Vertex2i(w/2+2, h/2-2)
				gl.Vertex2i(w/2-2, h/2+2)

				gl.Color4f(1, 1, 1, 1)
				gl.Vertex2i(w/2, h/2)
			})

			// Blue horizontal line to show
			With(Primitive{gl.LINE_LOOP}, func() {
				gl.Color4f(0, 0, 1, 1)
				gl.Vertex2i(0, h/2-4)
				gl.Vertex2i(w, h/2-4)

				gl.Vertex2i(w/2-4, 0)
				gl.Vertex2i(w/2-4, h)
			})

			// Remove top left and top right pixel
			gl.PointSize(1)
			gl.Color4f(0, 0, 0, 1)
			With(Primitive{gl.POINTS}, func() {
				gl.Vertex2i(0, 0)
				gl.Vertex2i(w, 0)
			})
		})
	}, func() {
		CaptureToPng("TestWindowCoords.png")
	})
}
