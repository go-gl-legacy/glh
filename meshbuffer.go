// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"github.com/go-gl/gl"
)

// Mesh describes the data offsets for a single
// mesh inside a mesh buffer.
type Mesh struct {
	PS int // Start index for position data.
	PC int // Number of positions.
	CS int // Start index for color data.
	CC int // Number of colors.
	NS int // Start index for normal data.
	NC int // Number of normals.
	TS int // Start index for texture coordinate data.
	TC int // Number of texture coordinates.
	IS int // Start index for index data.
	IC int // Number of indices.
}

// MeshBuffer represents a mesh buffer.
// It can cache and render vertices, colors, texture coords and normals
// for an arbitrary amount of independent meshes.
type MeshBuffer struct {
	meshes []*Mesh    // List of mesh descriptors.
	attr   []MeshAttr // Attribute list.
}

// NewMeshBuffer returns a new mesh buffer object.
//
// The given attributes define the type and size of each vertex component.
// Setting an attribute to nil, means the component is omitted and no data
// is generated for it. This allowing us to limit the buffer size to the
// space we actually require. For example:
// 
//    mb := NewMeshBuffer(
//        // Indices: 1 unsigned int per index, static data.
//        NewUint32Attr(1, gl.STATIC_DRAW),
//        
//        // Positions: 3 floats, static data.
//        NewFloat32Attr(3, gl.STATIC_DRAW),
//        
//        // Colors: 4 floats, changing regularly.
//        NewFloat32Attr(4, gl.DYNAMIC_DRAW),
//        
//        nil, // No vertex normals.
//        nil, // No texture coords.
//    )
//
// Any meshes loaded into this buffer through MeshBuffer.Add(), must adhere
// to the format defined by these attributes.
func NewMeshBuffer(index, position, color, normal, texture MeshAttr) *MeshBuffer {
	if index == nil {
		index = NewNullAttr()
	}

	index.SetTarget(gl.ELEMENT_ARRAY_BUFFER)

	if position == nil {
		position = NewNullAttr()
	}

	if color == nil {
		color = NewNullAttr()
	}

	if normal == nil {
		normal = NewNullAttr()
	}

	if texture == nil {
		texture = NewNullAttr()
	}

	mb := new(MeshBuffer)

	// Do not change the order of these.
	mb.attr = []MeshAttr{position, color, normal, texture, index}
	return mb
}

// Release releases all resources for this buffer.
func (mb *MeshBuffer) Release() {
	for i := range mb.attr {
		mb.attr[i].Release()
		mb.attr[i] = nil
	}

	mb.attr = nil
	mb.meshes = nil
}

// Clear clears the mesh buffer.
func (mb *MeshBuffer) Clear() {
	for _, a := range mb.attr {
		a.Clear()
	}

	mb.meshes = mb.meshes[:0]
}

// Render renders the entire mesh buffer.
// The mode defines one of the symbolic constants like GL_TRIANGLES,
// GL_QUADS, GL_POLYGON, GL_TRIANGLE_STRIP, etc.
func (mb *MeshBuffer) Render(mode gl.GLenum) {
	pc := mb.attr[0].Len()
	cc := mb.attr[1].Len()
	nc := mb.attr[2].Len()
	tc := mb.attr[3].Len()
	ic := mb.attr[4].Len()

	mb.render(mode, 0, pc, 0, cc, 0, nc, 0, tc, 0, ic)
}

// RenderMesh renders a single mesh, idenfified by its index.
// The mode defines one of the symbolic constants like GL_TRIANGLES,
// GL_QUADS, GL_POLYGON, GL_TRIANGLE_STRIP, etc.
func (mb *MeshBuffer) RenderMesh(index int, mode gl.GLenum) {
	if index >= 0 && index < len(mb.meshes) {
		m := mb.meshes[index]
		mb.render(mode, m.PS, m.PC, m.CS, m.CC, m.NS, m.NC, m.TS, m.TC, m.IS, m.IC)
	}
}

// render draws the elements defined by the given start and count values for
// [v]ertices, [c]olors, [t]exture coords, [n]ormals and [i]ndices.
func (mb *MeshBuffer) render(mode gl.GLenum, ps, pc, cs, cc, ns, nc, ts, tc, is, ic int) {
	/// Re-commit data if necessary.
	for _, attr := range mb.attr {
		if attr.Dirty() {
			mb.commit()
			break
		}
	}

	gl.PushClientAttrib(gl.CLIENT_VERTEX_ARRAY_BIT)
	defer gl.PopClientAttrib()

	gl.EnableClientState(gl.VERTEX_ARRAY)
	defer gl.DisableClientState(gl.VERTEX_ARRAY)

	if cc > 0 {
		gl.EnableClientState(gl.COLOR_ARRAY)
		defer gl.DisableClientState(gl.COLOR_ARRAY)
	}

	if tc > 0 {
		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
	}

	if nc > 0 {
		gl.EnableClientState(gl.NORMAL_ARRAY)
		defer gl.DisableClientState(gl.NORMAL_ARRAY)
	}

	if ic > 0 {
		attr := mb.attr[4] // index attribute.
		attr.Bind()

		gl.DrawElements(mode, ic, attr.Type(), uintptr(is*attr.Stride()))

		attr.Unbind()
	} else {
		attr := mb.attr[0] // position attribute.
		attr.Bind()

		gl.DrawArrays(mode, ps, pc)

		attr.Unbind()
	}
}

// commit pushes the buffer data to the GPU where necessary.
func (mb *MeshBuffer) commit() {
	for i, attr := range mb.attr {
		if !attr.Dirty() || attr.Len() == 0 {
			continue
		}

		attr.Bind()
		attr.Buffer()

		switch i {
		case 0:
			gl.VertexPointer(attr.Size(), attr.Type(), 0, uintptr(0))
		case 1:
			gl.ColorPointer(attr.Size(), attr.Type(), 0, uintptr(0))
		case 2:
			gl.NormalPointer(attr.Type(), 0, uintptr(0))
		case 3:
			gl.TexCoordPointer(attr.Size(), attr.Type(), 0, uintptr(0))
		case 4:
			gl.IndexPointer(attr.Type(), 0, uintptr(0))
		}

		attr.Unbind()
	}

	gl.Finish()
}

// Add appends new mesh data to the buffer.
// The data specified in these lists, should match the buffer attributes.
// We expect to receive lists like []float32, []byte, or nil.
//
// Returns an index into the MeshBuffer.Meshes() list.
func (mb *MeshBuffer) Add(indices, positions, colors, normals, textures interface{}) int {
	pa := mb.attr[0]
	ca := mb.attr[1]
	na := mb.attr[2]
	ta := mb.attr[3]
	ia := mb.attr[4]

	m := new(Mesh)

	if pa.Size() > 0 {
		if positions == nil {
			panic("Missing index list")
		}

		m.PS = pa.Len() / pa.Size()
		m.PC = pa.Append(positions) / pa.Size()
	}

	if ca.Size() > 0 {
		if colors == nil {
			panic("Missing color list")
		}

		m.CC = ca.Len() / ca.Size()
		m.CC = ca.Append(colors) / ca.Size()
	}

	if na.Size() > 0 {
		if normals == nil {
			panic("Missing normal list")
		}

		m.NS = na.Len() / na.Size()
		m.NC = na.Append(normals) / na.Size()
	}

	if ta.Size() > 0 {
		if textures == nil {
			panic("Missing texture coordinate list")
		}

		m.TS = ta.Len() / ta.Size()
		m.TC = ta.Append(textures) / ta.Size()
	}

	if ia.Size() > 0 {
		if indices == nil {
			panic("Missing index list")
		}

		m.IS = ia.Len() / ia.Size()
		m.IC = ia.Append(indices) / ia.Size()
		ia.Increment(m.IS, m.PS)
	}

	mb.meshes = append(mb.meshes, m)
	return len(mb.meshes) - 1
}

// Meshes returns a list of meshes in the buffer.
func (mb *MeshBuffer) Meshes() []*Mesh { return mb.meshes }

// Position returns the mesh attribute for position data.
func (mb *MeshBuffer) Position() MeshAttr { return mb.attr[0] }

// Colors returns the mesh attribute for color data.
func (mb *MeshBuffer) Colors() MeshAttr { return mb.attr[1] }

// Normals returns the mesh attribute for normal data.
func (mb *MeshBuffer) Normals() MeshAttr { return mb.attr[2] }

// Textures returns the mesh attribute for texture coordinate data.
func (mb *MeshBuffer) Textures() MeshAttr { return mb.attr[3] }

// Indices returns the mesh attribute for index data.
func (mb *MeshBuffer) Indices() MeshAttr { return mb.attr[4] }
