package glhelpers

import (
	"image"
	"image/draw"
	"io/ioutil"
	"log"

	"code.google.com/p/freetype-go/freetype"

	"github.com/andrebq/gas"

	"github.com/go-gl/gl"
)

var FontFile string

func init() {
	font := "code.google.com/p/freetype-go/luxi-fonts/luximr.ttf"
	var err error
	FontFile, err = gas.Abs(font)
	if err != nil {
		log.Panicf("Unable to load font file: %s", font)
	}
}

type Text struct {
	str string
	*Texture
	Flipped   bool
	DebugRect bool
}

// Create a *Texture containing a rendering of `str` with `size`.
// TODO: allow for alternative fonts
func MakeText(str string, size float64) *Text {
	if str == "" {
		panic("Trying to build empty text")
	}

	defer OpenGLSentinel()()

	// TODO: Something if font doesn't exist
	fontBytes, err := ioutil.ReadFile(FontFile)
	if err != nil {
		log.Panic(err)
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Panic(err)
	}

	fg, bg := image.White, image.Black

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(size)

	pt := freetype.Pt(0, int(c.PointToFix32(size)>>8))
	s, err := c.DrawString(str, pt)
	if err != nil {
		log.Panic("Error: ", err)
	}

	spacing, offset := 1.5, 3

	w := int(s.X >> 8)
	h := int(c.PointToFix32(size*spacing)>>8) - offset

	text := &Text{str: str, Texture: NewTexture(w, h)}

	if text.W > 4096 {
		text.W = 4096
	}

	rgba := image.NewRGBA(image.Rect(0, 0, text.W, text.H))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)

	_, err = c.DrawString(text.str, pt)
	if err != nil {
		log.Panic("Error: ", err)
	}

	With(text, func() {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, text.W, text.H, 0, gl.RGBA,
			gl.UNSIGNED_BYTE, rgba.Pix)
	})

	if gl.GetError() != gl.NO_ERROR {
		log.Panic("Failed to load a texture, err = ", gl.GetError(),
			" str = ", str, " w = ", text.W, " h = ", text.H)
	}

	return text
}

// Delete the underlying texture.
func (text *Text) Destroy() {
	text.Texture.Delete()
}

// Draw `text` at `x`, `y`.
func (text *Text) Draw(x, y int) {
	var w, h = text.W, text.H

	if text.Flipped {
		h = -h
	}
	With(Attrib{gl.ENABLE_BIT | gl.COLOR_BUFFER_BIT}, func() {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.DST_ALPHA)
		//gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		With(text, func() {
			gl.TexEnvi(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
			DrawQuadi(x, y, w, h)
		})
	})

	if text.DebugRect {
		gl.LineWidth(1)
		With(Primitive{gl.LINE_LOOP}, func() {
			sp := 0
			Squarei(x-sp, y+sp, w+sp*2, h-sp*2)
		})
	}
}
