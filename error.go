// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"fmt"
	"github.com/go-gl/gl"
	"github.com/go-gl/glu"
)

// CheckGLError returns an opengl error if one exists.
func CheckGLError() error {
	errno := gl.GetError()

	if errno == gl.NO_ERROR {
		return nil
	}

	str, err := glu.ErrorString(errno)
	if err != nil {
		return fmt.Errorf("Unknown GL error: %d", errno)
	}

	return fmt.Errorf(str)
}
