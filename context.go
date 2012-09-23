package glhelpers

import (
	"github.com/banthar/gl"
)

// Types implementing Context can be used with `With`
type Context interface {
	Enter()
	Exit()
}

// I like python. Therefore `func With`. Useful to ensure that changes to state
// don't leak further than you would like and to insert implicit error checking
// Example usage:
//     With(Matrix{gl.PROJECTION}, func() { .. operations which modify matrix .. })
//     // Changes to the matrix are undone here
func With(c Context, action func()) {
	defer OpenGLSentinel()()
	c.Enter()
	defer c.Exit()
	action()
}

// Combine multiple contexts into one
func Compound(contexts ...Context) CompoundContextImpl {
	return CompoundContextImpl{contexts}
}

type CompoundContextImpl struct{ contexts []Context }

func (c CompoundContextImpl) Enter() {
	for i := range c.contexts {
		c.contexts[i].Enter()
	}
}
func (c CompoundContextImpl) Exit() {
	for i := range c.contexts {
		c.contexts[len(c.contexts)-i-1].Exit()
	}
}

// A context which preserves the matrix mode, drawing, etc.
type Matrix struct{ Type gl.GLenum }

func (m Matrix) Enter() {
	gl.PushAttrib(gl.TRANSFORM_BIT)
	gl.MatrixMode(m.Type)
	gl.PushMatrix()
}

func (m Matrix) Exit() {
	gl.PopMatrix()
	gl.PopAttrib()
}

// A context which preserves Attrib bits
type Attrib struct{ Bits gl.GLbitfield }

func (a Attrib) Enter() {
	gl.PushAttrib(a.Bits)
}

func (a Attrib) Exit() {
	gl.PopAttrib()
}

// Context which does `gl.Begin` and `gl.End`
// Example:
// 	With(Primitive{gl.LINES}, func() { gl.Vertex2f(0, 0); gl.Vertex2f(1,1) })
type Primitive struct{ Type gl.GLenum }

func (p Primitive) Enter() { gl.Begin(p.Type) }
func (p Primitive) Exit()  { gl.End() }

// Set the GL_PROJECTION matrix to use window co-ordinates and load the identity
// matrix into the GL_MODELVIEW matrix
type WindowCoords struct{}

func (wc WindowCoords) Enter() {
	w, h := GetViewportWH()
	Matrix{gl.PROJECTION}.Enter()
	gl.LoadIdentity()
	gl.Ortho(0, float64(w), float64(h), 0, -1, 1)
	Matrix{gl.MODELVIEW}.Enter()
	gl.LoadIdentity()
}

func (wc WindowCoords) Exit() {
	Matrix{gl.MODELVIEW}.Exit()
	Matrix{gl.PROJECTION}.Exit()
}
