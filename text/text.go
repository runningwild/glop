package text

import (
	"encoding/gob"
	"fmt"
	"github.com/errcw/glow/gl-core/3.3/gl"
	"github.com/runningwild/glop/render"
	"image"
	"io"
	"log"
	"sync"
	"unsafe"
)

const font_vshader = `
#version 330
in vec3 position;
in vec2 texCoord;

uniform float height;
uniform vec2 pen;
uniform vec2 screen;

out vec3 theColor;
out vec2 theTexCoord;

void main() {
   vec2 p2 = vec2(position.x / screen.x, position.y / screen.y);
   p2 *= height;
   p2 *= 2.0;
   p2 -= vec2(1.0, 1.0);

   vec2 pen2 = vec2(pen.x / screen.x, pen.y / screen.y);
   pen2 *= 2.0;
   p2 += pen2;

   gl_Position = vec4(p2.xy, position.z, 1.0);
   theTexCoord = texCoord;
}
`

const font_fshader = `
#version 330
in vec2 theTexCoord;
uniform sampler2D tex;
const float band = 0.025;
const float low = 0.5 - band;
const float high = 0.5 + band;
out vec4 fragColor;

float weight(float a, float b) {
	if (b < a) {
		float buf = b;
		b = a;
		a = buf;
	}
	if (b < low) {
		return 0.0;
	}
	if (a > high) {
		return 1.0;
	}

	float midVal = (smoothstep(low, high, a) + smoothstep(low, high, b)) / 2.0;
	if (a >= low && b <= high) {
		return midVal;
	}
	float midWeight = b - a;

	float lowWeight = 0.0;
	if (a < low) {
		lowWeight = low - a;
		a = low;
	}
	float highWeight = 0.0;
	if (b > high) {
		highWeight = b - high;
		b = high;
	}
	return (midVal * midWeight + 1.0 * highWeight) / (lowWeight + highWeight + midWeight);
}

void main() {
	vec4 t = texture(tex, theTexCoord);
	float v = t.r;
	float vdx = dFdx(v);
	float vdy = dFdy(v);
	fragColor = vec4(1.0, 1.0, 1.0, (weight(v, v+vdx) + weight(v, v+vdy)) / 2.0);
	return;
}
`

type Dictionary2 struct {
	vaos    [2]uint32
	vbos    [4]uint32
	tex     uint32
	sampler uint32

	strAo  uint32
	strBo  [2]uint32 // pos and tex
	strPos []float32
	strTex []float32
}
type RunePair struct {
	A, B rune
}
type RuneInfo struct {
	PixBounds       image.Rectangle
	GlyphBounds     image.Rectangle
	AdvanceHeight   int
	TopSideBearing  int
	AdvanceWidth    int
	LeftSideBearing int
}
type strData struct {
	varrays  [1]uint32
	vbuffers [2]uint32
	count    int32
}
type Dictionary struct {
	Runes   map[rune]RuneInfo
	Kerning map[RunePair]int

	// Width and height of the atlas
	Dx, Dy int32

	// Greyscale bytes of the atlas
	Pix []byte

	// Maximum bounds of any glyphs.
	GlyphMax image.Rectangle

	// atlas texture and sampler
	atlas struct {
		texture  uint32
		sampler  uint32
		varrays  [1]uint32
		vbuffers [2]uint32 // position, tex coord
	}

	strs map[string]strData
}

func LoadDictionary(r io.Reader, l *log.Logger) (*Dictionary, error) {
	errChan := make(chan error)
	init_once.Do(func() {
		render.Queue(func() {
			// errChan <- render.RegisterShader("glop.font", []byte(font_vertex_shader), []byte(font_fragment_shader))
			errChan <- render.RegisterShader("glop.font", []byte(font_vshader), []byte(font_fshader))
			v, _ := render.GetAttribLocation("glop.font", "position")
			l.Printf("position: %v", v)
			v, _ = render.GetAttribLocation("glop.font", "color")
			l.Printf("color: %v", v)
			v, _ = render.GetAttribLocation("glop.font", "texCoord")
			l.Printf("texCoord: %v", v)
			v, _ = render.GetUniformLocation("glop.font", "tex")
			l.Printf("tex: %v", v)
		})
	})
	// err1 := <-errChan
	err2 := <-errChan
	// if err1 != nil {
	// 	return nil, err1
	// }
	l.Printf("%v", err2)
	if err2 != nil {
		return nil, err2
	}

	var dict Dictionary
	dec := gob.NewDecoder(r)
	err := dec.Decode(&dict)
	if err != nil {
		return nil, err
	}

	render.Queue(func() {
		// Create the gl texture for the atlas
		gl.GenTextures(1, &dict.atlas.texture)
		glerr := gl.GetError()
		if glerr != 0 {
			errChan <- fmt.Errorf("Gl Error on gl.GenTextures: %v", glerr)
			return
		}

		// Send the atlas to opengl
		gl.BindTexture(gl.TEXTURE_2D, dict.atlas.texture)
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RED,
			dict.Dx,
			dict.Dy,
			0,
			gl.RED,
			gl.UNSIGNED_BYTE,
			gl.Ptr(&dict.Pix[0]))
		glerr = gl.GetError()
		if glerr != 0 {
			errChan <- fmt.Errorf("Gl Error on creating texture: %v", glerr)
			return
		}

		// Create the atlas sampler and set the parameters we want for it
		gl.GenSamplers(1, &dict.atlas.sampler)
		gl.SamplerParameteri(dict.atlas.sampler, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.SamplerParameteri(dict.atlas.sampler, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.SamplerParameteri(dict.atlas.sampler, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.SamplerParameteri(dict.atlas.sampler, gl.TEXTURE_WRAP_T, gl.REPEAT)
		glerr = gl.GetError()
		if glerr != 0 {
			errChan <- fmt.Errorf("Gl Error on creating sampler: %v", glerr)
			return
		}

		// Create some positions and tex coords
		{
			gl.GenVertexArrays(1, &dict.atlas.varrays[0])
			gl.BindVertexArray(dict.atlas.varrays[0])

			gl.GenBuffers(2, &dict.atlas.vbuffers[0])
			var quad = [18]float32{
				-5.5, -1, 0,
				-5.5, 1, 0,
				5.5, 1, 0,

				-5.5, -1, 0,
				5.5, 1, 0,
				5.5, -1, 0,
			}
			gl.BindBuffer(gl.ARRAY_BUFFER, dict.atlas.vbuffers[0])
			gl.BufferData(gl.ARRAY_BUFFER, len(quad)*int(unsafe.Sizeof(quad[0])), gl.Ptr(&quad[0]), gl.STATIC_DRAW)
			location, err := render.GetAttribLocation("glop.font", "position")
			if err != nil {
				errChan <- err
				return
			}
			l.Printf("Position: %v", location)
			gl.EnableVertexAttribArray(uint32(location))
			gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
			glerr := gl.GetError()
			if glerr != 0 {
				errChan <- fmt.Errorf("Gl Error on creating position buffer setup: %v", glerr)
				return
			}

			var quadTexCoords = [12]float32{
				0, 1,
				0, 0,
				1, 0,

				0, 1,
				1, 0,
				1, 1,
			}
			gl.BindBuffer(gl.ARRAY_BUFFER, dict.atlas.vbuffers[1])
			gl.BufferData(gl.ARRAY_BUFFER, len(quadTexCoords)*int(unsafe.Sizeof(quadTexCoords[0])), gl.Ptr(&quadTexCoords[0]), gl.STATIC_DRAW)
			location, err = render.GetAttribLocation("glop.font", "texCoord")
			if err != nil {
				errChan <- err
				return
			}
			gl.EnableVertexAttribArray(uint32(location))
			gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

			glerr = gl.GetError()
			if glerr != 0 {
				errChan <- fmt.Errorf("Gl Error on creating tex coord buffer setup: %v", glerr)
				return
			}
		}

		errChan <- nil
	})
	err = <-errChan
	if err != nil {
		return nil, err
	}

	return &dict, nil
}

type pos struct {
	x, y float32
}

func (d *Dictionary) bindString(str string, l *log.Logger) strData {
	var data strData
	gl.GenVertexArrays(1, &data.varrays[0])
	gl.BindVertexArray(data.varrays[0])
	gl.GenBuffers(2, &data.vbuffers[0])

	var positions, texcoords []float32

	var pen pos
	var prev rune
	for _, r := range str {
		ri := d.Runes[r]

		var scale float32 = 1.0 / float32(d.GlyphMax.Dy())

		var posMin, posMax pos
		posMin.x = pen.x + float32(ri.GlyphBounds.Min.X)*scale
		posMin.y = pen.y + float32(ri.GlyphBounds.Min.Y)*scale
		posMax.x = pen.x + float32(ri.GlyphBounds.Max.X)*scale
		posMax.y = pen.y + float32(ri.GlyphBounds.Max.Y)*scale

		var texMin, texMax pos
		texMin.x = float32(ri.PixBounds.Min.X) / float32(d.Dx)
		texMin.y = float32(ri.PixBounds.Min.Y) / float32(d.Dy)
		texMax.x = float32(ri.PixBounds.Max.X) / float32(d.Dx)
		texMax.y = float32(ri.PixBounds.Max.Y) / float32(d.Dy)
		pen.x += float32(ri.AdvanceWidth) * scale
		pen.x += float32(d.Kerning[RunePair{prev, r}]) * scale
		//pen.x -= float32(d.Kerning[RunePair{prev, r}]) * scale

		positions = append(positions, posMin.x) // lower left
		positions = append(positions, posMin.y)
		positions = append(positions, 0)
		positions = append(positions, posMin.x) // upper left
		positions = append(positions, posMax.y)
		positions = append(positions, 0)
		positions = append(positions, posMax.x) // upper right
		positions = append(positions, posMax.y)
		positions = append(positions, 0)
		positions = append(positions, posMin.x) // lower left
		positions = append(positions, posMin.y)
		positions = append(positions, 0)
		positions = append(positions, posMax.x) // upper right
		positions = append(positions, posMax.y)
		positions = append(positions, 0)
		positions = append(positions, posMax.x) // lower right
		positions = append(positions, posMin.y)
		positions = append(positions, 0)
		l.Printf("Rune(%c): %v", r, ri.PixBounds)
		texcoords = append(texcoords, texMin.x) // lower left
		texcoords = append(texcoords, texMax.y)
		texcoords = append(texcoords, texMin.x) // upper left
		texcoords = append(texcoords, texMin.y)
		texcoords = append(texcoords, texMax.x) // upper right
		texcoords = append(texcoords, texMin.y)
		texcoords = append(texcoords, texMin.x) // lower left
		texcoords = append(texcoords, texMax.y)
		texcoords = append(texcoords, texMax.x) // upper right
		texcoords = append(texcoords, texMin.y)
		texcoords = append(texcoords, texMax.x) // lower right
		texcoords = append(texcoords, texMax.y)

		prev = r
	}
	l.Printf("Positions: %v\n", positions)
	data.count = int32(len(positions))
	gl.BindBuffer(gl.ARRAY_BUFFER, data.vbuffers[0])
	gl.BufferData(gl.ARRAY_BUFFER, len(positions)*int(unsafe.Sizeof(positions[0])), gl.Ptr(&positions[0]), gl.STATIC_DRAW)
	location, _ := render.GetAttribLocation("glop.font", "position")
	gl.EnableVertexAttribArray(uint32(location))
	gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

	l.Printf("Positions: %v\n", texcoords)
	gl.BindBuffer(gl.ARRAY_BUFFER, data.vbuffers[1])
	gl.BufferData(gl.ARRAY_BUFFER, len(texcoords)*int(unsafe.Sizeof(texcoords[0])), gl.Ptr(&texcoords[0]), gl.STATIC_DRAW)
	location, _ = render.GetAttribLocation("glop.font", "texCoord")
	gl.EnableVertexAttribArray(uint32(location))
	gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

	return data
}

// RenderString must be called on the render thread.
func (d *Dictionary) RenderString(str string, x, y, z, height float64, l *log.Logger) {
	// No synchronization necessary because everything is run serially on the render thread anyway.
	if d.strs == nil {
		d.strs = make(map[string]strData)
	}
	data, ok := d.strs[str]
	if !ok {
		data = d.bindString(str, l)
		d.strs[str] = data
	}

	render.EnableShader("glop.font")
	defer render.EnableShader("")

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, d.atlas.texture)
	location, _ := render.GetUniformLocation("glop.font", "tex")
	gl.Uniform1i(location, 0)
	gl.BindSampler(0, d.atlas.sampler)

	location, _ = render.GetUniformLocation("glop.font", "height")
	gl.Uniform1f(location, float32(height))

	var viewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &viewport[0])
	location, _ = render.GetUniformLocation("glop.font", "screen")
	gl.Uniform2f(location, float32(viewport[2]), float32(viewport[3]))

	location, _ = render.GetUniformLocation("glop.font", "pen")
	gl.Uniform2f(location, float32(x)+float32(viewport[0]), float32(y)+float32(viewport[1]))

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.BindVertexArray(data.varrays[0])
	gl.DrawArrays(gl.TRIANGLES, 0, data.count)
}
func foo() {
	// f := gl.GetError()
	// if f != 0 {
	// 	l.Printf("Gl error: %v", f)
	// }
	// render.EnableShader("glop.font")
	// defer render.EnableShader("")
	// f = gl.GetError()
	// if f != 0 {
	// 	l.Printf("Gl error: %v", f)
	// }

	// gl.ActiveTexture(gl.TEXTURE0)
	// f = gl.GetError()
	// if f != 0 {
	// 	l.Printf("Gl error: %v", f)
	// }
	// gl.BindTexture(gl.TEXTURE_2D, d.atlas.texture)
	// f = gl.GetError()
	// if f != 0 {
	// 	l.Printf("Gl error: %v", f)
	// }
	// location, err := render.GetUniformLocation("glop.font", "tex")
	// if err != nil {
	// 	l.Printf("Error: %v", err)
	// }
	// gl.Uniform1i(location, 0)
	// f = gl.GetError()
	// if f != 0 {
	// 	l.Printf("Gl error: %v", f)
	// }
	// gl.BindSampler(0, d.atlas.sampler)
	// f = gl.GetError()
	// if f != 0 {
	// 	l.Printf("Gl error: %v", f)
	// }

	// location, err = render.GetUniformLocation("glop.font", "height")
	// if err != nil {
	// 	l.Printf("Error: %v", err)
	// }
	// gl.Uniform1f(location, float32(height))

	// gl.Enable(gl.BLEND)
	// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	// gl.BindVertexArray(d.atlas.varrays[0])
	// gl.DrawArrays(gl.TRIANGLES, 0, 6*3)
	// f = gl.GetError()
	// if f != 0 {
	// 	l.Printf("Gl error: %v", f)
	// }
	// return
	// // gl.BindVertexArray(d.vaos[0])
	// // gl.DrawArrays(gl.TRIANGLES, 0, 3)
	// gl.BindVertexArray(d.atlas.varrays[0])
	// gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

var triangle = [9]float32{-0.4, 0.1, 0.0, 0.4, 0.1, 0.0, 0.3, 0.7, 0.0}
var triangleColor = [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1}
var triangleTexCoords = [6]float32{0, 1, 1, 1, 1, 0}
var quad = [12]float32{-0.2, -0.1, 0.0, -0.2, -0.6, 0.0, 0.2, -0.1, 0.0, 0.2, -0.6, 0.0}
var quadColor = [12]float32{1, 1, 1, 0, 0, 1, 0, 1, 0, 1, 0, 0}
var quadTexCoords = [8]float32{0, 0, 0, 1, 1, 1, 1, 0}

var init_once sync.Once

// func LoadDictionary2(r io.Reader, l *log.Logger) (*Dictionary2, error) {
// 	errChan := make(chan error)
// 	init_once.Do(func() {
// 		render.Queue(func() {
// 			// errChan <- render.RegisterShader("glop.font", []byte(font_vertex_shader), []byte(font_fragment_shader))
// 			errChan <- render.RegisterShader("glop.font", []byte(test_vshader), []byte(test_fshader))
// 			v, _ := render.GetAttribLocation("glop.font", "position")
// 			l.Printf("position: %v", v)
// 			v, _ = render.GetAttribLocation("glop.font", "color")
// 			l.Printf("color: %v", v)
// 			v, _ = render.GetAttribLocation("glop.font", "texCoord")
// 			l.Printf("texCoord: %v", v)
// 			v, _ = render.GetUniformLocation("glop.font", "tex")
// 			l.Printf("tex: %v", v)
// 		})
// 	})
// 	// err1 := <-errChan
// 	err2 := <-errChan
// 	// if err1 != nil {
// 	// 	return nil, err1
// 	// }
// 	l.Printf("%v", err2)
// 	if err2 != nil {
// 		return nil, err2
// 	}

// 	var dx, dy int
// 	var pix []byte
// 	{
// 		l.Printf("erer")
// 		data, err := ioutil.ReadAll(r)
// 		if err != nil {
// 			return nil, fmt.Errorf("Unable to read file: %v", err)
// 		}
// 		l.Printf("erer")
// 		font, err := freetype.ParseFont(data)
// 		if err != nil {
// 			return nil, fmt.Errorf("Unable to parse font file: %v", err)
// 		}
// 		l.Printf("erer")
// 		d, err := glyph.Render(font, '@', 1000, 0.3)
// 		if err != nil {
// 			return nil, fmt.Errorf("Unable to parse font file: %v", err)
// 		}
// 	}
// 	l.Printf("erer")
// 	var d Dictionary2
// 	render.Queue(func() {
// 		gl.GenTextures(1, &d.tex)
// 		f := gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.BindTexture(gl.TEXTURE_2D, d.tex)
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		d.data.Dx = 32
// 		d.data.Dy = 32
// 		d.data.Pix = make([]byte, d.data.Dx*d.data.Dy)
// 		for y := 0; y < dy; y++ {
// 			for x := 0; x < dx; x++ {
// 				srcPos := x + y*dx
// 				dstPos := x + y*int(d.data.Dx)
// 				d.data.Pix[dstPos] = pix[srcPos]
// 			}
// 		}
// 		gl.TexImage2D(
// 			gl.TEXTURE_2D,
// 			0,
// 			gl.RED,
// 			d.data.Dx,
// 			d.data.Dy,
// 			0,
// 			gl.RED,
// 			gl.UNSIGNED_BYTE,
// 			gl.Ptr(&d.data.Pix[0]))
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}

// 		gl.GenSamplers(1, &d.sampler)
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.SamplerParameteri(d.sampler, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
// 		gl.SamplerParameteri(d.sampler, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
// 		gl.SamplerParameteri(d.sampler, gl.TEXTURE_WRAP_S, gl.REPEAT)
// 		gl.SamplerParameteri(d.sampler, gl.TEXTURE_WRAP_T, gl.REPEAT)
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 	})

// 	render.Queue(func() {
// 		gl.GenVertexArrays(2, &d.vaos[0])
// 		f := gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.GenBuffers(4, &d.vbos[0])
// 		l.Printf("%v %v", d.vaos, d.vbos)
// 		if d.vaos[0] == 0 || d.vaos[1] == 0 || d.vaos[0] == d.vaos[1] {
// 			panic("SDF")
// 		}

// 		var x0 float32 = -0.4
// 		var y0 float32 = -1
// 		var size float32 = 0.3
// 		var x1 float32 = x0 + size
// 		var y1 float32 = y0 + size

// 		location, err := render.GetUniformLocation("glop.font", "band")
// 		l.Printf("Location: %v", location)
// 		if err != nil {
// 			l.Printf("Error: %v", err)
// 		}
// 		render.EnableShader("glop.font")
// 		band := 0.5 / (size * 10)
// 		if band > 0.5 {
// 			band = 0.5
// 		}
// 		gl.Uniform1f(location, band)
// 		render.EnableShader("")
// 		errVal := gl.GetError()
// 		if errVal != 0 {
// 			l.Printf("Error: %v", errVal)
// 		}

// 		d.strPos = append(d.strPos, x0)
// 		d.strPos = append(d.strPos, y0)
// 		d.strPos = append(d.strPos, x0)
// 		d.strPos = append(d.strPos, y1)
// 		d.strPos = append(d.strPos, x1)
// 		d.strPos = append(d.strPos, y1)

// 		d.strPos = append(d.strPos, x1)
// 		d.strPos = append(d.strPos, y1)
// 		d.strPos = append(d.strPos, x1)
// 		d.strPos = append(d.strPos, y0)
// 		d.strPos = append(d.strPos, x0)
// 		d.strPos = append(d.strPos, y0)

// 		d.strTex = append(d.strTex, 0)
// 		d.strTex = append(d.strTex, 0)
// 		d.strTex = append(d.strTex, 0)
// 		d.strTex = append(d.strTex, -1)
// 		d.strTex = append(d.strTex, 1)
// 		d.strTex = append(d.strTex, -1)

// 		d.strTex = append(d.strTex, 1)
// 		d.strTex = append(d.strTex, -1)
// 		d.strTex = append(d.strTex, 1)
// 		d.strTex = append(d.strTex, 0)
// 		d.strTex = append(d.strTex, 0)
// 		d.strTex = append(d.strTex, 0)

// 		gl.GenVertexArrays(1, &d.strAo)
// 		gl.GenBuffers(2, &d.strBo[0])
// 		gl.BindVertexArray(d.strAo)

// 		{
// 			gl.BindBuffer(gl.ARRAY_BUFFER, d.strBo[0])
// 			gl.BufferData(gl.ARRAY_BUFFER, len(d.strPos)*int(unsafe.Sizeof(d.strPos[0])), gl.Ptr(&d.strPos[0]), gl.STATIC_DRAW)
// 			location, _ := render.GetAttribLocation("glop.font", "position")
// 			gl.EnableVertexAttribArray(uint32(location))
// 			gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
// 		}
// 		{
// 			gl.BindBuffer(gl.ARRAY_BUFFER, d.strBo[1])
// 			gl.BufferData(gl.ARRAY_BUFFER, len(d.strTex)*int(unsafe.Sizeof(d.strTex[0])), gl.Ptr(&d.strTex[0]), gl.STATIC_DRAW)
// 			location, _ := render.GetAttribLocation("glop.font", "texCoord")
// 			gl.EnableVertexAttribArray(uint32(location))
// 			gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
// 		}
// 	})
// 	render.Purge()
// 	return &d, nil
// 	render.Queue(func() {
// 		gl.GenVertexArrays(2, &d.vaos[0])
// 		f := gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.GenBuffers(4, &d.vbos[0])
// 		l.Printf("%v %v", d.vaos, d.vbos)
// 		if d.vaos[0] == 0 || d.vaos[1] == 0 || d.vaos[0] == d.vaos[1] {
// 			panic("SDF")
// 		}
// 		str := "Blarg!"
// 		var xPos float32 = -1
// 		var scale float32 = 0.003
// 		var prev rune = 0
// 		for _, r := range str {
// 			info := d.data.Info[r]
// 			xleft := xPos + float32(info.Full_bounds.Min.X)*scale
// 			xright := xPos + float32(info.Full_bounds.Max.X)*scale
// 			ytop := (float32(info.Full_bounds.Max.Y) - float32(d.data.Miny)) * scale
// 			ybot := (float32(info.Full_bounds.Min.Y) - float32(d.data.Miny)) * scale
// 			// ytop -= ybot
// 			ybot = 0
// 			xPos += float32(info.Advance) * scale
// 			if _, ok := d.data.Kerning[prev]; ok {
// 				xPos -= float32(d.data.Kerning[prev][r]) * scale
// 			}
// 			prev = r
// 			d.strPos = append(d.strPos, xleft)
// 			d.strPos = append(d.strPos, ybot)
// 			d.strPos = append(d.strPos, xleft)
// 			d.strPos = append(d.strPos, ytop)
// 			d.strPos = append(d.strPos, xright)
// 			d.strPos = append(d.strPos, ybot)
// 			d.strPos = append(d.strPos, xleft)
// 			d.strPos = append(d.strPos, ytop)
// 			d.strPos = append(d.strPos, xright)
// 			d.strPos = append(d.strPos, ybot)
// 			d.strPos = append(d.strPos, xright)
// 			d.strPos = append(d.strPos, ytop)

// 			// d.strPos = []float32{0.0, 0.0, 1.0, 0.0, 1.0, 1.0, 0.0, 1.0}
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Min.X)/float32(d.data.Dx))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Max.Y)/float32(d.data.Dy))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Min.X)/float32(d.data.Dx))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Min.Y)/float32(d.data.Dy))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Max.X)/float32(d.data.Dx))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Max.Y)/float32(d.data.Dy))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Min.X)/float32(d.data.Dx))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Min.Y)/float32(d.data.Dy))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Max.X)/float32(d.data.Dx))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Max.Y)/float32(d.data.Dy))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Max.X)/float32(d.data.Dx))
// 			d.strTex = append(d.strTex, float32(d.data.Info[r].Pos.Min.Y)/float32(d.data.Dy))
// 		}
// 		gl.GenVertexArrays(1, &d.strAo)
// 		gl.GenBuffers(2, &d.strBo[0])
// 		gl.BindVertexArray(d.strAo)

// 		{
// 			gl.BindBuffer(gl.ARRAY_BUFFER, d.strBo[0])
// 			gl.BufferData(gl.ARRAY_BUFFER, len(d.strPos)*int(unsafe.Sizeof(d.strPos[0])), gl.Ptr(&d.strPos[0]), gl.STATIC_DRAW)
// 			location, _ := render.GetAttribLocation("glop.font", "position")
// 			gl.EnableVertexAttribArray(uint32(location))
// 			gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
// 		}
// 		{
// 			gl.BindBuffer(gl.ARRAY_BUFFER, d.strBo[1])
// 			gl.BufferData(gl.ARRAY_BUFFER, len(d.strTex)*int(unsafe.Sizeof(d.strTex[0])), gl.Ptr(&d.strTex[0]), gl.STATIC_DRAW)
// 			location, _ := render.GetAttribLocation("glop.font", "texCoord")
// 			gl.EnableVertexAttribArray(uint32(location))
// 			gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
// 		}

// 		return
// 		// Setup whole quad
// 		gl.BindVertexArray(d.vaos[1])
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}

// 		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[2])
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.BufferData(gl.ARRAY_BUFFER, len(quad)*int(unsafe.Sizeof(quad[0])), gl.Ptr(&quad[0]), gl.STATIC_DRAW)
// 		location, _ := render.GetAttribLocation("glop.font", "position")
// 		gl.EnableVertexAttribArray(uint32(location))
// 		gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

// 		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[3])
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.BufferData(gl.ARRAY_BUFFER, len(quadTexCoords)*int(unsafe.Sizeof(quadTexCoords[0])), gl.Ptr(&quadTexCoords[0]), gl.STATIC_DRAW)
// 		location, _ = render.GetAttribLocation("glop.font", "texCoord")
// 		gl.EnableVertexAttribArray(uint32(location))
// 		gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

// 		// Setup whole triangle
// 		gl.BindVertexArray(d.vaos[0])
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}

// 		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[0])
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.BufferData(gl.ARRAY_BUFFER, len(triangle)*int(unsafe.Sizeof(triangle[0])), gl.Ptr(&triangle[0]), gl.STATIC_DRAW)
// 		location, _ = render.GetAttribLocation("glop.font", "position")
// 		gl.EnableVertexAttribArray(uint32(location))
// 		gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

// 		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[1])
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.BufferData(gl.ARRAY_BUFFER, len(triangleTexCoords)*int(unsafe.Sizeof(triangleTexCoords[0])), gl.Ptr(&triangleTexCoords[0]), gl.STATIC_DRAW)
// 		location, _ = render.GetAttribLocation("glop.font", "texCoord")
// 		gl.EnableVertexAttribArray(uint32(location))
// 		gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

// 		gl.BindBuffer(gl.ARRAY_BUFFER, 0)

// 		gl.ActiveTexture(gl.TEXTURE0)
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.BindTexture(gl.TEXTURE_2D, d.tex)
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 	})
// 	render.Purge()
// 	return &d, nil
// }

func (d *Dictionary2) RenderString(s string, x, y, z, height float64, l *log.Logger) {
	l.Printf("Here")
	f := gl.GetError()
	if f != 0 {
		l.Printf("Gl error: %v", f)
	}
	render.EnableShader("glop.font")
	defer render.EnableShader("")
	f = gl.GetError()
	if f != 0 {
		l.Printf("Gl error: %v", f)
	}

	gl.ActiveTexture(gl.TEXTURE0)
	f = gl.GetError()
	if f != 0 {
		l.Printf("Gl error: %v", f)
	}
	gl.BindTexture(gl.TEXTURE_2D, d.tex)
	f = gl.GetError()
	if f != 0 {
		l.Printf("Gl error: %v", f)
	}
	location, err := render.GetUniformLocation("glop.font", "tex")
	l.Printf("Location: %v", location)
	if err != nil {
		l.Printf("Error: %v", err)
	}
	gl.Uniform1i(location, 0)
	f = gl.GetError()
	if f != 0 {
		l.Printf("Gl error: %v", f)
	}
	gl.BindSampler(0, d.sampler)
	f = gl.GetError()
	if f != 0 {
		l.Printf("Gl error: %v", f)
	}
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.BindVertexArray(d.strAo)
	gl.DrawArrays(gl.TRIANGLES, 0, 6*6)
	return
	// gl.BindVertexArray(d.vaos[0])
	// gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.BindVertexArray(d.vaos[1])
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}
