package text

import (
	// "code.google.com/p/freetype-go/freetype"
	// "fmt"
	"github.com/errcw/glow/gl-core/3.3/gl"
	"github.com/runningwild/glop/render"
	// "github.com/runningwild/glop/text/tool/glyph"
	"image"
	// "io"
	// "io/ioutil"
	"log"
	"sync"
	// "unsafe"
)

const test_vshader = `
#version 330
in vec3 position;
in vec2 texCoord;

out vec3 theColor;
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
uniform float band;
out vec4 fragColor;
void main() {
	vec4 df = texture(tex, theTexCoord);
	float alpha = smoothstep(0.5-band, 0.5+band, df.r);
	fragColor = vec4(1.0, 1.0, 1.0, alpha);
	return;
}
`

type dictData struct {
	// The Pix data from the original image.Rgba
	Pix []byte

	Kerning map[rune]map[rune]int

	// Dx and Dy of the original image.Rgba
	Dx, Dy int32

	// Map from rune to that rune's RuneInfo.
	Info map[rune]RuneInfo

	// RuneInfo for all r < 256 will be stored here as well as in info so we can
	// avoid map lookups if possible.
	Ascii_info []RuneInfo

	// At what vertical value is the line on which text is logically rendered.
	// This is determined by the positioning of the '.' rune.
	Baseline int

	// Amount glyphs were scaled down during packing.
	Scale float64

	Miny, Maxy int
}
type Dictionary2 struct {
	data dictData

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
type Dictionary struct {
	Runes   map[rune]RuneInfo
	Kerning map[RunePair]int

	// Atlas encoded in png format
	Pix []byte
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
// 			errChan <- render.RegisterShader("glop.test", []byte(test_vshader), []byte(test_fshader))
// 			v, _ := render.GetAttribLocation("glop.test", "position")
// 			l.Printf("position: %v", v)
// 			v, _ = render.GetAttribLocation("glop.test", "color")
// 			l.Printf("color: %v", v)
// 			v, _ = render.GetAttribLocation("glop.test", "texCoord")
// 			l.Printf("texCoord: %v", v)
// 			v, _ = render.GetUniformLocation("glop.test", "tex")
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

// 		location, err := render.GetUniformLocation("glop.test", "band")
// 		l.Printf("Location: %v", location)
// 		if err != nil {
// 			l.Printf("Error: %v", err)
// 		}
// 		render.EnableShader("glop.test")
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
// 			location, _ := render.GetAttribLocation("glop.test", "position")
// 			gl.EnableVertexAttribArray(uint32(location))
// 			gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
// 		}
// 		{
// 			gl.BindBuffer(gl.ARRAY_BUFFER, d.strBo[1])
// 			gl.BufferData(gl.ARRAY_BUFFER, len(d.strTex)*int(unsafe.Sizeof(d.strTex[0])), gl.Ptr(&d.strTex[0]), gl.STATIC_DRAW)
// 			location, _ := render.GetAttribLocation("glop.test", "texCoord")
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
// 			location, _ := render.GetAttribLocation("glop.test", "position")
// 			gl.EnableVertexAttribArray(uint32(location))
// 			gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))
// 		}
// 		{
// 			gl.BindBuffer(gl.ARRAY_BUFFER, d.strBo[1])
// 			gl.BufferData(gl.ARRAY_BUFFER, len(d.strTex)*int(unsafe.Sizeof(d.strTex[0])), gl.Ptr(&d.strTex[0]), gl.STATIC_DRAW)
// 			location, _ := render.GetAttribLocation("glop.test", "texCoord")
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
// 		location, _ := render.GetAttribLocation("glop.test", "position")
// 		gl.EnableVertexAttribArray(uint32(location))
// 		gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

// 		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[3])
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.BufferData(gl.ARRAY_BUFFER, len(quadTexCoords)*int(unsafe.Sizeof(quadTexCoords[0])), gl.Ptr(&quadTexCoords[0]), gl.STATIC_DRAW)
// 		location, _ = render.GetAttribLocation("glop.test", "texCoord")
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
// 		location, _ = render.GetAttribLocation("glop.test", "position")
// 		gl.EnableVertexAttribArray(uint32(location))
// 		gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

// 		gl.BindBuffer(gl.ARRAY_BUFFER, d.vbos[1])
// 		f = gl.GetError()
// 		if f != 0 {
// 			l.Printf("Gl error: %v", f)
// 		}
// 		gl.BufferData(gl.ARRAY_BUFFER, len(triangleTexCoords)*int(unsafe.Sizeof(triangleTexCoords[0])), gl.Ptr(&triangleTexCoords[0]), gl.STATIC_DRAW)
// 		location, _ = render.GetAttribLocation("glop.test", "texCoord")
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
	render.EnableShader("glop.test")
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
	location, err := render.GetUniformLocation("glop.test", "tex")
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
