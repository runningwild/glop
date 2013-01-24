package render

import (
  "fmt"
  gl "github.com/chsc/gogl/gl21"
  "unsafe"
)

var shader_progs map[string]gl.Uint

func init() {
  shader_progs = make(map[string]gl.Uint)
}

type shaderError string

func (err shaderError) Error() string {
  return string(err)
}

func EnableShader(name string) error {
  if name == "" {
    gl.UseProgram(0)
    return nil
  }
  prog_obj, ok := shader_progs[name]
  if !ok {
    return shaderError(fmt.Sprintf("Tried to use unknown shader '%s'", name))
  }
  gl.UseProgram(prog_obj)
  return nil
}

func SetUniformI(shader, variable string, n int) error {
  prog, ok := shader_progs[shader]
  if !ok {
    return shaderError(fmt.Sprintf("Tried to set a uniform in an unknown shader '%s'", shader))
  }
  bvariable := []byte(fmt.Sprintf("%s\x00", variable))
  loc := gl.GetUniformLocation(prog, (*gl.Char)(unsafe.Pointer(&bvariable[0])))
  gl.Uniform1i(loc, gl.Int(n))
  return nil
}

func SetUniformF(shader, variable string, f float32) error {
  prog, ok := shader_progs[shader]
  if !ok {
    return shaderError(fmt.Sprintf("Tried to set a uniform in an unknown shader '%s'", shader))
  }
  bvariable := []byte(fmt.Sprintf("%s\x00", variable))
  loc := gl.GetUniformLocation(prog, (*gl.Char)(unsafe.Pointer(&bvariable[0])))
  gl.Uniform1f(loc, gl.Float(f))
  return nil
}

func RegisterShader(name string, vertex, fragment []byte) error {
  if _, ok := shader_progs[name]; ok {
    return shaderError(fmt.Sprintf("Tried to register a shader called '%s' twice", name))
  }

  vertex_id := gl.CreateShader(gl.VERTEX_SHADER)
  pointer := &vertex[0]
  length := gl.Int(len(vertex))
  gl.ShaderSource(vertex_id, 1, (**gl.Char)(unsafe.Pointer(&pointer)), &length)
  gl.CompileShader(vertex_id)
  var param gl.Int
  gl.GetShaderiv(vertex_id, gl.COMPILE_STATUS, &param)
  if param == 0 {
    return shaderError(fmt.Sprintf("Failed to compile vertex shader '%s': %v", name, param))
  }

  fragment_id := gl.CreateShader(gl.FRAGMENT_SHADER)
  pointer = &fragment[0]
  length = gl.Int(len(fragment))
  gl.ShaderSource(fragment_id, 1, (**gl.Char)(unsafe.Pointer(&pointer)), &length)
  gl.CompileShader(fragment_id)
  gl.GetShaderiv(fragment_id, gl.COMPILE_STATUS, &param)
  if param == 0 {
    return shaderError(fmt.Sprintf("Failed to compile fragment shader '%s': %v", name, param))
  }

  // shader successfully compiled - now link
  program_id := gl.CreateProgram()
  gl.AttachShader(program_id, vertex_id)
  gl.AttachShader(program_id, fragment_id)
  gl.LinkProgram(program_id)
  gl.GetProgramiv(program_id, gl.LINK_STATUS, &param)
  if param == 0 {
    return shaderError(fmt.Sprintf("Failed to link shader '%s': %v", name, param))
  }

  shader_progs[name] = program_id
  return nil
}
