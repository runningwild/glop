package game

import (
  "game/stats"
  "fmt"
  "reflect"
  "path/filepath"
  "os"
  "strings"
  "io/ioutil"
  "encoding/json"
)

type Icon interface {
  IconPath() string
}

type basicIcon string
func (icon basicIcon) IconPath() string {
  return string(icon)
}

type nonInterrupt struct {}
func (nonInterrupt) Interrupt() bool { return false }

type uninterruptable struct {}
func (uninterruptable) Pause() bool { panic("This action should never be paused.") }

type ActionCommit int
const (
  NoAction ActionCommit = iota
  StandardAction
  StandardInterrupt
)

type MaintenanceStatus int
const (
  InProgress MaintenanceStatus = iota
  CheckForInterrupts
  Complete
)

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

  // Actions will be paused if interrupted.  This method will only be called
  // immediately after a call to Maintain() if the return value of the call to
  // Maintain() is CheckForInterrupts.  Actions may choose to completely cancel
  // themselves when they are paused, if they do so they should return false
  // from this method, otherwise they should return true.
  Pause() bool

  // Called regularly after Prep() but before Do() and gives the action a
  // chance to modify any UI in response to mouse location.
  // bx and by are board coordinates
  MouseOver(bx,by float64)

  // Called after Prep() and indicates that the user clicked at a particular
  // location.  This return value indicates one of the following:
  // NoAction - The user has not committed to this action
  // StandardAction - The user has committed to this action, begin resolving it.
  // StandardInterrupt - The user has committed to this action, keep it resident
  // and poll it as an interrupt when appropriate to see if it activates.
  MouseClick(bx,by float64) ActionCommit

  // Called periodically after Prep() and after the action has been committed.
  // The method should return InProgress until the action is complete, at which
  // point it should return Complete.  After this method returns true the
  // Mouse*() methods and Maintain() will not be called again until the action
  // has been prepped again.
  Maintain(dt int64) MaintenanceStatus

  // Actions that act as interrupts should return true when this method is
  // called if they want to take effect.  Once an interrupt returns true from
  // this method it will take over as the active action and its Maintain method
  // will be called regularly until it returns true.
  Interrupt() bool
}

type ActionSpec struct {
  Type       string
  Icon_path  string
  Name       string
  Effects    []string
  Int_params map[string]int
}

var action_type_registry map[string]reflect.Type

func registerActionType(name string, action Action) {
  if action_type_registry == nil {
    action_type_registry = make(map[string]reflect.Type)
  }
  action_type_registry[name] = reflect.TypeOf(action).Elem()
}

func assignParams(spec_name string, action_val reflect.Value, ent *Entity, icon_path string, effects []string, int_params map[string]int) {
  ent_field := reflect.Indirect(action_val).FieldByName("Ent")
  if ent_field.Kind() == reflect.Invalid {
    panic(fmt.Sprintf("Action '%s' has no Entity field.", spec_name))
  }
  ent_field.Set(reflect.ValueOf(ent))

  icon_field := reflect.Indirect(action_val).FieldByName("basicIcon")
  if icon_field.Kind() == reflect.Invalid {
    panic(fmt.Sprintf("Action '%s' has no basicIcon field.", spec_name))
  }
  icon_field.Set(reflect.ValueOf(basicIcon(icon_path)))

  effects_field := reflect.Indirect(action_val).FieldByName("Effects")
  if effects_field.IsValid() {
    if effects == nil {
      panic(fmt.Sprintf("Action '%s' expected an Effects field, but it was not supplied.", spec_name))
    }
    func () {
      defer func() {
        if r := recover(); r != nil {
          panic(fmt.Sprintf("Action '%s' specified an unknown effect: %v\n", spec_name, r))
        }
      } ()
      // Just check that all of the effects are valid effects
      for _,effect := range effects {
        stats.MakeEffect(effect)
      }
    } ()
    effects_field.Set(reflect.ValueOf(effects))
  } else {
    if effects != nil {
      panic(fmt.Sprintf("Action '%s' did not expect an Effects field, but it was supplied.", spec_name))
    }
  }

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
  assignParams(spec_name, action, ent, spec.Icon_path, spec.Effects, spec.Int_params)
  return action.Interface().(Action)
}

// Finds all *.json files in dir and registers all
func RegisterAllSpecsInDir(dir string) {
  err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
    if info.IsDir() {
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
