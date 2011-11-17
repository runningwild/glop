package game

import (
  "fmt"
  "reflect"
  "path/filepath"
  "os"
  "strings"
  "io/ioutil"
  "json"
)

type Icon interface {
  IconPath() string
}

type basicIcon string
func (icon basicIcon) IconPath() string {
  return string(icon)
}

// An Action represents anything that a unit can spend AP to do, move, attack,
// charge, anything where the user will have to use some sort of GUI to make
// it explicit how to perform the action, etc...
type Action interface {
  Icon

  // Run when the player selects this action but before he commits to it.
  // Performs any pre-processing desired and highlights anything necessary or
  // creates any GUI necessary for the user to decide how to proceed.  Returns
  // true if this action could be prepped, in which case it is available for 
  // an immediate call to Do() or Maintain().
  Prep() bool

  // Run if Prep() was selected but the user decided not to perform the action.
  // Should tear down any GUI created and undo anything else no long necessary
  // to keep around.  Can be called redundantly.
  Cancel()

  // Called regularly after Prep() but before Do() and gives the action a
  // chance to modify any UI in response to mouse location.
  // bx and by are board coordinates
  MouseOver(bx,by float64)

  // Called after Prep() and indicates that the user clicked at a particular
  // location.  This function should return true if this click means that the
  // user has chosen to commit to this action.
  // bx and by are board coordinates
  MouseClick(bx,by float64) bool

  // Called periodically after Prep() and after the action has been committed.
  // The method should return false until the action is complete, at which
  // point it should return true.  After this method returns true the Mouse*()
  // methods and Maintain() will not be called again until the action has been
  // prepped again.
  Maintain(dt int64) bool
}

type ActionSpec struct {
  Type       string
  Icon_path  string
  Name       string
  Int_params map[string]int
}

var action_type_registry map[string]reflect.Type

func registerActionType(name string, action Action) {
  if action_type_registry == nil {
    action_type_registry = make(map[string]reflect.Type)
  }
  action_type_registry[name] = reflect.TypeOf(action).Elem()
}

func assignParams(action_val reflect.Value, ent *Entity, icon_path string, int_params map[string]int) {
  ent_field := reflect.Indirect(action_val).FieldByName("Ent")
  if ent_field.Kind() == reflect.Invalid {
    panic(fmt.Sprintf("Action %v has no Entity field.", action_val))
  }
  ent_field.Set(reflect.ValueOf(ent))

  icon_field := reflect.Indirect(action_val).FieldByName("basicIcon")
  if icon_field.Kind() == reflect.Invalid {
    panic(fmt.Sprintf("Action %v has no basicIcon field.", action_val))
  }
  icon_field.Set(reflect.ValueOf(basicIcon(icon_path)))

  for k,v := range int_params {
    field := reflect.Indirect(action_val).FieldByName(k)
    if field.Kind() == reflect.Invalid {
      panic(fmt.Sprintf("int param '%s' specified, but corresponding field was not found.", k))
    }
    if field.Kind() != reflect.Int {
      panic(fmt.Sprintf("int param '%s' specified, but field is of type %s.", k, field.Kind()))
    }
    field.SetInt(int64(v))
  }
}

var action_spec_registry map[string]ActionSpec

func registerActionSpec(spec ActionSpec) {
  if action_spec_registry == nil {
    action_spec_registry = make(map[string]ActionSpec)
  }
  if _,ok := action_spec_registry[spec.Name]; ok {
    panic(fmt.Sprintf("Tried to register the spec '%s' more than once.", spec.Name))
  }
  if _,ok := action_type_registry[spec.Type]; !ok {
    panic(fmt.Sprintf("Tried to register a spec for the action type '%s', which doesn't exist.", spec.Type))
  }
  action_spec_registry[spec.Name] = spec

  // Test to make sure this thing can really make an action without panicing,
  // this way we fail fast.
  MakeAction(spec.Name, nil)
}

func MakeAction(spec_name string, ent *Entity) Action {
  spec,ok := action_spec_registry[spec_name]
  if !ok {
    panic(fmt.Sprintf("Tried to load an unknown ActionSpec '%s'.", spec_name))
  }
  action := reflect.New(action_type_registry[spec.Type])
  assignParams(action, ent, spec.Icon_path, spec.Int_params)
  return action.Interface().(Action)
}

// Finds all *.json files in dir and registers all
func RegisterAllSpecsInDir(dir string) {
  err := filepath.Walk(dir, func(path string, info *os.FileInfo, err error) error {
    if info.IsDirectory() {
      return nil
    }
    if !strings.HasSuffix(path, ".json") {
      return nil
    }
    f,err := os.Open(path)
    if err != nil { return err }
    data,err := ioutil.ReadAll(f)
    f.Close()
    if err != nil { return err }
    var specs []ActionSpec
    err = json.Unmarshal(data, &specs)
    if err != nil { return err }
    for _,spec := range specs {
      registerActionSpec(spec)
    }
    return nil
  })
  if err != nil {
    panic(fmt.Sprintf("Unable to load all specs in directory '%s': %s\n", dir, err.Error()))
  }
}
