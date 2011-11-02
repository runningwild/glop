package game

import(
  "fmt"
  "rand"
  "strings"
  "regexp"
  "strconv"
)

var dice_exp *regexp.Regexp

func init() {
  base := "\\d+d\\d+"
  dice_exp = regexp.MustCompile(fmt.Sprintf("%s(\\s*\\+\\s*%s)?", base))
}

func doRoll(roll string) int {
  vals := strings.Split(roll, "d")
  n,_ := strconv.Atoi(vals[0])
  d,_ := strconv.Atoi(vals[1])

  total := 0
  for i := 0; i < n; i++ {
    total += rand.Intn(d) + 1
  }
  return total
}

func Dice(d string) int {
  if !dice_exp.MatchString(d) {
    panic(fmt.Sprintf("'%s' is not a valid dice string.", d))
  }
  rolls := strings.Split(d, "+")
  total := 0
  for _,roll := range rolls {
    total += doRoll(strings.Trim(roll, " \t\r\f\n"))
  }
  return total
}
