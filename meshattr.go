// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"github.com/go-gl/gl"
)

// A MeshAttr describes the type and size of a single vertex component.
// These tell the MeshBuffer how to interpret mesh data. For example:
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
type MeshAttr interface {
	// Append expects a slice of values, equal to the type of the attribute.
	// E.g.: []uint32, []float32, []byte, etc.
	// It returns the number of elements that were added.
	Append(interface{}) int

	// Removes elements in the given range.
	Remove(start, count int)

	// Increment the value of the elements in the given range.
	// This is necessary to adjust offsets for indices.
	Increment(start, value int)

	Buffer()
	Bind()
	Unbind()

	Clear()
	Release()

	Len() int
	Size() int
	Stride() int
	Type() gl.GLenum
	Usage() gl.GLenum

	Target() gl.GLenum
	SetTarget(gl.GLenum)

	GpuSize() int
	SetGpuSize(int)

	// Dirty returns true if the attribute data has been changed 
	// and it needs to be re-committed.
	Dirty() bool
	SetDirty(bool)
}

// #############################################################################
// AttrBase is the base type for attributes. It takes care of 
// operations common to all attribute types, so we don't have to keep
// repeating it.
type AttrBase struct {
	buffer  gl.Buffer // Buffer identity.
	usage   gl.GLenum // Usage type of this attribute.
	target  gl.GLenum // Buffer type.
	typ     gl.GLenum // Attribute type.
	size    int       // Component size (number of elements).
	stride  int       // Size of component in bytes.
	gpuSize int       // Size of data on GPU.
	dirty   bool      // Do we require re-committing?
}

// NewAttrBase creates a new attribute.
func NewAttrBase(size int, typ, usage gl.GLenum) *AttrBase {
	a := new(AttrBase)
	a.size = size
	a.target = gl.ARRAY_BUFFER
	a.usage = usage
	a.typ = typ
	a.stride = int(Sizeof(typ))

	if size > 0 {
		a.buffer = gl.GenBuffer()
	}

	return a
}

// Release releases attribute resources.
func (a *AttrBase) Release() {
	if a.size > 0 {
		a.buffer.Delete()
		a.buffer = 0
	}

	a.gpuSize = 0
}

func (a *AttrBase) Clear() {
	a.gpuSize = 0
	a.dirty = true
}

func (a *AttrBase) Dirty() bool           { return a.dirty }
func (a *AttrBase) SetDirty(v bool)       { a.dirty = v }
func (a *AttrBase) Size() int             { return a.size }
func (a *AttrBase) Stride() int           { return a.stride }
func (a *AttrBase) Bind()                 { a.buffer.Bind(a.target) }
func (a *AttrBase) Unbind()               { a.buffer.Unbind(a.target) }
func (a *AttrBase) Target() gl.GLenum     { return a.target }
func (a *AttrBase) SetTarget(t gl.GLenum) { a.target = t }
func (a *AttrBase) Type() gl.GLenum       { return a.typ }
func (a *AttrBase) Usage() gl.GLenum      { return a.usage }
func (a *AttrBase) GpuSize() int          { return a.gpuSize }
func (a *AttrBase) SetGpuSize(sz int)     { a.gpuSize = sz }

// #############################################################################
// A NullAttr is a mesh attribute for null values.
// This is a special case used when one supplies 'nil' for a given
// mesh attribute. It does not allocate a buffer and is used as a mock.
type NullAttr struct {
	*AttrBase
}

// NewNullAttr creates a new attribute.
func NewNullAttr() MeshAttr {
	a := new(NullAttr)
	a.AttrBase = NewAttrBase(0, gl.BYTE, 0)
	return a
}

func (a *NullAttr) Len() int               { return 0 }
func (a *NullAttr) Release()               {}
func (a *NullAttr) Clear()                 {}
func (a *NullAttr) Buffer()                {}
func (a *NullAttr) Increment(int, int)     {}
func (a *NullAttr) Append(interface{}) int { return 0 }
func (a *NullAttr) Remove(int, int)        {}

// #############################################################################
// A Int8Attr is a mesh attribute for signed bytes.
type Int8Attr struct {
	*AttrBase
	data []int8
}

// NewInt8Attr creates a new attribute.
func NewInt8Attr(size int, usage gl.GLenum) MeshAttr {
	a := new(Int8Attr)
	a.AttrBase = NewAttrBase(size, gl.BYTE, usage)
	return a
}

func (a *Int8Attr) Data() []int8 { return a.data }
func (a *Int8Attr) Len() int     { return len(a.data) }
func (a *Int8Attr) Release() {
	a.AttrBase.Release()
	a.AttrBase = nil
	a.data = nil
}

func (a *Int8Attr) Clear() {
	a.AttrBase.Clear()
	a.data = nil
}

func (a *Int8Attr) Buffer() {
	size := len(a.data) * a.Stride()

	if size != a.GpuSize() {
		gl.BufferData(a.Target(), size, a.data, a.Usage())
		a.SetGpuSize(size)
	} else {
		gl.BufferSubData(a.Target(), 0, size, a.data)
	}

	a.SetDirty(false)
}

func (a *Int8Attr) Increment(start, value int) {
	for i := start; i < len(a.data); i++ {
		a.data[i] += int8(value)
	}
	a.SetDirty(false)
}

func (a *Int8Attr) Append(data interface{}) int {
	v := data.([]int8)
	a.data = append(a.data, v...)
	a.SetDirty(true)
	return len(v)
}

func (a *Int8Attr) Remove(s, c int) {
	s *= a.Size()
	c *= a.Size()

	copy(a.data[s:], a.data[s+c:])
	a.data = a.data[:len(a.data)-c]
	a.SetDirty(true)
}

// #############################################################################
// A Uint8Attr is a mesh attribute for unsigned bytes.
type Uint8Attr struct {
	*AttrBase
	data []uint8
}

// NewUint8Attr creates a new attribute.
func NewUint8Attr(size int, usage gl.GLenum) MeshAttr {
	a := new(Uint8Attr)
	a.AttrBase = NewAttrBase(size, gl.UNSIGNED_BYTE, usage)
	return a
}

func (a *Uint8Attr) Data() []uint8 { return a.data }
func (a *Uint8Attr) Len() int      { return len(a.data) }
func (a *Uint8Attr) Release() {
	a.AttrBase.Release()
	a.AttrBase = nil
	a.data = nil
}

func (a *Uint8Attr) Clear() {
	a.AttrBase.Clear()
	a.data = nil
}

func (a *Uint8Attr) Buffer() {
	size := len(a.data) * a.Stride()

	if size != a.GpuSize() {
		gl.BufferData(a.Target(), size, a.data, a.Usage())
		a.SetGpuSize(size)
	} else {
		gl.BufferSubData(a.Target(), 0, size, a.data)
	}

	a.SetDirty(false)
}

func (a *Uint8Attr) Increment(start, value int) {
	for i := start; i < len(a.data); i++ {
		a.data[i] += uint8(value)
	}
	a.SetDirty(true)
}

func (a *Uint8Attr) Append(data interface{}) int {
	v := data.([]uint8)
	a.data = append(a.data, v...)
	a.SetDirty(true)
	return len(v)
}

func (a *Uint8Attr) Remove(s, c int) {
	s *= a.Size()
	c *= a.Size()

	copy(a.data[s:], a.data[s+c:])
	a.data = a.data[:len(a.data)-c]
	a.SetDirty(true)
}

// #############################################################################
// A Int16Attr is a mesh attribute for signed shorts.
type Int16Attr struct {
	*AttrBase
	data []int16
}

// NewInt16Attr creates a new attribute.
func NewInt16Attr(size int, usage gl.GLenum) MeshAttr {
	a := new(Int16Attr)
	a.AttrBase = NewAttrBase(size, gl.SHORT, usage)
	return a
}

func (a *Int16Attr) Data() []int16 { return a.data }
func (a *Int16Attr) Len() int      { return len(a.data) }
func (a *Int16Attr) Release() {
	a.AttrBase.Release()
	a.AttrBase = nil
	a.data = nil
}

func (a *Int16Attr) Clear() {
	a.AttrBase.Clear()
	a.data = nil
}

func (a *Int16Attr) Buffer() {
	size := len(a.data) * a.Stride()

	if size != a.GpuSize() {
		gl.BufferData(a.Target(), size, a.data, a.Usage())
		a.SetGpuSize(size)
	} else {
		gl.BufferSubData(a.Target(), 0, size, a.data)
	}

	a.SetDirty(true)
}

func (a *Int16Attr) Increment(start, value int) {
	for i := start; i < len(a.data); i++ {
		a.data[i] += int16(value)
	}
	a.SetDirty(true)
}

func (a *Int16Attr) Append(data interface{}) int {
	v := data.([]int16)
	a.data = append(a.data, v...)
	a.SetDirty(true)
	return len(v)
}

func (a *Int16Attr) Remove(s, c int) {
	s *= a.Size()
	c *= a.Size()

	copy(a.data[s:], a.data[s+c:])
	a.data = a.data[:len(a.data)-c]
	a.SetDirty(true)
}

// #############################################################################
// A Uint16Attr is a mesh attribute for unsigned shorts.
type Uint16Attr struct {
	*AttrBase
	data []uint16
}

func NewUint16Attr(size int, usage gl.GLenum) MeshAttr {
	a := new(Uint16Attr)
	a.AttrBase = NewAttrBase(size, gl.UNSIGNED_SHORT, usage)
	return a
}

func (a *Uint16Attr) Data() []uint16 { return a.data }
func (a *Uint16Attr) Len() int       { return len(a.data) }
func (a *Uint16Attr) Release() {
	a.AttrBase.Release()
	a.AttrBase = nil
	a.data = nil
}

func (a *Uint16Attr) Clear() {
	a.AttrBase.Clear()
	a.data = nil
}

func (a *Uint16Attr) Buffer() {
	size := len(a.data) * a.Stride()

	if size != a.GpuSize() {
		gl.BufferData(a.Target(), size, a.data, a.Usage())
		a.SetGpuSize(size)
	} else {
		gl.BufferSubData(a.Target(), 0, size, a.data)
	}

	a.SetDirty(false)
}

func (a *Uint16Attr) Increment(start, value int) {
	for i := start; i < len(a.data); i++ {
		a.data[i] += uint16(value)
	}
	a.SetDirty(true)
}

func (a *Uint16Attr) Append(data interface{}) int {
	v := data.([]uint16)
	a.data = append(a.data, v...)
	a.SetDirty(true)
	return len(v)
}

func (a *Uint16Attr) Remove(s, c int) {
	s *= a.Size()
	c *= a.Size()

	copy(a.data[s:], a.data[s+c:])
	a.data = a.data[:len(a.data)-c]
	a.SetDirty(true)
}

// #############################################################################
// A Int32Attr is a mesh attribute for signed ints.
type Int32Attr struct {
	*AttrBase
	data []int32
}

// NewInt32Attr creates a new attribute.
func NewInt32Attr(size int, usage gl.GLenum) MeshAttr {
	a := new(Int32Attr)
	a.AttrBase = NewAttrBase(size, gl.INT, usage)
	return a
}

func (a *Int32Attr) Data() []int32 { return a.data }
func (a *Int32Attr) Len() int      { return len(a.data) }
func (a *Int32Attr) Release() {
	a.AttrBase.Release()
	a.AttrBase = nil
	a.data = nil
}

func (a *Int32Attr) Clear() {
	a.AttrBase.Clear()
	a.data = nil
}

func (a *Int32Attr) Buffer() {
	size := len(a.data) * a.Stride()

	if size != a.GpuSize() {
		gl.BufferData(a.Target(), size, a.data, a.Usage())
		a.SetGpuSize(size)
	} else {
		gl.BufferSubData(a.Target(), 0, size, a.data)
	}

	a.SetDirty(false)
}

func (a *Int32Attr) Increment(start, value int) {
	for i := start; i < len(a.data); i++ {
		a.data[i] += int32(value)
	}
	a.SetDirty(true)
}

func (a *Int32Attr) Append(data interface{}) int {
	v := data.([]int32)
	a.data = append(a.data, v...)
	a.SetDirty(true)
	return len(v)
}

func (a *Int32Attr) Remove(s, c int) {
	s *= a.Size()
	c *= a.Size()

	copy(a.data[s:], a.data[s+c:])
	a.data = a.data[:len(a.data)-c]
	a.SetDirty(true)
}

// #############################################################################
// A Uint32Attr is a mesh attribute for unsigned ints.
type Uint32Attr struct {
	*AttrBase
	data []uint32
}

func NewUint32Attr(size int, usage gl.GLenum) MeshAttr {
	a := new(Uint32Attr)
	a.AttrBase = NewAttrBase(size, gl.UNSIGNED_INT, usage)
	return a
}

func (a *Uint32Attr) Data() []uint32 { return a.data }
func (a *Uint32Attr) Len() int       { return len(a.data) }
func (a *Uint32Attr) Release() {
	a.AttrBase.Release()
	a.AttrBase = nil
	a.data = nil
}

func (a *Uint32Attr) Clear() {
	a.AttrBase.Clear()
	a.data = nil
}

func (a *Uint32Attr) Buffer() {
	size := len(a.data) * a.Stride()

	if size != a.GpuSize() {
		gl.BufferData(a.Target(), size, a.data, a.Usage())
		a.SetGpuSize(size)
	} else {
		gl.BufferSubData(a.Target(), 0, size, a.data)
	}

	a.SetDirty(false)
}

func (a *Uint32Attr) Increment(start, value int) {
	for i := start; i < len(a.data); i++ {
		a.data[i] += uint32(value)
	}
	a.SetDirty(true)
}

func (a *Uint32Attr) Append(data interface{}) int {
	v := data.([]uint32)
	a.data = append(a.data, v...)
	a.SetDirty(true)
	return len(v)
}

func (a *Uint32Attr) Remove(s, c int) {
	s *= a.Size()
	c *= a.Size()

	copy(a.data[s:], a.data[s+c:])
	a.data = a.data[:len(a.data)-c]
	a.SetDirty(true)
}

// #############################################################################
// A Floa32tAttr is a mesh attribute for float values.
type Float32Attr struct {
	*AttrBase
	data []float32
}

// NewFloatAttr creates a new attribute.
func NewFloat32Attr(size int, usage gl.GLenum) MeshAttr {
	a := new(Float32Attr)
	a.AttrBase = NewAttrBase(size, gl.FLOAT, usage)
	return a
}

func (a *Float32Attr) Data() []float32 { return a.data }
func (a *Float32Attr) Len() int        { return len(a.data) }
func (a *Float32Attr) Release() {
	a.AttrBase.Release()
	a.AttrBase = nil
	a.data = nil
}

func (a *Float32Attr) Clear() {
	a.AttrBase.Clear()
	a.data = nil
}

func (a *Float32Attr) Buffer() {
	size := len(a.data) * a.Stride()

	if size != a.GpuSize() {
		gl.BufferData(a.Target(), size, a.data, a.Usage())
		a.SetGpuSize(size)
	} else {
		gl.BufferSubData(a.Target(), 0, size, a.data)
	}

	a.SetDirty(false)
}

func (a *Float32Attr) Increment(start, value int) {
	for i := start; i < len(a.data); i++ {
		a.data[i] += float32(value)
	}
	a.SetDirty(true)
}

func (a *Float32Attr) Append(data interface{}) int {
	v := data.([]float32)
	a.data = append(a.data, v...)
	a.SetDirty(true)
	return len(v)
}

func (a *Float32Attr) Remove(s, c int) {
	s *= a.Size()
	c *= a.Size()

	copy(a.data[s:], a.data[s+c:])
	a.data = a.data[:len(a.data)-c]
	a.SetDirty(true)
}

// #############################################################################
// A Floa64tAttr is a mesh attribute for double values.
type Float64Attr struct {
	*AttrBase
	data []float64
}

// NewFloatAttr creates a new attribute.
func NewFloat64Attr(size int, usage gl.GLenum) MeshAttr {
	a := new(Float32Attr)
	a.AttrBase = NewAttrBase(size, gl.DOUBLE, usage)
	return a
}

func (a *Float64Attr) Data() []float64 { return a.data }
func (a *Float64Attr) Len() int        { return len(a.data) }
func (a *Float64Attr) Release() {
	a.AttrBase.Release()
	a.AttrBase = nil
	a.data = nil
}

func (a *Float64Attr) Clear() {
	a.AttrBase.Clear()
	a.data = nil
}

func (a *Float64Attr) Buffer() {
	size := len(a.data) * a.Stride()

	if size != a.GpuSize() {
		gl.BufferData(a.Target(), size, a.data, a.Usage())
		a.SetGpuSize(size)
	} else {
		gl.BufferSubData(a.Target(), 0, size, a.data)
	}

	a.SetDirty(false)
}

func (a *Float64Attr) Increment(start, value int) {
	for i := start; i < len(a.data); i++ {
		a.data[i] += float64(value)
	}
	a.SetDirty(true)
}

func (a *Float64Attr) Append(data interface{}) int {
	v := data.([]float64)
	a.data = append(a.data, v...)
	a.SetDirty(true)
	return len(v)
}

func (a *Float64Attr) Remove(s, c int) {
	s *= a.Size()
	c *= a.Size()

	copy(a.data[s:], a.data[s+c:])
	a.data = a.data[:len(a.data)-c]
	a.SetDirty(true)
}
