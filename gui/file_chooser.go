package gui

import (
  "path/filepath"
  "os"
)

type FileWidget struct {
  *Button
  path string
}

// If path represents a directory, returns path
// If path represents a file, returns the directory containing path
// The path is always cleaned before it is returned
// If there is an error stating path, "/" is returned
func pathToDir(path string) string {
  info,err := os.Stat(path)
  if err != nil {
    return "/"
  }
  if info.IsDir() {
    return filepath.Clean(path)
  }
  return filepath.Clean(filepath.Join(path, ".."))
}

func MakeFileWidget(path string, gui *Gui) *FileWidget {
  var fw FileWidget
  fw.path = path
  fw.Button = MakeButton("standard", pathToDir(fw.path), 250, 1, 1, 1, 1, func(int64) {
    anchor := MakeAnchorBox(gui.root.Render_region.Dims)
    choose := MakeFileChooser(pathToDir(fw.path), func(f string, err error) {
      defer gui.RemoveChild(anchor)
      if err != nil { return }
      fw.path = f
      fw.Button.SetText(filepath.Base(fw.path))
    })
    anchor.AddChild(choose, Anchor{ 0.5, 0.5, 0.5, 0.5 })
    gui.AddChild(anchor)
  })
  return &fw
}

type FileChooser struct {
  *VerticalTable
  filename    *TextLine
  up_button   *Button
  list_scroll *ScrollFrame
  list        *SelectBox
  choose      *Button
  callback    func(string, error)
}

func (fc *FileChooser) setList() {
  f,err := os.Open(fc.filename.GetText())
  if err != nil {
    fc.callback("", err)
    return
  }
  defer f.Close()
  names,err := f.Readdirnames(0)
  if err != nil {
    fc.callback("", err)
    return
  }
  nlist := MakeSelectTextBox(names, 300)
  fc.list_scroll.ReplaceChild(fc.list, nlist)
  fc.list = nlist
}

func (fc *FileChooser) up() {
  path := fc.filename.GetText()
  dir,file := filepath.Split(path)
  if file == "" {
    dir,file = filepath.Split(path[0 : len(path) - 1])
  }
  fc.filename.SetText(dir)
  fc.setList()
}

func MakeFileChooser(dir string, callback func(string, error)) *FileChooser {
  var fc FileChooser
  fc.callback = callback
  fc.filename = MakeTextLine("standard", dir, 300, 1, 1, 1, 1)
  fc.up_button = MakeButton("standard", "Go up a directory", 200, 1, 1, 1, 1, func(int64) { 
    fc.up()
  })
  fc.list = nil
  fc.choose = MakeButton("standard", "Choose", 200, 1, 1, 1, 1, func(int64) {
    next := filepath.Join(fc.filename.GetText(), fc.list.GetSelectedOption().(string))
    f,err := os.Stat(next)
    if err != nil {
      callback("", err)
      return
    }
    if f.IsDir() {
      fc.filename.SetText(next)
      fc.setList()
    } else {
      callback(next, nil)
    }
  })
  fc.list_scroll = MakeScrollFrame(fc.list, 300, 300)
  fc.VerticalTable = MakeVerticalTable()
  fc.VerticalTable.AddChild(fc.filename)
  fc.VerticalTable.AddChild(fc.up_button)
  fc.VerticalTable.AddChild(fc.list_scroll)
  fc.VerticalTable.AddChild(fc.choose)

  fc.setList()
  return &fc
}

func (w *FileChooser) String() string {
  return "file chooser"
}
