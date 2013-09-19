// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"github.com/go-gl/gl"
	"image"
	"image/draw"
	"image/png"
	"io"
)

// A 2D Texture which implements Context, so can be used as:
//   `With(texture, func() { .. textured primitives .. })`
// which sets gl.ENABLE_BIT
type Texture struct {
	gl.Texture
	W, H int
}

// Create a new texture, initialize it to have a `gl.LINEAR` filter and use
// `gl.CLAMP_TO_EDGE`.
func NewTexture(w, h int) *Texture {
	texture := &Texture{gl.GenTexture(), w, h}
	With(texture, func() {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	})
	return texture
}

// Initialize texture storage. _REQUIRED_ before using it as a framebuffer target.
func (t *Texture) Init() {
	With(t, func() {
		// generate base level storage
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, t.W, t.H, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
		// generate required number of mipmaps given texture dimensions
		gl.GenerateMipmap(gl.TEXTURE_2D)
	})
}

func (b Texture) Enter() {
	gl.PushAttrib(gl.ENABLE_BIT)
	gl.Enable(gl.TEXTURE_2D)
	b.Bind(gl.TEXTURE_2D)
}
func (b Texture) Exit() {
	b.Unbind(gl.TEXTURE_2D)
	gl.PopAttrib()
}

// Return the OpenGL texture as a golang `image.RGBA`
func (t *Texture) AsImage() *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, t.W, t.H))
	With(t, func() {
		// TODO: check internal format (with GetIntegerv?)
		gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)
	})
	return rgba
}

func (t *Texture) FromImageRGBA(rgba *image.RGBA, level int) {
	With(t, func() {
		gl.TexImage2D(gl.TEXTURE_2D, level, gl.RGBA,
			rgba.Bounds().Dx(), rgba.Bounds().Dy(),
			0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)
	})
}

// Initialize this texture with image data from `im`. Note: copies the texture
// if it isn't in RGBA format. This can happen if you load from png files.
func (t *Texture) FromImage(im image.Image, level int) {
	switch trueim := im.(type) {
	case *image.RGBA:
		t.FromImageRGBA(trueim, level)

	default:
		copy := image.NewRGBA(trueim.Bounds())
		draw.Draw(copy, trueim.Bounds(), trueim, image.Pt(0, 0), draw.Src)
		t.FromImageRGBA(copy, level)
	}
}

func (t *Texture) FromPngReader(in io.Reader, level int) error {
	im, err := png.Decode(in)
	if err != nil {
		return err
	}
	t.FromImage(im, level)
	return nil
}
