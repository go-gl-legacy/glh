package glhelpers

import (
	"image"

	"github.com/go-gl/gl"
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
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, t.W, t.H, 0, gl.RGBA,
			gl.UNSIGNED_BYTE, nil)
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
