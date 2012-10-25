// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"github.com/go-gl/gl"
)

// Mesh describes data for a single mesh object.
type Mesh struct {
	Colors   [][4]float32
	Vertices [][3]float32
	Textures [][3]float32
	Normals  [][3]float32
	Indices  []uint
}

// MeshDescriptor describes the data offsets for a single
// mesh inside a mesh buffer.
type MeshDescriptor struct {
	VertexStart  int // Start index for vertex data.
	VertexCount  int // Number of vertices.
	TextureStart int // Start index for texture coordinate data.
	TextureCount int // Number of texture coordinates.
	NormalStart  int // Start index for normal data.
	NormalCount  int // Number of normals.
	ColorStart   int // Start index for color data.
	ColorCount   int // Number of colors.
	IndexStart   int // Start index for index data.
	IndexCount   int // Number of indices.
}

// Known mesh buffer states.
const (
	mbClean = iota
	mbDirty
	mbFrozen
)

// These define the size of each component, in bytes.
var (
	mbVertexStride  = int(Sizeof(gl.FLOAT)) * 3
	mbColorStride   = int(Sizeof(gl.FLOAT)) * 4
	mbNormalStride  = int(Sizeof(gl.FLOAT)) * 3
	mbTextureStride = int(Sizeof(gl.FLOAT)) * 3
	mbIndexStride   = int(Sizeof(gl.UNSIGNED_INT))
)

// MeshBuffer represents a mesh buffer.
// It can cache and render vertices, colors, texture coords and normals
// for an arbitrary amount of independent meshes.
type MeshBuffer struct {
	meshes   []*MeshDescriptor // Mesh descriptor.
	vertices [][3]float32      // List of vertices.
	colors   [][4]float32      // List of colors.
	normals  [][3]float32      // List of normals.
	textures [][3]float32      // List of texture coords.
	indices  []uint            // List of indices.

	vertexId  gl.Buffer // Vertex buffer identity.
	colorId   gl.Buffer // Color buffer identity.
	textureId gl.Buffer // Texture buffer identity.
	normalId  gl.Buffer // Normal buffer identity.
	indexId   gl.Buffer // Index buffer identity.

	gpuVertexSize  int // Current size of the vertex buffer in GPU.
	gpuColorSize   int // Current size of the color buffer in GPU.
	gpuTextureSize int // Current size of the texture coord buffer in GPU.
	gpuNormalSize  int // Current size of the normal buffer in GPU.
	gpuIndexSize   int // Current size of the indices buffer in GPU.

	state int       // Mesh buffer state.
	usage gl.GLenum // Usage flag. GL_DYNAMIC, STATIC_DRAW. GL_STREAM_DRAW, etc/
}

// NewMeshBuffer returns a new mesh buffer object.
func NewMeshBuffer() *MeshBuffer {
	mb := new(MeshBuffer)
	mb.state = mbDirty
	mb.usage = gl.STATIC_DRAW
	mb.vertexId = gl.GenBuffer()
	mb.colorId = gl.GenBuffer()
	mb.textureId = gl.GenBuffer()
	mb.normalId = gl.GenBuffer()
	mb.indexId = gl.GenBuffer()
	return mb
}

// Release releases all resources for this buffer.
func (mb *MeshBuffer) Release() {
	mb.vertexId.Delete()
	mb.indexId.Delete()
	mb.vertexId.Delete()
	mb.colorId.Delete()
	mb.textureId.Delete()
	mb.normalId.Delete()
	mb.indexId.Delete()
	mb.Clear()
}

// Clear clears the mesh buffer.
func (mb *MeshBuffer) Clear() {
	mb.state = mbFrozen
	mb.vertices = nil
	mb.colors = nil
	mb.textures = nil
	mb.normals = nil
	mb.indices = nil
	mb.meshes = nil
	mb.gpuVertexSize = 0
	mb.gpuColorSize = 0
	mb.gpuTextureSize = 0
	mb.gpuNormalSize = 0
	mb.gpuIndexSize = 0
	mb.state = mbDirty
}

// Render renders the entire mesh buffer.
// The mode defines one of the symbolic constants like GL_TRIANGLES,
// GL_QUADS, GL_POLYGON, GL_TRIANGLE_STRIP, etc.
func (mb *MeshBuffer) Render(mode gl.GLenum) {
	if len(mb.meshes) == 0 {
		return
	}

	if mb.state != mbClean {
		mb.Commit()
	}

	gl.PushClientAttrib(gl.CLIENT_VERTEX_ARRAY_BIT)
	defer gl.PopClientAttrib()

	if len(mb.vertices) > 0 {
		gl.EnableClientState(gl.VERTEX_ARRAY)
		defer gl.DisableClientState(gl.VERTEX_ARRAY)
	}

	if len(mb.colors) > 0 {
		gl.EnableClientState(gl.COLOR_ARRAY)
		defer gl.DisableClientState(gl.COLOR_ARRAY)
	}

	if len(mb.textures) > 0 {
		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
	}

	if len(mb.normals) > 0 {
		gl.EnableClientState(gl.NORMAL_ARRAY)
		defer gl.DisableClientState(gl.NORMAL_ARRAY)
	}

	if len(mb.indices) > 0 {
		mb.indexId.Bind(gl.ELEMENT_ARRAY_BUFFER)
		gl.DrawElements(mode, len(mb.indices), gl.UNSIGNED_INT, uintptr(0))
		mb.indexId.Unbind(gl.ELEMENT_ARRAY_BUFFER)
	} else {
		mb.vertexId.Bind(gl.ARRAY_BUFFER)
		gl.DrawArrays(mode, 0, len(mb.vertices))
		mb.vertexId.Unbind(gl.ARRAY_BUFFER)
	}
}

// RenderMesh renders a single mesh, idenfified by its index.
// The mode defines one of the symbolic constants like GL_TRIANGLES,
// GL_QUADS, GL_POLYGON, GL_TRIANGLE_STRIP, etc.
func (mb *MeshBuffer) RenderMesh(index int, mode gl.GLenum) {
	if index < 0 || index >= len(mb.meshes) {
		return
	}

	if mb.state != mbClean {
		mb.Commit()
	}

	md := mb.meshes[index]

	gl.PushClientAttrib(gl.CLIENT_VERTEX_ARRAY_BIT)
	defer gl.PopClientAttrib()

	if md.VertexCount > 0 {
		gl.EnableClientState(gl.VERTEX_ARRAY)
		defer gl.DisableClientState(gl.VERTEX_ARRAY)
	}

	if md.ColorCount > 0 {
		gl.EnableClientState(gl.COLOR_ARRAY)
		defer gl.DisableClientState(gl.COLOR_ARRAY)
	}

	if md.TextureCount > 0 {
		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		defer gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
	}

	if md.NormalCount > 0 {
		gl.EnableClientState(gl.NORMAL_ARRAY)
		defer gl.DisableClientState(gl.NORMAL_ARRAY)
	}

	if md.IndexCount > 0 {
		start := md.IndexStart * mbIndexStride

		mb.indexId.Bind(gl.ELEMENT_ARRAY_BUFFER)
		gl.DrawElements(mode, md.IndexCount, gl.UNSIGNED_INT, uintptr(start))
		mb.indexId.Unbind(gl.ELEMENT_ARRAY_BUFFER)
	} else {
		start := md.VertexStart * mbVertexStride

		mb.vertexId.Bind(gl.ARRAY_BUFFER)
		gl.DrawArrays(mode, start, md.VertexCount)
		mb.vertexId.Unbind(gl.ARRAY_BUFFER)
	}
}

// Commit pushes the buffer data to the GPU.
//
// This is normally called implicitely by MeshBuffer.Render and only
// when necessary. However, it may need to be called manually, when the
// buffer data is changed.
func (mb *MeshBuffer) Commit() {
	if mb.state == mbFrozen {
		return
	}

	// Upload vertices.
	size := len(mb.vertices) * mbVertexStride
	if size > 0 {
		mb.vertexId.Bind(gl.ARRAY_BUFFER)

		if size != mb.gpuVertexSize {
			gl.BufferData(gl.ARRAY_BUFFER, size, mb.vertices, mb.usage)
			mb.gpuVertexSize = size
		} else {
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, size, mb.vertices)
		}

		gl.VertexPointer(3, gl.FLOAT, 0, uintptr(0))
		mb.vertexId.Unbind(gl.ARRAY_BUFFER)
	}

	// Upload colors.
	size = len(mb.colors) * mbColorStride
	if size > 0 {
		mb.colorId.Bind(gl.ARRAY_BUFFER)

		if size != mb.gpuColorSize {
			gl.BufferData(gl.ARRAY_BUFFER, size, mb.colors, mb.usage)
			mb.gpuColorSize = size
		} else {
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, size, mb.colors)
		}

		gl.ColorPointer(4, gl.FLOAT, 0, uintptr(0))
		mb.colorId.Unbind(gl.ARRAY_BUFFER)
	}

	// Upload normals.
	size = len(mb.normals) * mbNormalStride
	if size > 0 {
		mb.normalId.Bind(gl.ARRAY_BUFFER)

		if size != mb.gpuNormalSize {
			gl.BufferData(gl.ARRAY_BUFFER, size, mb.normals, mb.usage)
			mb.gpuNormalSize = size
		} else {
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, size, mb.normals)
		}

		gl.NormalPointer(gl.FLOAT, 0, uintptr(0))
		mb.normalId.Unbind(gl.ARRAY_BUFFER)
	}

	// Upload texture coords.
	size = len(mb.textures) * mbTextureStride
	if size > 0 {
		mb.textureId.Bind(gl.ARRAY_BUFFER)

		if size != mb.gpuTextureSize {
			gl.BufferData(gl.ARRAY_BUFFER, size, mb.textures, mb.usage)
			mb.gpuTextureSize = size
		} else {
			gl.BufferSubData(gl.ARRAY_BUFFER, 0, size, mb.textures)
		}

		gl.TexCoordPointer(3, gl.FLOAT, 0, uintptr(0))
		mb.textureId.Unbind(gl.ARRAY_BUFFER)
	}

	// Upload indices.
	size = len(mb.indices) * mbIndexStride
	if size > 0 {
		mb.indexId.Bind(gl.ELEMENT_ARRAY_BUFFER)

		if size != mb.gpuIndexSize {
			gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, mb.indices, mb.usage)
			mb.gpuIndexSize = size
		} else {
			gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, size, mb.indices)
		}

		gl.IndexPointer(gl.UNSIGNED_INT, 0, uintptr(0))
		mb.indexId.Unbind(gl.ELEMENT_ARRAY_BUFFER)
	}

	gl.Finish()
	mb.state = mbClean
}

// Append appends new mesh to the buffer.
// Returns an index for a MeshDescriptor object. This can be used to index
// the list returned by MeshBuffer.Meshes()
func (mb *MeshBuffer) Append(m *Mesh) int {
	mb.state = mbFrozen

	md := new(MeshDescriptor)
	md.VertexStart = len(mb.vertices)
	md.VertexCount = len(m.Vertices)
	md.TextureStart = len(mb.textures)
	md.TextureCount = len(m.Textures)
	md.NormalStart = len(mb.normals)
	md.NormalCount = len(m.Normals)
	md.ColorStart = len(mb.colors)
	md.ColorCount = len(m.Colors)
	md.IndexStart = len(mb.indices)
	md.IndexCount = len(m.Indices)

	mb.meshes = append(mb.meshes, md)
	mb.vertices = append(mb.vertices, m.Vertices...)
	mb.textures = append(mb.textures, m.Textures...)
	mb.normals = append(mb.normals, m.Normals...)
	mb.colors = append(mb.colors, m.Colors...)
	mb.indices = append(mb.indices, m.Indices...)

	// Update indices into the mesh buffers.
	for i := md.IndexStart; i < len(mb.indices); i++ {
		mb.indices[i] += uint(md.VertexStart)
	}

	mb.state = mbDirty
	return len(mb.meshes) - 1
}

// Delete removes the mesh with the given index from the buffer.
func (mb *MeshBuffer) Delete(index int) {
	if index < 0 || index >= len(mb.meshes) {
		return
	}

	md := mb.meshes[index]

	// Remove vertices.
	s, c := md.VertexStart, md.VertexCount
	copy(mb.vertices[s:], mb.vertices[s+c:])
	mb.vertices = mb.vertices[:len(mb.vertices)-c]

	// Remove texture coords.
	s, c = md.TextureStart, md.TextureCount
	copy(mb.textures[s:], mb.textures[s+c:])
	mb.textures = mb.textures[:len(mb.textures)-c]

	// Remove normals.
	s, c = md.NormalStart, md.NormalCount
	copy(mb.normals[s:], mb.normals[s+c:])
	mb.normals = mb.normals[:len(mb.normals)-c]

	// Remove colors.
	s, c = md.ColorStart, md.ColorCount
	copy(mb.colors[s:], mb.colors[s+c:])
	mb.colors = mb.colors[:len(mb.colors)-c]

	// Remove indices.
	s, c = md.IndexStart, md.IndexCount
	copy(mb.indices[s:], mb.indices[s+c:])
	mb.indices = mb.indices[:len(mb.indices)-c]

	// Remove mesh descriptor.
	copy(mb.meshes[index:], mb.meshes[index+1:])
	mb.meshes = mb.meshes[:len(mb.meshes)-1]

	mb.state = mbDirty
}

// Usage returns the usage type associated with this buffer.
func (mb *MeshBuffer) Usage() gl.GLenum { return mb.usage }

// SetUsage sets the usage type for this buffer. Picking the right value for
// the purpose of this buffer can greatly affect performance.
// It defaults to GL_DYNAMIC_DRAW.
//
// The expected values are:
//
//    GL_STATIC_DRAW
//    GL_DYNAMIC_DRAW
//    GL_STREAM_DRAW
//
// "STATIC" means the data in mbO will not be changed (specified once and used
// many times).
//
// "DYNAMIC" means the data will be changed frequently (specified and used
// repeatedly).
//
// "STREAM" means the data will be changed every frame (specified once and
// used once).
//
// The mbO memory manager will choose the best memory places for the buffer
// object based on this usage flag. For example: GL_STATIC_DRAW and
// GL_STREAM_DRAW may use video memory, while GL_DYNAMIC_DRAW may use AGP memory.
//
// Note: This issues a re-commit of the buffer data to take effect.
func (mb *MeshBuffer) SetUsage(usage gl.GLenum) {
	mb.usage = usage
	mb.state = mbDirty
}

// Len returns the number of mesh objects in the buffer.
func (mb *MeshBuffer) Len() int { return len(mb.meshes) }

// Meshes returns a list of mesh descriptors.
func (mb *MeshBuffer) Meshes() []*MeshDescriptor { return mb.meshes }

// Vertices returns a list of all vertices in the buffer.
func (mb *MeshBuffer) Vertices() [][3]float32 { return mb.vertices }

// Colors returns a list of all colors in the buffer.
func (mb *MeshBuffer) Colors() [][4]float32 { return mb.colors }

// Textures returns a list of all texture coords in the buffer.
func (mb *MeshBuffer) Textures() [][3]float32 { return mb.textures }

// Normals returns a list of all normals in the buffer.
func (mb *MeshBuffer) Normals() [][3]float32 { return mb.normals }

// Indices returns a list of all indices in the buffer.
func (mb *MeshBuffer) Indices() []uint { return mb.indices }
