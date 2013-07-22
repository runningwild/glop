package gui

import (
	"encoding/gob"
	gl "github.com/chsc/gogl/gl21"
	"github.com/runningwild/glop/render"
	"image"
	"io"
	"math"
	"runtime"
	"sync"
	"unsafe"
)

// Shader stuff - The font stuff requires that we use some simple shaders
const font_vertex_shader = `
	varying vec3 pos;

	void main() {
	  gl_Position = ftransform();
	  gl_ClipVertex = gl_ModelViewMatrix * gl_Vertex;
	  gl_FrontColor = gl_Color;
	  gl_TexCoord[0] = gl_MultiTexCoord0;
	  gl_TexCoord[1] = gl_MultiTexCoord1;
	  pos = gl_Vertex.xyz;
	}
`

const font_fragment_shader = `
	uniform sampler2D tex;
	uniform float dist_min;
	uniform float dist_max;

	void main() {
	  vec2 tpos = gl_TexCoord[0].xy;
	  float dist = texture2D(tex, tpos).a;
	  float alpha = smoothstep(dist_min, dist_max, dist);
	  gl_FragColor = gl_Color * vec4(1.0, 1.0, 1.0, alpha);
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
	Dx, Dy int

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

	texture uint32

	strs map[string]strBuffer
	pars map[string]strBuffer

	dlists map[string]uint32
}
type strBuffer struct {
	vbuffer uint32
	vs      []dictVert

	ibuffer uint32
	is      []uint16
}
type dictVert struct {
	x, y float32
	u, v float32
}

func (d *Dictionary) Scale() float64 {
	return d.data.Scale
}

func (d *Dictionary) getInfo(r rune) runeInfo {
	var info runeInfo
	if r >= 0 && r < 256 {
		info = d.data.Ascii_info[r]
	} else {
		info, _ = d.data.Info[r]
	}
	return info
}

// Figures out how wide a string will be if rendered at its natural size.
func (d *Dictionary) figureWidth(s string) float64 {
	w := 0.0
	for _, r := range s {
		w += d.getInfo(r).Advance
	}
	return w
}

type Justification int

const (
	Center Justification = iota
	Left
	Right
	Top
	Bottom
)

func (d *Dictionary) split(s string, dx, height float64) []string {
	var lines []string
	var line []rune
	var word []rune
	pos := 0.0
	for _, r := range s {
		if r == ' ' {
			if len(line) > 0 {
				line = append(line, ' ')
			}
			for _, r := range word {
				line = append(line, r)
			}
			word = word[0:0]
		} else {
			word = append(word, r)
		}
		pos += d.getInfo(r).Advance
		if pos >= dx {
			pos = 0.0
			for _, r := range word {
				pos += d.getInfo(r).Advance
			}
			lines = append(lines, string(line))
			line = line[0:0]
		}
	}
	if pos < dx {
		if len(line) > 0 {
			line = append(line, ' ')
		}
		for _, r := range word {
			line = append(line, r)
		}
		lines = append(lines, string(line))
	} else {
		lines = append(lines, string(line))
		lines = append(lines, string(word))
	}
	return lines
}

//TODO: This isn't working - not even a little
func (d *Dictionary) RenderParagraph(s string, x, y, z, dx, height float64, halign, valign Justification) {
	lines := d.split(s, dx, height)
	total_height := height * float64(len(lines)-1)
	switch valign {
	case Bottom:
		y += total_height
	case Center:
		y += total_height / 2
	}
	for _, line := range lines {
		d.RenderString(line, x, y, z, height, halign)
		y -= height
	}
}

func (d *Dictionary) StringWidth(s string) float64 {
	width := 0.0
	for _, r := range s {
		info := d.getInfo(r)
		width += info.Advance
	}
	return width
}

func (d *Dictionary) RenderString(s string, x, y, z, height float64, just Justification) {
	if len(s) == 0 {
		return
	}
	strbuf, ok := d.strs[s]
	if !ok {
		defer d.RenderString(s, x, y, z, height, just)
	} else {
		render.EnableShader("glop.font")
		diff := 20/math.Pow(height, 1.0) + 5*math.Pow(d.data.Scale, 1.0)/math.Pow(height, 1.0)
		if diff > 0.2 {
			diff = 0.2
		}
		render.SetUniformF("glop.font", "dist_min", float32(0.5-diff))
		render.SetUniformF("glop.font", "dist_max", float32(0.5+diff))
		defer render.EnableShader("")
	}
	size := unsafe.Sizeof(dictVert{})
	scale := height / float64(d.data.Maxy-d.data.Miny)
	width := float32(d.figureWidth(s) * scale)
	x_pos := float32(x)
	switch just {
	case Center:
		x_pos -= width / 2
	case Right:
		x_pos -= width
	}
	if ok {
		gl.PushMatrix()
		defer gl.PopMatrix()
		gl.Translated(gl.Double(x_pos), gl.Double(y), gl.Double(z))
		gl.Scaled(gl.Double(scale), gl.Double(scale), 1)

		gl.PushAttrib(gl.COLOR_BUFFER_BIT)
		defer gl.PopAttrib()
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

		gl.Enable(gl.TEXTURE_2D)
		gl.BindTexture(gl.TEXTURE_2D, gl.Uint(d.texture))

		gl.BindBuffer(gl.ARRAY_BUFFER, gl.Uint(strbuf.vbuffer))

		gl.EnableClientState(gl.VERTEX_ARRAY)
		gl.VertexPointer(2, gl.FLOAT, gl.Sizei(size), nil)

		gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
		gl.TexCoordPointer(2, gl.FLOAT, gl.Sizei(size), gl.Pointer(unsafe.Offsetof(strbuf.vs[0].u)))

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, gl.Uint(strbuf.ibuffer))
		gl.DrawElements(gl.TRIANGLES, gl.Sizei(len(strbuf.is)), gl.UNSIGNED_SHORT, nil)

		gl.DisableClientState(gl.VERTEX_ARRAY)
		gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
		return
	}
	x_pos = 0
	var prev rune
	for _, r := range s {
		if _, ok := d.data.Kerning[prev]; ok {
			x_pos += float32(d.data.Kerning[prev][r])
		}
		prev = r
		info := d.getInfo(r)
		xleft := x_pos + float32(info.Full_bounds.Min.X)      //- float32(info.Full_bounds.Min.X-info.Bounds.Min.X)
		xright := x_pos + float32(info.Full_bounds.Max.X)     //+ float32(info.Full_bounds.Max.X-info.Bounds.Max.X)
		ytop := float32(d.data.Maxy - info.Full_bounds.Max.Y) //- float32(info.Full_bounds.Min.Y-info.Bounds.Min.Y)
		ybot := float32(d.data.Maxy - info.Full_bounds.Min.Y) //+ float32(info.Full_bounds.Max.X-info.Bounds.Max.X)
		start := uint16(len(strbuf.vs))
		strbuf.is = append(strbuf.is, start+0)
		strbuf.is = append(strbuf.is, start+1)
		strbuf.is = append(strbuf.is, start+2)
		strbuf.is = append(strbuf.is, start+0)
		strbuf.is = append(strbuf.is, start+2)
		strbuf.is = append(strbuf.is, start+3)
		strbuf.vs = append(strbuf.vs, dictVert{
			x: xleft,
			y: ytop,
			u: float32(info.Pos.Min.X) / float32(d.data.Dx),
			v: float32(info.Pos.Max.Y) / float32(d.data.Dy),
		})
		strbuf.vs = append(strbuf.vs, dictVert{
			x: xleft,
			y: ybot,
			u: float32(info.Pos.Min.X) / float32(d.data.Dx),
			v: float32(info.Pos.Min.Y) / float32(d.data.Dy),
		})
		strbuf.vs = append(strbuf.vs, dictVert{
			x: xright,
			y: ybot,
			u: float32(info.Pos.Max.X) / float32(d.data.Dx),
			v: float32(info.Pos.Min.Y) / float32(d.data.Dy),
		})
		strbuf.vs = append(strbuf.vs, dictVert{
			x: xright,
			y: ytop,
			u: float32(info.Pos.Max.X) / float32(d.data.Dx),
			v: float32(info.Pos.Max.Y) / float32(d.data.Dy),
		})
		x_pos += float32(info.Advance) // - float32((info.Full_bounds.Dx() - info.Bounds.Dx()))
	}
	gl.GenBuffers(1, (*gl.Uint)(&strbuf.vbuffer))
	gl.BindBuffer(gl.ARRAY_BUFFER, gl.Uint(strbuf.vbuffer))
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(int(size)*len(strbuf.vs)), gl.Pointer(&strbuf.vs[0].x), gl.STATIC_DRAW)

	gl.GenBuffers(1, (*gl.Uint)(&strbuf.ibuffer))
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, gl.Uint(strbuf.ibuffer))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(int(unsafe.Sizeof(strbuf.is[0]))*len(strbuf.is)), gl.Pointer(&strbuf.is[0]), gl.STATIC_DRAW)
	d.strs[s] = strbuf
}

var init_once sync.Once

func LoadDictionary(r io.Reader) (*Dictionary, error) {
	init_once.Do(func() {
		render.Queue(func() {
			err := render.RegisterShader("glop.font", []byte(font_vertex_shader), []byte(font_fragment_shader))
			if err != nil {
				panic(err)
			}
		})
		render.Purge()
	})

	var d Dictionary
	err := gob.NewDecoder(r).Decode(&d.data)
	if err != nil {
		return nil, err
	}
	d.setupGlStuff()
	return &d, nil
}

// Sets up anything that wouldn't have been loaded from disk, including
// all opengl data, and sets up finalizers for that data.
func (d *Dictionary) setupGlStuff() {
	d.dlists = make(map[string]uint32)
	d.strs = make(map[string]strBuffer)
	d.pars = make(map[string]strBuffer)

	// TODO: This finalizer is untested
	runtime.SetFinalizer(d, func(d *Dictionary) {
		render.Queue(func() {
			for _, v := range d.dlists {
				gl.DeleteLists(gl.Uint(v), 1)
			}
		})
	})

	render.Queue(func() {
		gl.Enable(gl.TEXTURE_2D)
		gl.GenTextures(1, (*gl.Uint)(&d.texture))
		gl.BindTexture(gl.TEXTURE_2D, gl.Uint(d.texture))
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.ALPHA,
			gl.Sizei(d.data.Dx),
			gl.Sizei(d.data.Dy),
			0,
			gl.ALPHA,
			gl.UNSIGNED_BYTE,
			gl.Pointer(&d.data.Pix[0]))

		gl.Disable(gl.TEXTURE_2D)
	})
}
