package sprite

import (
  "bytes"
  "encoding/gob"
  "fmt"
  "hash/fnv"
  "image"
  "image/draw"
  "os"
  "path/filepath"
  "github.com/runningwild/glop/render"
  "github.com/runningwild/opengl/gl"
  "github.com/runningwild/opengl/glu"
  "github.com/runningwild/yedparse"
)

// An id that specifies a specific frame along with its facing.  This is used
// to index into sprite sheets.
type frameId struct {
  facing int
  node   int
}

// A sheet contains a group of frames of animations indexed by frameId
type sheet struct {
  rects map[frameId]FrameRect
  dx,dy int
  path string
  anim *yed.Graph

  // Unique name that is based on the path of the sprite and the list of
  // frameIds used to generate this sheet.  This name is used to store the
  // sheet on disk when not in use.
  name string

  reference_chan chan int
  load_chan chan bool
  references int
  texture gl.Texture
  pix []byte
  image *image.RGBA
}

func (s *sheet) Load() {
  s.reference_chan <- 1
}

func (s *sheet) Unload() {
  s.reference_chan <- -1
}

func (s *sheet) compose() {
  filename := filepath.Join(s.path, s.name)
  f, err := os.Open(filename)
  if err == nil {
    dec := gob.NewDecoder(f)
    err = dec.Decode(&s.pix)
    f.Close()
    if err == nil {
      return
    }
  }
  canvas := image.NewRGBA(image.Rect(0, 0, s.dx, s.dy))
  for fid,rect := range s.rects {
    name := s.anim.Node(fid.node).Line(0) + ".png"
    file,err := os.Open(filepath.Join(s.path, fmt.Sprintf("%d", fid.facing), name))
    // if a file isn't there that's ok
    if err != nil { continue }

    im,_,err := image.Decode(file)
    file.Close()
    // if a file can't be read that is *not* ok, TODO: Log an error or something
    if err != nil { continue }
    draw.Draw(canvas, image.Rect(rect.X, s.dy - rect.Y, rect.X2, s.dy - rect.Y2), im, image.Point{}, draw.Over)
  }
  s.pix = canvas.Pix
  f, err = os.Create(filename)
  if err == nil {
    enc := gob.NewEncoder(f)
    err = enc.Encode(s.pix)
    f.Close()
    if err != nil {
      os.Remove(filename)
    }
  }
}

// TODO: This was copied from the gui package, probably should just have some basic
// texture loading utils that do this common stuff
func nextPowerOf2(n uint32) uint32 {
  if n == 0 {
    return 1
  }
  for i := uint(0); i < 32; i++ {
    p := uint32(1) << i
    if n <= p {
      return p
    }
  }
  return 0
}

func (s *sheet) makeTexture() {
  gl.Enable(gl.TEXTURE_2D)
  s.texture = gl.GenTexture()
  s.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, s.dx, s.dy, gl.RGBA, s.pix)
  s.pix = nil
}

func (s *sheet) loadRoutine() {
  ready := make(chan bool, 1)
  for load := range s.load_chan {
    if load {
      go func() {
        s.compose()
        render.Queue(func() {
          s.makeTexture()
          ready <- true
        })
      } ()
    } else {
      go func() {
        <-ready
        render.Queue(func() {
          s.texture.Delete()
          s.texture = 0
        })
      } ()
    }
  }
}

// TODO: Need to set up a finalizer on this thing so that we don't keep this
// texture memory around forever if we forget about it
func (s *sheet) routine() {
  go s.loadRoutine()
  for load := range s.reference_chan {
    if load < 0 {
      if s.references == 0 {
        panic("Tried to unload a sprite sheet more times than it was loaded")
      }
      s.references--
      if s.references == 0 {
        s.load_chan <- false
      }
    } else if load > 0 {
      if s.references == 0 {
        s.load_chan <- true
      }
      s.references++
    } else {
      panic("value of 0 should never be sent along load_chan")
    }
  }
}

func uniqueName(fids []frameId) (string, error) {
  buf := bytes.NewBuffer(nil)
  enc := gob.NewEncoder(buf)
  for i := range fids {
    err := enc.Encode(fids[i].facing)
    if err != nil {
      return "", err
    }
    err = enc.Encode(fids[i].node)
    if err != nil {
      return "", err
    }
  }
  h := fnv.New64()
  h.Write(buf.Bytes())
  return fmt.Sprintf("%x.gob", h.Sum64()), nil
}

func makeSheet(path string, anim *yed.Graph, fids []frameId) (*sheet, error) {
  uname, err := uniqueName(fids)
  if err != nil {
    return nil, err
  }

  s := sheet{ path: path, anim: anim, name: uname }
  s.rects = make(map[frameId]FrameRect)
  cy := 0
  cx := 0
  cdy := 0
  tdx := 0
  max_width := 2048
  for _,fid := range fids {
    name := anim.Node(fid.node).Line(0) + ".png"
    file,err := os.Open(filepath.Join(path, fmt.Sprintf("%d", fid.facing), name))
    // if a file isn't there that's ok
    if err != nil { continue }

    config,_,err := image.DecodeConfig(file)
    file.Close()
    // if a file can't be read that is *not* ok
    if err != nil { return nil, err }

    if cx + config.Width > max_width {
      cx = 0
      cy += cdy
      cdy = 0
    }
    if config.Height > cdy {
      cdy = config.Height
    }
    s.rects[fid] = FrameRect{ X: cx, X2: cx + config.Width, Y: cy, Y2: cy + config.Height }
    cx += config.Width
    if cx > tdx {
      tdx = cx
    }
  }
  s.dx = int(nextPowerOf2(uint32(tdx)))
  s.dy = int(nextPowerOf2(uint32(cy + cdy)))
  s.load_chan = make(chan bool)
  s.reference_chan = make(chan int)
  go s.routine()

  return &s, nil
}
