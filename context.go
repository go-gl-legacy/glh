// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"github.com/go-gl/gl"
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

type _enable struct {
	enums []gl.GLenum
}

func (e _enable) Enter() {
	gl.PushAttrib(gl.ENABLE_BIT)
	for _, item := range e.enums {
		gl.Enable(item)
	}
}

func (e _enable) Exit() {
	gl.PopAttrib()
}

func Enable(enums ...gl.GLenum) Context {
	return _enable{enums}
}

type _disable struct {
	enums []gl.GLenum
}

func (e _disable) Enter() {
	gl.PushAttrib(gl.ENABLE_BIT)
	for _, item := range e.enums {
		gl.Disable(item)
	}
}

func (e _disable) Exit() {
	gl.PopAttrib()
}

func Disable(enums ...gl.GLenum) Context {
	return _disable{enums}
}

// Context which does `gl.Begin` and `gl.End`
// Example:
// 	With(Primitive{gl.LINES}, func() { gl.Vertex2f(0, 0); gl.Vertex2f(1,1) })
type Primitive struct{ Type gl.GLenum }

func (p Primitive) Enter() { gl.Begin(p.Type) }
func (p Primitive) Exit()  { gl.End() }

// Set the GL_PROJECTION matrix to use window co-ordinates and load the identity
// matrix into the GL_MODELVIEW matrix
type WindowCoords struct {
	NoReset bool
	Invert  bool
}

func (wc WindowCoords) Enter() {
	w, h := GetViewportWH()
	Matrix{gl.PROJECTION}.Enter()
	if !wc.NoReset {
		gl.LoadIdentity()
	}
	if wc.Invert {
		gl.Ortho(0, float64(w), float64(h), 0, -1, 1)
	} else {
		gl.Ortho(0, float64(w), 0, float64(h), -1, 1)
	}
	Matrix{gl.MODELVIEW}.Enter()
	gl.LoadIdentity()
}

func (wc WindowCoords) Exit() {
	Matrix{gl.MODELVIEW}.Exit()
	Matrix{gl.PROJECTION}.Exit()
}
