// Package text supports a few simple functions for doing distance field font rendering.
// Dictionaries are created from truetype font files using the text/tool binary.  One of these .dict
// files can be loaded using text.LoadDictionary(), then font can be rendered using
// dict.RenderString().
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

// Dictionary contains all of the information about a font necessary for rendering it using
// distance field font rendering.
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

var initOnce sync.Once

// LoadDictionary reads a gobbed Dictionary object from r, registers its atlas texture with opengl,
// and returns a Dictionary that is ready to render text.
func LoadDictionary(r io.Reader) (*Dictionary, error) {
	errChan := make(chan error)
	initOnce.Do(func() {
		render.Queue(func() {
			// errChan <- render.RegisterShader("glop.font", []byte(font_vertex_shader), []byte(font_fragment_shader))
			errChan <- render.RegisterShader("glop.font", []byte(font_vshader), []byte(font_fshader))
		})
	})
	err := <-errChan
	if err != nil {
		return nil, err
	}

	var dict Dictionary
	dec := gob.NewDecoder(r)
	err = dec.Decode(&dict)
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

// bindString generates all of the vertex buffers and vertex arrays for a single constant line of
// text.  No error checking is done.
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
		l.Printf("Kerning[%d] (%c, %c): %v", len(d.Kerning), prev, r, float32(d.Kerning[RunePair{prev, r}]))
		// pen.x -= float32(d.Kerning[RunePair{prev, r}]) * scale

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
	data.count = int32(len(positions))
	gl.BindBuffer(gl.ARRAY_BUFFER, data.vbuffers[0])
	gl.BufferData(gl.ARRAY_BUFFER, len(positions)*int(unsafe.Sizeof(positions[0])), gl.Ptr(&positions[0]), gl.STATIC_DRAW)
	location, _ := render.GetAttribLocation("glop.font", "position")
	gl.EnableVertexAttribArray(uint32(location))
	gl.VertexAttribPointer(uint32(location), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))

	gl.BindBuffer(gl.ARRAY_BUFFER, data.vbuffers[1])
	gl.BufferData(gl.ARRAY_BUFFER, len(texcoords)*int(unsafe.Sizeof(texcoords[0])), gl.Ptr(&texcoords[0]), gl.STATIC_DRAW)
	location, _ = render.GetAttribLocation("glop.font", "texCoord")
	gl.EnableVertexAttribArray(uint32(location))
	gl.VertexAttribPointer(uint32(location), 2, gl.FLOAT, false, 0, gl.PtrOffset(0))

	return data
}

// RenderString must be called on the render thread.  x and y are the initial position of the pen,
// in screen coordinates, and height is the height of a full line of text, in screen coordinates.
func (d *Dictionary) RenderString(str string, x, y, height float64, l *log.Logger) {
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
