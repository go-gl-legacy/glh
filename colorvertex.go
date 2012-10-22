// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"image/color"
	"unsafe"

	"github.com/go-gl/gl"
)

type Vertex struct{ X, Y float32 }

type ColorVertex struct {
	color.RGBA
	Vertex
}
type ColorVertices []ColorVertex

func MkRGBA(c color.Color) color.RGBA {
	if rgba, ok := c.(color.RGBA); ok {
		return rgba
	}
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r / 0xff), uint8(g / 0xff), uint8(b / 0xff), uint8(a / 0xff)}
}

func (vcs ColorVertices) Draw(primitives gl.GLenum) {
	if len(vcs) < 1 {
		return
	}

	gl.PushClientAttrib(0xFFFFFFFF) //gl.CLIENT_ALL_ATTRIB_BITS)
	defer gl.PopClientAttrib()

	gl.InterleavedArrays(gl.C4UB_V2F, 0, unsafe.Pointer(&vcs[0]))
	gl.DrawArrays(primitives, 0, len(vcs))
}

func (vcs ColorVertices) DrawPartial(i, N int64) {
	if len(vcs) < 1 {
		return
	}
	if i+N > int64(len(vcs)) {
		i = int64(len(vcs)) - N
	}
	if i < 0 {
		i = 0
	}
	if N < 1 {
		return
	}

	if i+N > int64(len(vcs)) {
		N = int64(len(vcs)) - i
	}

	gl.PushClientAttrib(0xFFFFFFFF)
	defer gl.PopClientAttrib()

	gl.InterleavedArrays(gl.C4UB_V2F, 0, unsafe.Pointer(&vcs[0]))
	gl.DrawArrays(gl.POINTS, int(i), int(N))
}

func (vcs *ColorVertices) Add(v ColorVertex) {
	*vcs = append(*vcs, v)
}
