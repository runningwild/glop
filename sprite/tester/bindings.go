package main

import (
  "glop/gin"
  "strings"
  "fmt"
  "os"
  "encoding/json"
  "io/ioutil"
)

// Opens the file named by path, reads it all, decodes it as json into target,
// then closes the file.  Returns the first error found while doing this or nil.
func LoadJson(path string, target interface{}) error {
  f, err := os.Open(path)
  if err != nil {
    return err
  }
  defer f.Close()
  data, err := ioutil.ReadAll(f)
  if err != nil {
    return err
  }
  err = json.Unmarshal(data, target)
  return err
}

func SaveJson(path string, source interface{}) error {
  data, err := json.Marshal(source)
  if err != nil {
    return err
  }
  f, err := os.Create(path)
  if err != nil {
    return err
  }
  defer f.Close()
  _,err = f.Write(data)
  return err
}

type KeyBinds map[string]string
type KeyMap map[string]gin.Key

func getKeysFromString(str string) []gin.KeyId {
  parts := strings.Split(str, "+")
  var kids []gin.KeyId
  for _,part := range parts {
    var kid gin.KeyId
    switch {
    case len(part) == 1:  // Single character - should be ascii
      kid = gin.KeyId(part[0])

    case part == "ctrl":
      kid = gin.EitherControl

    case part == "shift":
      kid = gin.EitherShift

    case part == "alt":
      kid = gin.EitherAlt

    case part == "gui":
      kid = gin.EitherGui

    default:
      key := gin.In().GetKeyByName(part)
      if key == nil {
        panic(fmt.Sprintf("Unknown key '%s'", part))
      }
      kid = key.Id()
    }
    kids = append(kids, kid)
  }
  return kids
}

func (kb KeyBinds) MakeKeyMap() KeyMap {
  key_map := make(KeyMap)
  for key,val := range kb {
    kids := getKeysFromString(val)

    if len(kids) == 1 {
      key_map[key] = gin.In().GetKey(kids[0])
    } else {
      // The last kid is the main kid and the rest are modifiers
      main := kids[len(kids) - 1]
      kids = kids[0 : len(kids) - 1]
      var down []bool
      for _ = range kids {
        down = append(down, true)
      }
      key_map[key] = gin.In().BindDerivedKey(key, gin.In().MakeBinding(main, kids, down))
    }
  }
  return key_map
}
