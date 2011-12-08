package gui

import (
  "path/filepath"
  "os"
)

type FileChooser struct {
  *VerticalTable
  filename    *TextLine
  up_button   *Button
  list_scroll *ScrollFrame
  list        *SelectBox
  choose      *Button
  callback    func(Widget, string, error)
}

func (fc *FileChooser) setList() {
  f,err := os.Open(fc.filename.GetText())
  if err != nil {
    fc.callback(nil, "", err)
    return
  }
  defer f.Close()
  names,err := f.Readdirnames(0)
  if err != nil {
    fc.callback(nil, "", err)
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

func MakeFileChooser(dir string, callback func(Widget, string, error)) *FileChooser {
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
      callback(nil, "", err)
      return
    }
    if f.IsDir() {
      fc.filename.SetText(next)
      fc.setList()
    } else {
      callback(&fc, next, nil)
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
