// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"image"
	"log"

	"github.com/go-gl/gl"
)

// Mapping from texture dimensions onto ready made framebuffer/renderbuffer
// therefore we only construct one per image dimensions
// This number should be less than O(1000) otherwise opengl throws OUT_OF_MEMORY
// on some cards
var framebuffers map[image.Point]*fborbo = make(map[image.Point]*fborbo)

type fborbo struct {
	fbo gl.Framebuffer
	rbo gl.Renderbuffer
}

// Internal function to generate a framebuffer/renderbuffer of the correct
// dimensions exactly once per execution
func getFBORBO(t *Texture) *fborbo {
	p := image.Point{t.W, t.H}
	result, ok := framebuffers[p]
	if ok {
		return result
	}

	result = &fborbo{}

	result.rbo = gl.GenRenderbuffer()
	OpenGLSentinel()
	result.fbo = gl.GenFramebuffer()
	OpenGLSentinel()

	result.fbo.Bind()

	result.rbo.Bind()
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, t.W, t.H)
	result.rbo.Unbind()

	result.rbo.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT,
		gl.RENDERBUFFER)

	result.fbo.Unbind()

	framebuffers[image.Point{t.W, t.H}] = result
	return result
}

// During this context, OpenGL drawing operations will instead render to `*Texture`.
// Example usage:
//     With(Framebuffer{my_texture}, func() { .. operations to render to texture .. })
//
// Internally this will permanently allocate a framebuffer with the appropriate
// dimensions. Beware that using a large number of textures with differing sizes
// will therefore cause some graphics cards to run out of memory.
type Framebuffer struct {
	*Texture
	*fborbo
	Level int
}

func (b *Framebuffer) Enter() {
	if b.fborbo == nil {
		b.fborbo = getFBORBO(b.Texture)
	}

	b.fbo.Bind()

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D,
		b.Texture.Texture, b.Level)

	s := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if s != gl.FRAMEBUFFER_COMPLETE {
		log.Panicf("Incomplete framebuffer, reason: %x", s)
	}
}

func (b *Framebuffer) Exit() {
	b.fbo.Unbind()
}
