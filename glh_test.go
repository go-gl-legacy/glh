package glhelpers

import (
	"testing"

	"github.com/go-gl/gl"
	"github.com/go-gl/testutils"
)

// Draw a test pattern
func TestWindowCoords(t *testing.T) {
	gltest.OnTheMainThread(func() {
		gltest.SetWindowSize(40, 40)

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
