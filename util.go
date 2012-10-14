package glhelpers

import (
	"log"

	"github.com/go-gl/gl"
	"github.com/go-gl/glu"
)

// Used as "defer OpenGLSentinel()()" checks the gl error code on call and exit
func OpenGLSentinel() func() {
	check := func() {
		e := gl.GetError()
		if e != gl.NO_ERROR {
			s, err := glu.ErrorString(e)
			if err != nil {
				log.Panic("Invalid error code: ", err)
			}
			log.Panic("Encountered GLError: ", e, " = ", s)
		}
	}
	check()
	return check
}

// Returns w, h of viewport
func GetViewportWH() (int, int) {
	var viewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, viewport[0:3])
	return int(viewport[2]), int(viewport[3])
}

func GetViewportWHD() (float64, float64) {
	w, h := GetViewportWH()
	return float64(w), float64(h)
}

// Returns x, y in window co-ordinates at 0 in the z direction
func WindowToProj(x, y int) (float64, float64) {
	var projmat, modelmat [16]float64
	var viewport [4]int32

	gl.GetDoublev(gl.PROJECTION_MATRIX, projmat[0:15])
	gl.GetDoublev(gl.MODELVIEW_MATRIX, modelmat[0:15])

	gl.GetIntegerv(gl.VIEWPORT, viewport[0:3])
	// Need to convert so that y is at lower left
	y = int(viewport[3]) - y

	px, py, _ := glu.UnProject(float64(x), float64(y), 0,
		&modelmat, &projmat, &viewport)

	return px, py
}

// Returns x, y in window co-ordinates at 0 in the z direction
func ProjToWindow(x, y float64) (float64, float64) {
	var projmat, modelmat [16]float64
	var viewport [4]int32

	gl.GetDoublev(gl.PROJECTION_MATRIX, projmat[0:15])
	gl.GetDoublev(gl.MODELVIEW_MATRIX, modelmat[0:15])
	gl.GetIntegerv(gl.VIEWPORT, viewport[0:3])

	px, py, _ := glu.Project(float64(x), float64(y), 0,
		&modelmat, &projmat, &viewport)

	//return int(px), int(viewport[3]) - int(py)
	return px, float64(viewport[3]) - py
}

// Draws a cross on the screen with known lengths, useful for understanding
// screen dimensions
func DebugLines() {
	gl.MatrixMode(gl.PROJECTION)
	gl.PushMatrix()
	//gl.LoadIdentity()
	//gl.Ortho(-2.1, 6.1, -4, 8, 1, -1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.PushMatrix()
	gl.LoadIdentity()

	gl.LoadIdentity()
	gl.LineWidth(5)
	gl.Color4f(1, 1, 0, 1)
	gl.Begin(gl.LINES)
	gl.Vertex2d(0, -1.6)
	gl.Vertex2d(0, 0.8)
	gl.Vertex2d(-0.8, 0)
	gl.Vertex2d(0.8, 0)
	gl.End()
	gl.PopMatrix()

	gl.MatrixMode(gl.PROJECTION)
	gl.PopMatrix()
	gl.MatrixMode(gl.MODELVIEW)
}

// Emit Vertices of a square with texture co-ordinates which wind anti-clockwise 
func Squarei(x, y, w, h int) {
	u, v, u2, v2 := 0, 1, 1, 0

	gl.TexCoord2i(u, v)
	gl.Vertex2i(x, y)

	gl.TexCoord2i(u2, v)
	gl.Vertex2i(x+w, y)

	gl.TexCoord2i(u2, v2)
	gl.Vertex2i(x+w, y+h)

	gl.TexCoord2i(u, v2)
	gl.Vertex2i(x, y+h)
}

// Draw a Quad with integer co-ordinates (Using Squarei)
func DrawQuadi(x, y, w, h int) {
	With(Primitive{gl.QUADS}, func() {
		Squarei(x, y, w, h)
	})
}

// Same as Squarei, double co-ordinates
func Squared(x, y, w, h float64) {
	u, v, u2, v2 := 0, 1, 1, 0

	gl.TexCoord2i(u, v)
	gl.Vertex2d(x, y)

	gl.TexCoord2i(u2, v)
	gl.Vertex2d(x+w, y)

	gl.TexCoord2i(u2, v2)
	gl.Vertex2d(x+w, y+h)

	gl.TexCoord2i(u, v2)
	gl.Vertex2d(x, y+h)
}

// Same as DrawQuadi, double co-ordinates
func DrawQuadd(x, y, w, h float64) {
	With(Primitive{gl.QUADS}, func() {
		Squared(x, y, w, h)
	})
}
