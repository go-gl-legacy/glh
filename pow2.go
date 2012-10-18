// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"fmt"
	"image"
	"image/color"
)

// Pow2 returns the first power-of-two value >= to n.
// This can be used to create suitable texture dimensions.
func Pow2(x uint32) uint32 {
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	return x + 1
}

// IsPow2 returns true if the given value is a power-of-two.
func IsPow2(x uint32) bool { return (x & (x - 1)) == 0 }

// Pow2Image returns the given image, scaled to the smallest power-of-two
// dimensions larger or equal to the input dimensions.
// It preserves the image format and contents.
//
// This is useful if an image is to be used as an OpenGL texture.
// These often require image data to have power-of-two dimensions.
func Pow2Image(src image.Image) image.Image {
	sb := src.Bounds()
	w, h := uint32(sb.Dx()), uint32(sb.Dy())

	if IsPow2(w) && IsPow2(h) {
		return src // Nothing to do.
	}

	rect := image.Rect(0, 0, int(Pow2(w)), int(Pow2(h)))

	switch src := src.(type) {
	case *image.Alpha:
		dst := image.NewAlpha(rect)
		copyImg(src, dst)
		return dst

	case *image.Alpha16:
		dst := image.NewAlpha16(rect)
		copyImg(src, dst)
		return dst

	case *image.Gray:
		dst := image.NewGray(rect)
		copyImg(src, dst)
		return dst

	case *image.Gray16:
		dst := image.NewGray16(rect)
		copyImg(src, dst)
		return dst

	case *image.NRGBA:
		dst := image.NewNRGBA(rect)
		copyImg(src, dst)
		return dst

	case *image.NRGBA64:
		dst := image.NewNRGBA64(rect)
		copyImg(src, dst)
		return dst

	case *image.Paletted:
		dst := image.NewPaletted(rect, src.Palette)
		copyImg(src, dst)
		return dst

	case *image.RGBA:
		dst := image.NewRGBA(rect)
		copyImg(src, dst)
		return dst

	case *image.RGBA64:
		dst := image.NewRGBA64(rect)
		copyImg(src, dst)
		return dst
	}

	panic(fmt.Sprintf("Unsupported image format: %T", src))
}

// Why the image.Image interface does not support this,
// I can never understand.
type copyable interface {
	image.Image
	Set(x, y int, clr color.Color)
}

func copyImg(src, dst copyable) {
	var x, y int
	sb := src.Bounds()

	for y = 0; y < sb.Dy(); y++ {
		for x = 0; x < sb.Dx(); x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
}
