package text

import (
	"encoding/gob"
	"github.com/errcw/glow/gl-core/3.3/gl"
	"github.com/runningwild/glop/render"
	"image"
	"io"
	"log"
	"sync"
	"unsafe"
)

// Shader stuff - The font stuff requires that we use some simple shaders
// const font_vertex_shader = `
// #version 330
// 	void main() {
// 	  gl_Position = ftransform();
// 	  gl_ClipVertex = gl_ModelViewMatrix * gl_Vertex;
// 	  gl_TexCoord[0] = gl_MultiTexCoord0;
// 	  gl_TexCoord[1] = gl_MultiTexCoord1;
// 	}
// `

// const font_fragment_shader = `
// #version 330
// 	uniform vec4 color;
// 	uniform sampler2D tex;
// 	uniform float dist_min;
// 	uniform float dist_max;

// 	void main() {
// 	  vec2 tpos = gl_TexCoord[0].xy;
// 	  float dist = texture2D(tex, tpos).a;
// 	  float alpha = smoothstep(dist_min, dist_max, dist);
// 	  gl_FragColor = color * vec4(1.0, 1.0, 1.0, alpha);
// 	}
// `

const test_vshader = `
#version 330
in vec3 position;
in vec2 texCoord;

out vec2 theTexCoord; 

void main() { 
   gl_Position = vec4(position, 1.0); 
   theTexCoord = texCoord;
}
`

const test_fshader = `
#version 330
in vec2 theTexCoord;
uniform sampler2D tex;
out vec4 color;

void main() {
	vec4 dd = texture(tex, theTexCoord);
	color = vec4(dd.r+dd.g+dd.b+dd.a, 0.0, dd.a, 1.0);
}
`

type runeInfo struct {
	Pos         image.Rectangle
	Bounds      image.Rectangle
	Full_bounds image.Rectangle
	Advance     float64
}
type dictData struct {
	// The Pix data from the original image.Rgba
	Pix []byte

	Kerning map[rune]map[rune]int

	// Dx and Dy of the original image.Rgba
	Dx, Dy int32

	// Map from rune to that rune's runeInfo.
	Info map[rune]runeInfo

	// runeInfo for all r < 256 will be stored here as well as in info so we can
	// avoid map lookups if possible.
	Ascii_info []runeInfo

	// At what vertical value is the line on which text is logically rendered.
	// This is determined by the positioning of the '.' rune.
	Baseline int

	// Amount glyphs were scaled down during packing.
	Scale float64

	Miny, Maxy int
}
type Dictionary struct {
	data dictData

	vaos [2]uint32
	vbos [4]uint32
	tex  uint32
}

var triangle = [9]float32{-0.4, 0.1, 0.0, 0.4, 0.1, 0.0, 0.3, 0.7, 0.0}
var triangleTexCoords = [6]float32{0, 1, 1, 1, 1, 0}
var quad = [12]float32{-0.2, -0.1, 0.0, -0.2, -0.6, 0.0, 0.2, -0.1, 0.0, 0.2, -0.6, 0.0}
var quadTexCoords = [8]float32{0, 0, 0, 1, 1, 1, 1, 0}

var init_once sync.Once

func LoadDictionary(r io.Reader, l *log.Logger) (*Dictionary, error) {
	errChan := make(chan error)
	init_once.Do(func() {
		render.Queue(func() {
			// errChan <- render.RegisterShader("glop.font", []byte(font_vertex_shader), []byte(font_fragment_shader))
			errChan <- render.RegisterShader("glop.test", []byte(test_vshader), []byte(test_fshader))
			v, _ := render.GetAttribLocation("glop.test", "position")
			l.Printf("position: %v", v)
			v, _ = render.GetAttribLocation("glop.test", "texCoord")
			l.Printf("texCoord: %v", v)
			v, _ = render.GetUniformLocation("glop.test", "tex")
			l.Printf("tex: %v", v)
		})
	})
	// err1 := <-errChan
	err2 := <-errChan
	// if err1 != nil {
	// 	return nil, err1
	// }
	if err2 != nil {
		return nil, err2
	}

	var d Dictionary
	err := gob.NewDecoder(r).Decode(&d.data)
	if err != nil {
		return nil, err
	}
	render.Queue(func() {
		gl.GenTextures(1, &d.tex)
		gl.BindTexture(gl.TEXTURE_2D, d.tex)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RED,
			d.data.Dx,
			d.data.Dy,
			0,
			gl.RED,
			gl.UNSIGNED_BYTE,
			gl.Ptr(&d.data.Pix[0]))
		f := gl.GetError()
		if f != 0 {
			l.Printf("Gl error: %v", f)
		}
	})

	render.Queue(func() {
		gl.GenVertexArrays(2, &d.vaos[0])
		f := gl.GetError()
		if f != 0 {
			l.Printf("Gl error: %v", f)
		}
		gl.GenBuffers(4, &d.vbos[0])
		l.Printf("%v %v", d.vaos, d.vbos)
		if d.vaos[0] == 0 || d.vaos[1] == 0 || d.vaos[0] == d.vaos[1] {
			panic("SDF")
		}
		// Setup whole triangle
		gl.BindVertexArray(d.vaos[0])
		f = gl.GetError()
		if f != 0 {
			l.Printf("Gl error: %v", f)
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[0])
		f = gl.GetError()
		if f != 0 {
			l.Printf("Gl error: %v", f)
		}
		gl.BufferData(gl.ARRAY_BUFFER, len(triangle)*int(unsafe.Sizeof(triangle[0])), gl.Ptr(&triangle[0]), gl.STATIC_DRAW)
		location, _ := render.GetAttribLocation("glop.test", "position")
		gl.EnableVertexAttribArray(uint32(location))
		gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
		l.Printf("Setting position at %v", location)

		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[1])
		f = gl.GetError()
		if f != 0 {
			l.Printf("Gl error: %v", f)
		}
		gl.BufferData(gl.ARRAY_BUFFER, len(triangleTexCoords)*int(unsafe.Sizeof(triangleTexCoords[0])), gl.Ptr(&triangleTexCoords[0]), gl.STATIC_DRAW)
		location, _ = render.GetAttribLocation("glop.test", "texCoord")
		gl.EnableVertexAttribArray(uint32(location))
		gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
		l.Printf("Setting texCoord at %v", location)

		// Setup whole quad
		gl.BindVertexArray(d.vaos[1])
		f = gl.GetError()
		if f != 0 {
			l.Printf("Gl error: %v", f)
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[2])
		f = gl.GetError()
		if f != 0 {
			l.Printf("Gl error: %v", f)
		}
		gl.BufferData(gl.ARRAY_BUFFER, len(quad)*int(unsafe.Sizeof(quad[0])), gl.Ptr(&quad[0]), gl.STATIC_DRAW)
		location, _ = render.GetAttribLocation("glop.test", "position")
		gl.EnableVertexAttribArray(uint32(location))
		gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[3])
		f = gl.GetError()
		if f != 0 {
			l.Printf("Gl error: %v", f)
		}
		gl.BufferData(gl.ARRAY_BUFFER, len(quadTexCoords)*int(unsafe.Sizeof(quadTexCoords[0])), gl.Ptr(&quadTexCoords[0]), gl.STATIC_DRAW)
		location, _ = render.GetAttribLocation("glop.test", "texCoord")
		gl.EnableVertexAttribArray(uint32(location))
		gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

		// gl.BindBuffer(gl.ARRAY_BUFFER, 0)

		// gl.ActiveTexture(gl.TEXTURE0)
		// f = gl.GetError()
		// if f != 0 {
		// 	l.Printf("Gl error: %v", f)
		// }
		// gl.BindTexture(gl.TEXTURE_2D, d.tex)
		// f = gl.GetError()
		// if f != 0 {
		// 	l.Printf("Gl error: %v", f)
		// }
	})
	render.Purge()
	return &d, nil
}

func (d *Dictionary) RenderString(s string, x, y, z, height float64, l *log.Logger) {
	render.EnableShader("glop.test")
	defer render.EnableShader("")

	location, err := render.GetUniformLocation("glop.test", "tex")
	if err != nil {
		l.Printf("Error: %v", err)
	}
	gl.Uniform1i(location, int32(d.tex))
	gl.ActiveTexture(gl.TEXTURE0 + uint32(location))
	gl.BindTexture(gl.TEXTURE_2D, d.tex)
	l.Printf("Setting %v to %v", location, d.tex)
	f := gl.GetError()
	if f != 0 {
		l.Printf("Gl error: %v", f)
	}

	gl.BindVertexArray(d.vaos[0])
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.BindVertexArray(d.vaos[1])
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}
