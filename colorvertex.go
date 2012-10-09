package glhelpers

import (
	"unsafe"

	"github.com/go-gl/gl"
)

type Vertex struct{ X, Y float32 }
type Color struct{ R, G, B, A uint8 }
type ColorVertex struct {
	Color
	Vertex
}
type ColorVertices []ColorVertex

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
