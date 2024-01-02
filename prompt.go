package prompt

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/araddon/dateparse"
)

var selectMaxLines = 25    // maximum number of lines to show
var selectScrollOffset = 5 // minimum number of lines above/below cursor
var optionSelected = fmt.Sprintf("%v[\u00D7] %%v%v", escBold, escReset)
var optionUnselected = "[ ] %v"
var keyInterrupt = fmt.Errorf("interrupt")
var keyEscape = fmt.Errorf("escape")

// Enter is a prompt that requires the Enter key to continue.
func Enter(label string) {
	fmt.Printf("%v [enter]: ", label)

	var res string
	fmt.Scanln(&res)
}

// YesNo is a prompt that requires a yes or no answer. It returns true for any of (1,y,yes,t,true), and false for any of (0,n,no,f,false). It is case-insensitive.
func YesNo(label string, deflt bool) bool {
	first := true

Prompt:
	if deflt {
		fmt.Printf("%v [Y/n]: ", label)
	} else {
		fmt.Printf("%v [y/N]: ", label)
	}
	fmt.Printf(escSavePos)

	var res string
	fmt.Scanln(&res)
	res = strings.TrimSpace(res)

	if res == "" {
		fmt.Printf(escMoveUp + escMoveStart + escClearLine)
		if deflt {
			fmt.Printf("%v [Y/n]: yes\n", label)
		} else {
			fmt.Printf("%v [y/N]: no\n", label)
		}
		return deflt
	} else {
		res = strings.ToLower(res)
	}

	var b bool
	var err error
	if res == "y" || res == "yes" {
		b = true
	} else if res == "n" || res == "no" {
		b = false
	}
	if err != nil {
		first = false
		fmt.Printf("%v%v%vERROR: %v%v%v", escClearLine, escRed, escBold, err, escReset, escMoveUp)
		fmt.Printf(escMoveStart + escClearLine)
		goto Prompt
	} else if !first {
		fmt.Printf(escClearLine) // clear error
	}
	return b
}

type defaultValue struct {
	idst   interface{}
	ideflt interface{}
	pos    int
}

// Default is the default value with the initial text caret position used for Prompt.
func Default(idst, ideflt interface{}, pos int) defaultValue {
	return defaultValue{idst, ideflt, pos}
}

// Prompt is a regular text prompt that can read into a (string,[]byte,bool,int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64,float32,float64,time.Time) or a type that implements the Scanner interface. The idst must be a pointer to a variable, its value determines the default/initial value.
// The initial value will be editable in-place. To set the text caret initial position when idst is editable, use prompt.Default(value, position). When editing, you can use the Left or Ctrl+B, Right or Ctrl+F, Home or Ctrl+A, End or Ctrl+E to move around; Backspace and Delete to delete a character; Ctrl+U and Ctrl+K to delete from the caret to the beginning and the end of the line respectively; Ctrl+C and Escape to quit; and Ctrl+Z and Enter to confirm the input.
// All validators must be satisfies, otherwise an error is printed and the answer should be corrected.
func Prompt(idst interface{}, label string, validators ...Validator) error {
	first := true

	pos := -1
	hasDeflt := false
	var ideflt interface{}
	if deflt, ok := idst.(defaultValue); ok {
		idst = deflt.idst
		ideflt = deflt.ideflt
		pos = deflt.pos
		hasDeflt = true
	}

	// get destination
	dst := reflect.ValueOf(idst)
	if dst.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be a pointer to a variable")
	}
	idst = dst.Elem().Interface()
	if !hasDeflt && ideflt == nil && !dst.Elem().IsZero() {
		ideflt = idst
	}

	editDefault := false
	switch idst.(type) {
	case nil:
		// ignore
	case []byte, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, time.Time:
		editDefault = true
	default:
		if _, ok := idst.(interface {
			String() string
		}); ok {
			editDefault = true
			if ideflt == nil {
				ideflt = idst
			}
		}
	}

	var result []rune
	if editDefault {
		switch deflt := ideflt.(type) {
		case nil:
			// no-op
		case []byte:
			result = []rune(string(deflt))
		case string:
			result = []rune(deflt)
		default:
			result = []rune(fmt.Sprint(ideflt))
		}
	}
	if pos == -1 {
		pos = len(result)
	} else if pos < 0 {
		pos = 0
	} else if len(result) < pos {
		pos = len(result)
	}

Prompt:
	// prompt input
	if _, ok := idst.(bool); ok {
		if deflt, ok := ideflt.(bool); ok {
			if deflt {
				fmt.Printf("%v [Y/n]: ", label)
			} else {
				fmt.Printf("%v [y/N]: ", label)
			}
		} else {
			fmt.Printf("%v [y/n]: ", label)
		}
		result = []rune{}
		pos = 0
	} else {
		fmt.Printf("%v: %v", label, string(result))
		fmt.Printf(strings.Repeat(escMoveLeft, len(result)-pos))
	}

	// make raw and hide input
	restore, err := MakeRawTerminal(false)
	if err != nil {
		return err
	}

	func() {
		defer restore()

		// read input
		input := bufio.NewReader(os.Stdin)
		for {
			var r rune
			if r, _, err = input.ReadRune(); err != nil {
				break
			}

			if r == '\x03' { // interrupt
				err = keyInterrupt
				break
			} else if r == '\x04' || r == '\r' || r == '\n' { // select
				break
			} else if r == '\x7F' { // backspace
				if pos != 0 {
					result = append(result[:pos-1], result[pos:]...)
					pos--
					fmt.Printf(escMoveLeft+"%v "+strings.Repeat(escMoveLeft, len(result)+1-pos), string(result[pos:]))
				}
			} else if r == '\x1B' { // escape
				if input.Buffered() == 0 {
					err = keyEscape
					break
				} else if r, _, err = input.ReadRune(); err != nil {
					break
				} else if r == '[' { // CSI
					if input.Buffered() == 0 {
						// ignore
					} else if r, _, err = input.ReadRune(); err != nil {
						break
					} else if r == 'D' { // left
						if pos != 0 {
							fmt.Printf(escMoveLeft)
							pos--
						}
					} else if r == 'C' { // right
						if pos != len(result) {
							fmt.Printf(escMoveRight)
							pos++
						}
					} else if r == 'H' { // home
						fmt.Printf(strings.Repeat(escMoveLeft, pos))
						pos = 0
					} else if r == 'F' { // end
						fmt.Printf(strings.Repeat(escMoveRight, len(result)-pos))
						pos = len(result)
					} else if r == '3' {
						if input.Buffered() == 0 {
							// ignore
						} else if r, _, err = input.ReadRune(); err != nil {
							break
						} else if r == '~' { // delete
							if pos != len(result) {

								result = append(result[:pos], result[pos+1:]...)
								fmt.Printf("%v "+strings.Repeat(escMoveLeft, len(result)+1-pos), string(result[pos:]))
							}
						}
					}
				}
			} else if r == '\x01' { // Ctrl+A - move to start of line
				fmt.Printf(strings.Repeat(escMoveLeft, pos))
				pos = 0
			} else if r == '\x02' { // Ctrl+B - move back
				fmt.Printf(escMoveLeft)
				pos--
			} else if r == '\x05' { // Ctrl+E - move to end of line
				fmt.Printf(strings.Repeat(escMoveRight, len(result)-pos))
				pos = len(result)
			} else if r == '\x06' { // Ctrl+F - move forward
				fmt.Printf(escMoveRight)
				pos++
			} else if r == '\x0B' { // Ctrl+K - delete to end of line
				fmt.Printf(strings.Repeat(" ", len(result)-pos))
				fmt.Printf(strings.Repeat(escMoveLeft, len(result)-pos))
				result = result[:pos]
			} else if r == '\x15' { // Ctrl+U - delete to start of line
				fmt.Printf(strings.Repeat(escMoveLeft, pos))
				fmt.Printf("%v"+strings.Repeat(" ", pos), string(result[pos:]))
				fmt.Printf(strings.Repeat(escMoveLeft, len(result)))
				result = result[pos:]
				pos = 0
			} else if ' ' <= r {
				result = append(result[:pos], append([]rune{r}, result[pos:]...)...)
				fmt.Printf("%v"+strings.Repeat(escMoveLeft, len(result)-pos-1), string(result[pos:]))
				pos++
			}
		}
	}()

	if err != nil {
		if !first {
			fmt.Printf(escMoveDown + escClearLine + escMoveUp)
		}
		if err == keyInterrupt {
			fmt.Printf(strings.Repeat(escMoveRight, len(result)-pos) + "^C")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
		fmt.Printf("\n")
		return err
	}

	fmt.Println(escMoveStart)

	// fill destination
	res := strings.TrimSpace(string(result))
	ival := ideflt
	if editDefault || res != "" || ival == nil {
		switch idst.(type) {
		case []byte:
			ival = []byte(res)
		case string:
			ival = res
		case bool:
			var b bool
			if res == "y" || res == "Y" || res == "yes" || res == "YES" {
				b = true
			} else if res == "n" || res == "N" || res == "no" || res == "NO" {
				b = false
			} else {
				var perr error
				b, perr = strconv.ParseBool(res)
				if perr != nil {
					err = fmt.Errorf("invalid boolean")
				}
			}
			ival = b
		case int:
			i, perr := strconv.ParseInt(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid integer")
			} else if math.MaxInt < i {
				err = fmt.Errorf("integer overflow")
			}
			ival = int(i)
		case int8:
			i, perr := strconv.ParseInt(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid integer")
			} else if math.MaxInt8 < i {
				err = fmt.Errorf("integer overflow")
			}
			ival = int8(i)
		case int16:
			i, perr := strconv.ParseInt(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid integer")
			} else if math.MaxInt16 < i {
				err = fmt.Errorf("integer overflow")
			}
			ival = int16(i)
		case int32:
			i, perr := strconv.ParseInt(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid integer")
			} else if math.MaxInt64 < i {
				err = fmt.Errorf("integer overflow")
			}
			ival = int32(i)
		case int64:
			i, perr := strconv.ParseInt(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid integer")
			}
			ival = i
		case uint:
			u, perr := strconv.ParseUint(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid positive integer")
			} else if math.MaxInt < u {
				err = fmt.Errorf("integer overflow")
			}
			ival = uint(u)
		case uint8:
			u, perr := strconv.ParseUint(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid positive integer")
			} else if math.MaxInt8 < u {
				err = fmt.Errorf("integer overflow")
			}
			ival = uint8(u)
		case uint16:
			u, perr := strconv.ParseUint(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid positive integer")
			} else if math.MaxInt16 < u {
				err = fmt.Errorf("integer overflow")
			}
			ival = uint16(u)
		case uint32:
			u, perr := strconv.ParseUint(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid positive integer")
			} else if math.MaxInt64 < u {
				err = fmt.Errorf("integer overflow")
			}
			ival = uint32(u)
		case uint64:
			u, perr := strconv.ParseUint(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid positive integer")
			}
			ival = u
		case float32:
			f, perr := strconv.ParseFloat(res, 32)
			if perr.(*strconv.NumError).Err == strconv.ErrRange {
				err = fmt.Errorf("floating point overflow")
			} else if perr != nil {
				err = fmt.Errorf("invalid floating point")
			}
			ival = float32(f)
		case float64:
			f, perr := strconv.ParseFloat(res, 64)
			if perr.(*strconv.NumError).Err == strconv.ErrRange {
				err = fmt.Errorf("floating point overflow")
			} else if perr != nil {
				err = fmt.Errorf("invalid floating point")
			}
			ival = f
		case time.Time:
			t, perr := dateparse.ParseAny(res)
			if perr != nil {
				err = fmt.Errorf("invalid datetime")
			}
			ival = t
		default:
			if scanner, ok := dst.Interface().(interface {
				Scan(interface{}) error
			}); ok {
				// already sets value to dst
				if perr := scanner.Scan(res); perr != nil {
					err = fmt.Errorf("invalid %T: %w", idst, perr)
				}
				ival = dst.Elem().Interface()
			} else {
				return fmt.Errorf("unsupported destination type: %T", idst)
			}
		}
	} else if deflt, ok := ideflt.(bool); ok {
		fmt.Printf(escMoveUp + escMoveStart + escClearLine)
		if deflt {
			fmt.Printf("%v [Y/n]: yes\n", label)
		} else {
			fmt.Printf("%v [y/N]: no\n", label)
		}
	}

	// validators
	if err == nil {
		for _, validator := range validators {
			if verr := validator(ival); verr != nil {
				err = verr
				break
			}
		}
	}

	if err != nil {
		first = false
		fmt.Printf("%v%v%vERROR: %v%v%v", escClearLine, escRed, escBold, err, escReset, escMoveUp)
		fmt.Printf(escMoveStart + escClearLine)
		goto Prompt
	} else if !first {
		fmt.Printf(escClearLine)
	}
	dst.Elem().Set(reflect.ValueOf(ival))
	return nil
}
