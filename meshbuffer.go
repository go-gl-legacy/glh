// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"github.com/go-gl/gl"
)

// Mesh describes the data offsets for a single mesh inside a mesh buffer.
type Mesh map[string][2]int

// A RenderMode determines how a MeshBuffer should buffer and render mesh data.
type RenderMode uint8

// Known render modes.
const (
	// Classic mode uses manual glBegin/glEnd calls to construct the
	// mesh. This is extremely slow, and mostly only useful for debugging
	// purposes. This implies OpenGL version 'Ye Olde'+.
	RenderClassic = iota

	// Arrays mode uses vertex arrays which involves gl*Pointer calls and
	// directly passing in the vertex data. This is slower than using VBO's,
	// because the data has to be uploaded to the GPU on every render pass.
	// It is useful for older systems where glBufferData is not available.
	// This implies OpenGL version 1.5+.
	RenderArrays

	// Buffered mode uses VBO's. This is the preferred mode for systems
	// where shader support is not present or deemed necessary. This implies
	// OpenGL version 2.1+.
	RenderBuffered
)

// MeshBuffer represents a mesh buffer. It caches and renders vertex data
// for an arbitrary amount of independent meshes.
type MeshBuffer struct {
	meshes []Mesh     // List of mesh descriptors.
	attr   []*Attr    // List of attributes.
	mesh   Mesh       // Internal mesh, representing all data.
	mode   RenderMode // Current render mode.
}

// NewMeshBuffer returns a new mesh buffer object.
//
// The mode parameter defines the mode in which this buffer treats and renders
// vertex data. It should hold one of the predefined RenderMode constants.
//
// The given attributes define the type and size of each vertex component.
// For example:
//
//    mb := NewMeshBuffer(
//        // Render our data using VBO's.
//        glh.RenderBuffered,
//
//        // Indices: 1 unsigned short per index; static data.
//        NewIndexAttr(1, gl.USIGNED_SHORT, gl.STATIC_DRAW),
//
//        // Positions: 3 floats; static data.
//        NewPositionAttr(3, gl.FLOAT, gl.STATIC_DRAW),
//
//        // Colors: 4 floats; changing regularly.
//        NewColorAttr(4, gl.FLOAT, gl.DYNAMIC_DRAW),
//    )
//
// Any mesh data loaded into this buffer through MeshBuffer.Add(), must adhere
// to the format defined by these attributes. THis includes the order in
// which the data is supplied. It must match the order in which the attributes
// are defined here.
func NewMeshBuffer(mode RenderMode, attr ...*Attr) *MeshBuffer {
	switch mode {
	case RenderClassic, RenderArrays, RenderBuffered:
	default:
		panic("Invalid render mode.")
	}

	mb := new(MeshBuffer)
	mb.mode = mode
	mb.attr = attr
	mb.mesh = make(Mesh)

	// All current modes expect the attributes to adhere to some requirements.
	// We require at least a position attribute. Other accepted attributes are
	// for indices, vertex colors, vertex texture coordinates and surface normals.

	pos := mb.find(mbPositionKey)
	if pos == nil || pos.size == 0 {
		panic("The current render mode requires at least a vertex position attribute with size > 0")
	}

	if mb.find(mbIndexKey) == nil {
		mb.attr = append(mb.attr, NewIndexAttr(0, 0, 0))
	}

	if mb.find(mbColorKey) == nil {
		mb.attr = append(mb.attr, NewColorAttr(0, 0, 0))
	}

	if mb.find(mbNormalKey) == nil {
		mb.attr = append(mb.attr, NewNormalAttr(0, 0, 0))
	}

	if mb.find(mbTexCoordKey) == nil {
		mb.attr = append(mb.attr, NewTexCoordAttr(0, 0, 0))
	}

	for _, attr := range mb.attr {
		mb.mesh[attr.name] = [2]int{0, 0}
		attr.init(mode)
	}

	return mb
}

// Release releases all resources for this buffer.
func (mb *MeshBuffer) Release() {
	for i := range mb.attr {
		mb.attr[i].release()
		mb.attr[i] = nil
	}

	mb.mesh = nil
	mb.attr = nil
	mb.meshes = nil
}

// Clear clears the mesh buffer.
func (mb *MeshBuffer) Clear() {
	for i := range mb.attr {
		mb.attr[i].Clear()
	}

	for key := range mb.mesh {
		mb.mesh[key] = [2]int{0, 0}
	}

	mb.meshes = mb.meshes[:0]
}

// find finds an attribute with the given name.
func (mb *MeshBuffer) find(name string) *Attr {
	for _, attr := range mb.attr {
		if attr.name == name {
			return attr
		}
	}
	return nil
}

// Render renders the entire mesh buffer.
// The mode defines one of the symbolic constants like GL_TRIANGLES,
// GL_QUADS, GL_POLYGON, GL_TRIANGLE_STRIP, etc.
func (mb *MeshBuffer) Render(mode gl.GLenum) {
	mb.render(mode, mb.mesh)
}

// RenderMesh renders a single mesh, idenfified by its index.
// The mode defines one of the symbolic constants like GL_TRIANGLES,
// GL_QUADS, GL_POLYGON, GL_TRIANGLE_STRIP, etc.
func (mb *MeshBuffer) RenderMesh(index int, mode gl.GLenum) {
	if index >= 0 && index < len(mb.meshes) {
		mb.render(mode, mb.meshes[index])
	}
}

// render draws the elements defined by the given mesh object.
func (mb *MeshBuffer) render(mode gl.GLenum, m Mesh) {
	pa := mb.find(mbPositionKey)
	ca := mb.find(mbColorKey)
	na := mb.find(mbNormalKey)
	ta := mb.find(mbTexCoordKey)
	ia := mb.find(mbIndexKey)

	switch mb.mode {
	case RenderClassic:
		mb.renderClassic(mode, m, pa, ca, na, ta, ia)
	case RenderArrays:
		mb.renderArrays(mode, m, pa, ca, na, ta, ia)
	case RenderBuffered:
		mb.renderBuffered(mode, m, pa, ca, na, ta, ia)
	}
}

// renderClassic uses manual glBegin/glEnd calls to construct the mesh. This is
// extremely slow, and mostly only useful for debugging purposes.
func (mb *MeshBuffer) renderClassic(mode gl.GLenum, m Mesh, pa, ca, na, ta, ia *Attr) {
	ps, pc := m[mbPositionKey][0], m[mbPositionKey][1]
	cs, cc := m[mbColorKey][0], m[mbColorKey][1]
	ns, nc := m[mbNormalKey][0], m[mbNormalKey][1]
	ts, tc := m[mbTexCoordKey][0], m[mbTexCoordKey][1]
	ic := m[mbIndexKey][1]

	count := pc
	if ic > 0 {
		count = ic
	}

	gl.Begin(mode)

	for i := 0; i < count; i++ {
		idx := i

		if ic > 0 {
			idx = ia.index(i)
		}

		if cc > 0 {
			ca.color(idx + cs)
		}

		if nc > 0 {
			na.normal(idx + ns)
		}

		if tc > 0 {
			ta.texcoord(idx + ts)
		}

		if pc > 0 {
			pa.vertex(idx + ps)
		}
	}

	gl.End()
}

// Arrays mode uses vertex arrays which involves gl*Pointer calls and
// directly passing in the vertex data on every render pass. This is slower
// than using VBO's, because the data has to be uploaded to the GPU on every
// render pass, but it is useful for older systems where glBufferData is
// not available.
func (mb *MeshBuffer) renderArrays(mode gl.GLenum, m Mesh, pa, ca, na, ta, ia *Attr) {
	ps, pc := m[mbPositionKey][0], m[mbPositionKey][1]
	is, ic := m[mbIndexKey][0], m[mbIndexKey][1]
	cc := m[mbColorKey][1]
	nc := m[mbNormalKey][1]
	tc := m[mbTexCoordKey][1]

	gl.PushClientAttrib(gl.CLIENT_VERTEX_ARRAY_BIT)
	defer gl.PopClientAttrib()

	if pc > 0 {
		gl.EnableClientState(gl.VERTEX_ARRAY)
		defer gl.DisableClientState(gl.VERTEX_ARRAY)
		gl.VertexPointer(pa.size, pa.typ, 0, pa.ptr(0))
	}

	if cc > 0 {
		gl.EnableClientState(gl.COLOR_ARRAY)
		defer gl.DisableClientState(gl.COLOR_ARRAY)
		gl.ColorPointer(ca.size, ca.typ, 0, ca.ptr(0))
	}

	if nc > 0 {
		gl.EnableClientState(gl.NORMAL_ARRAY)
		defer gl.DisableClientState(gl.NORMAL_ARRAY)
		gl.NormalPointer(na.typ, 0, na.ptr(0))
	}

	if tc > 0 {
		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
		gl.TexCoordPointer(ta.size, ta.typ, 0, ta.ptr(0))
	}

	if ic > 0 {
		gl.DrawElements(mode, ic, ia.typ, ia.ptr(is*ia.size))
	} else {
		gl.DrawArrays(mode, ps, pc)
	}
}

// renderBuffered uses VBO's. This is the preferred mode for systems
// where shader support is not present or deemed necessary.
func (mb *MeshBuffer) renderBuffered(mode gl.GLenum, m Mesh, pa, ca, na, ta, ia *Attr) {
	ps, pc := m[mbPositionKey][0], m[mbPositionKey][1]
	is, ic := m[mbIndexKey][0], m[mbIndexKey][1]
	cc := m[mbColorKey][1]
	nc := m[mbNormalKey][1]
	tc := m[mbTexCoordKey][1]

	if pc > 0 {
		gl.EnableClientState(gl.VERTEX_ARRAY)
		defer gl.DisableClientState(gl.VERTEX_ARRAY)

		pa.bind()
		if pa.Invalid() {
			pa.buffer()
		}
		gl.VertexPointer(pa.size, pa.typ, 0, uintptr(0))
		pa.unbind()
	}

	if cc > 0 {
		gl.EnableClientState(gl.COLOR_ARRAY)
		defer gl.DisableClientState(gl.COLOR_ARRAY)

		ca.bind()
		if ca.Invalid() {
			ca.buffer()
		}
		gl.ColorPointer(ca.size, ca.typ, 0, uintptr(0))
		ca.unbind()
	}

	if nc > 0 {
		gl.EnableClientState(gl.NORMAL_ARRAY)
		defer gl.DisableClientState(gl.NORMAL_ARRAY)

		na.bind()
		if na.Invalid() {
			na.buffer()
		}
		gl.NormalPointer(na.typ, 0, uintptr(0))
		na.unbind()
	}

	if tc > 0 {
		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)

		ta.bind()
		if ta.Invalid() {
			ta.buffer()
		}
		gl.TexCoordPointer(ta.size, ta.typ, 0, uintptr(0))
		ta.unbind()
	}

	if ic > 0 {
		ia.bind()

		if ia.Invalid() {
			ia.buffer()
		}

		gl.PushClientAttrib(gl.CLIENT_VERTEX_ARRAY_BIT)
		gl.DrawElements(mode, ic, ia.typ, uintptr(is*ia.stride))
		gl.PopClientAttrib()
		ia.unbind()
	} else {
		pa.bind()
		gl.PushClientAttrib(gl.CLIENT_VERTEX_ARRAY_BIT)
		gl.DrawArrays(mode, ps, pc)
		gl.PopClientAttrib()
		pa.unbind()
	}
}

// Add appends new mesh data to the buffer.
//
// The data specified in these lists should match the buffer attributes.
// We expect to receive lists like []float32, []byte in the same order as
// the attributes where supplied to NewMeshBuffer.
//
// Returns an index into the MeshBuffer.Meshes() list.
func (mb *MeshBuffer) Add(argv ...interface{}) int {
	m := make(Mesh)

	for i := 0; i < len(argv) && i < len(mb.attr); i++ {
		attr := mb.attr[i]

		if attr.size == 0 {
			continue
		}

		if argv[i] == nil {
			panic("Invalid data for attribute: " + attr.name)
		}

		start := attr.Len() / attr.size
		count := attr.append(argv[i]) / attr.size

		m[attr.name] = [2]int{start, count}
		mb.mesh[attr.name] = [2]int{0, start + count}
	}

	// Update indices if necessary.
	if index, ok := m[mbIndexKey]; ok {
		pos, ok := m[mbPositionKey]
		if !ok {
			panic("Invalid data for attribute: " + mbPositionKey)
		}

		ia := mb.find(mbIndexKey)
		ia.increment(index[0], float64(pos[0]))
	}

	mb.meshes = append(mb.meshes, m)
	return len(mb.meshes) - 1
}

// Mode returns the render mode for this buffer.
func (mb *MeshBuffer) Mode() RenderMode { return mb.mode }

// Meshes returns a list of meshes in the buffer.
func (mb *MeshBuffer) Meshes() []Mesh { return mb.meshes }

// Positions returns the mesh attribute for position data.
// This is relevant only for render modes other than RenderShader.
func (mb *MeshBuffer) Positions() *Attr { return mb.find(mbPositionKey) }

// Colors returns the mesh attribute for color data.
// This is relevant only for render modes other than RenderShader.
func (mb *MeshBuffer) Colors() *Attr { return mb.find(mbColorKey) }

// Normals returns the mesh attribute for normal data.
// This is relevant only for render modes other than RenderShader.
func (mb *MeshBuffer) Normals() *Attr { return mb.find(mbNormalKey) }

// TexCoords returns the mesh attribute for texture coordinate data.
// This is relevant only for render modes other than RenderShader.
func (mb *MeshBuffer) TexCoords() *Attr { return mb.find(mbTexCoordKey) }

// Indices returns the mesh attribute for index data.
// This is relevant only for render modes other than RenderShader.
func (mb *MeshBuffer) Indices() *Attr { return mb.find(mbIndexKey) }

// Attr returns the attribute for the given name.
func (mb *MeshBuffer) Attr(name string) *Attr { return mb.find(name) }
