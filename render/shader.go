package render

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"unsafe"
)

var shader_progs map[string]uint32

func init() {
	shader_progs = make(map[string]uint32)
}

func EnableShader(name string) error {
	if name == "" {
		gl.UseProgram(0)
		return nil
	}
	prog_obj, ok := shader_progs[name]
	if !ok {
		return fmt.Errorf("Tried to use unknown shader '%s'", name)
	}
	gl.UseProgram(prog_obj)
	return nil
}

func SetUniformI(shader, variable string, n int32) error {
	prog, ok := shader_progs[shader]
	if !ok {
		return fmt.Errorf("Tried to set a uniform in an unknown shader '%s'", shader)
	}
	bvariable := []byte(fmt.Sprintf("%s\x00", variable))
	loc := gl.GetUniformLocation(prog, (*uint8)(unsafe.Pointer(&bvariable[0])))
	gl.Uniform1i(loc, n)
	return nil
}

func SetUniformF(shader, variable string, f float32) error {
	prog, ok := shader_progs[shader]
	if !ok {
		return fmt.Errorf("Tried to set a uniform in an unknown shader '%s'", shader)
	}
	bvariable := []byte(fmt.Sprintf("%s\x00", variable))
	loc := gl.GetUniformLocation(prog, (*uint8)(unsafe.Pointer(&bvariable[0])))
	gl.Uniform1f(loc, f)
	return nil
}

func SetUniform4F(shader, variable string, vs []float32) error {
	prog, ok := shader_progs[shader]
	if !ok {
		return fmt.Errorf("Tried to set a uniform in an unknown shader '%s'", shader)
	}
	bvariable := []byte(fmt.Sprintf("%s\x00", variable))
	loc := gl.GetUniformLocation(prog, (*uint8)(unsafe.Pointer(&bvariable[0])))
	gl.Uniform4f(loc, vs[0], vs[1], vs[2], vs[3])
	return nil
}

func RegisterShader(name string, vertex, fragment []byte) error {
	if _, ok := shader_progs[name]; ok {
		return fmt.Errorf("Tried to register a shader called '%s' twice", name)
	}

	vertex_id := gl.CreateShader(gl.VERTEX_SHADER)
	pointer := &vertex[0]
	length := int32(len(vertex))
	gl.ShaderSource(vertex_id, 1, (**uint8)(unsafe.Pointer(&pointer)), &length)
	gl.CompileShader(vertex_id)
	var param int32
	gl.GetShaderiv(vertex_id, gl.COMPILE_STATUS, &param)
	if param == 0 {
		buf := make([]byte, 5*1024)
		var length int32
		gl.GetShaderInfoLog(vertex_id, int32(len(buf)), &length, (*uint8)(unsafe.Pointer(&buf[0])))
		if length > 0 {
			length--
		}
		maxVersion := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
		return fmt.Errorf("Failed to compile vertex shader (max version supported: %q) %q: %q", maxVersion, name, buf[0:int(length)])
	}

	fragment_id := gl.CreateShader(gl.FRAGMENT_SHADER)
	pointer = &fragment[0]
	length = int32(len(fragment))
	gl.ShaderSource(fragment_id, 1, (**uint8)(unsafe.Pointer(&pointer)), &length)
	gl.CompileShader(fragment_id)
	gl.GetShaderiv(fragment_id, gl.COMPILE_STATUS, &param)
	if param == 0 {
		buf := make([]byte, 5*1024)
		var length int32
		gl.GetShaderInfoLog(fragment_id, int32(len(buf)), &length, (*uint8)(unsafe.Pointer(&buf[0])))
		if length > 0 {
			length--
		}
		maxVersion := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
		return fmt.Errorf("Failed to compile fragment shader (max version supported: %q) %q: %q", maxVersion, name, buf[0:int(length)])
	}

	// shader successfully compiled - now link
	program_id := gl.CreateProgram()
	gl.AttachShader(program_id, vertex_id)
	gl.AttachShader(program_id, fragment_id)
	gl.LinkProgram(program_id)
	gl.GetProgramiv(program_id, gl.LINK_STATUS, &param)
	if param == 0 {
		return fmt.Errorf("Failed to link shader '%s': %v", name, param)
	}

	shader_progs[name] = program_id
	return nil
}

func GetAttribLocation(shaderName, attribName string) (int32, error) {
	prog, ok := shader_progs[shaderName]
	if !ok {
		return -1, fmt.Errorf("No shader named '%s'", shaderName)
	}
	return gl.GetAttribLocation(prog, gl.Str(fmt.Sprintf("%s\x00", attribName))), nil
}

func GetUniformLocation(shaderName, uniformName string) (int32, error) {
	prog, ok := shader_progs[shaderName]
	if !ok {
		return -1, fmt.Errorf("No shader named '%s'", shaderName)
	}
	return gl.GetUniformLocation(prog, gl.Str(fmt.Sprintf("%s\x00", uniformName))), nil
}
