// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glhelpers

import (
	"log"

	"github.com/go-gl/gl"
)

func MakeShader(shader_type gl.GLenum, source string) gl.Shader {

	shader := gl.CreateShader(shader_type)
	shader.Source(source)
	shader.Compile()
	OpenGLSentinel()

	compstat := shader.Get(gl.COMPILE_STATUS)
	if compstat != 1 {
		log.Print("vert shader compilation status: ", compstat)
		log.Print("Info log: ", shader.GetInfoLog())
		log.Panic("Problem creating shader?")
	}
	return shader
}

func ExammpleProgram() gl.Program {
	return MakeProgram(`
		#version 120
		// 330 compatibility

		//layout (location = 0) in vec4 position;
		//layout (location = 1) in vec4 color;

		varying out vec4 vertexcolor;
		//smooth out vec4 vertexcolor;

		// Passthrough vertex shader
		void main() {
			gl_Position = ftransform();
			gl_FrontColor = gl_Color;
		}
	`, `
		#version 400
		#extension GL_EXT_geometry_shader4 : enable

		layout (points) in;
		layout (line_strip, max_vertices = 4) out;

		// This shader takes points and turns them into lines
		void main () {
			for(int i = 0; i < gl_VerticesIn; i++) {
				gl_Position = gl_in[i].gl_Position;
				gl_FrontColor =  gl_in[i].gl_FrontColor;
				EmitVertex();

				gl_Position = gl_in[i].gl_Position + vec4(0.0, 0.005, 0, 0);
				gl_FrontColor =  gl_in[i].gl_FrontColor;
				EmitVertex();
			}
		}
	`, `
		#version 120
		// 440 compatibility
		
		//in vec4 vertexcolor;
		//varying out vec4 outputColor;
		
		void main() {
			//gl_FragColor = gl_Color; //vertexcolor;
			//outputColor = vertexcolor;
			gl_FragColor = gl_Color;
		}
	`)
}

func MakeProgram(vertex, geometry, fragment string) gl.Program {
	vert_shader := MakeShader(gl.VERTEX_SHADER, vertex)
	geom_shader := MakeShader(gl.GEOMETRY_SHADER, geometry)
	frag_shader := MakeShader(gl.FRAGMENT_SHADER, fragment)

	prog := gl.CreateProgram()
	prog.AttachShader(vert_shader)
	prog.AttachShader(geom_shader)
	prog.AttachShader(frag_shader)
	prog.Link()

	OpenGLSentinel()

	// Note: These functions aren't implemented in master of banathar/gl
	// prog.ParameterEXT(gl.GEOMETRY_INPUT_TYPE_EXT, gl.POINTS)
	// prog.ParameterEXT(gl.GEOMETRY_OUTPUT_TYPE_EXT, gl.POINTS)
	// prog.ParameterEXT(gl.GEOMETRY_VERTICES_OUT_EXT, 4)

	// OpenGLSentinel()

	linkstat := prog.Get(gl.LINK_STATUS)
	if linkstat != 1 {
		log.Panic("Program link failed, status=", linkstat,
			"Info log: ", prog.GetInfoLog())
	}

	prog.Validate()
	valstat := prog.Get(gl.VALIDATE_STATUS)
	if valstat != 1 {
		log.Panic("Program validation failed: ", valstat)
	}

	return prog
}
