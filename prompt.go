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

var selectMaxLines = 10    // maximum number of lines to show
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
		fmt.Printf(escRestorePos + escMoveUp)
		if deflt {
			fmt.Printf("yes\n")
		} else {
			fmt.Printf("no\n")
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
	} else {
		var perr error
		b, perr = strconv.ParseBool(res)
		if perr != nil {
			err = fmt.Errorf("invalid boolean")
		}
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
	val interface{}
	pos int
}

// Default is the default value with the initial text caret position used for Prompt.
func Default(val interface{}, pos int) defaultValue {
	return defaultValue{val, pos}
}

func (def defaultValue) String() string {
	return fmt.Sprint(def.val)
}

func (def defaultValue) Pos() int {
	return def.pos
}

// Prompt is a regular text prompt that can read into a (string,[]byte,bool,int,int8,int16,int32,int64,uint,uint8,uint16,uint32,uint64,float32,float64,time.Time) or a type that implements the Scanner interface. The idst must be a pointer to a variable, and ideflt must be of the same type as the variable.
// If ideflt is any of the basic types mentioned (except bool) or implements the Stringer interface, its value will be editable in-place. To set the text caret initial position when ideflt is editable, use prompt.Default(value, position). When editing, you can use the Left or Ctrl+B, Right or Ctrl+F, Home or Ctrl+A, End or Ctrl+E to move around; Backspace and Delete to delete a character; Ctrl+U and Ctrl+K to delete from the caret to the beginning and the end of the line respectively; Ctrl+C and Escape to quit; and Ctrl+Z and Enter to confirm the input.
// All validators must be satisfies, otherwise an error is printed and the answer should be corrected.
func Prompt(idst interface{}, label string, ideflt interface{}, validators ...Validator) error {
	first := true

	// get destination
	dst := reflect.ValueOf(idst)
	if dst.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be pointer to variable")
	}
	idst = dst.Elem().Interface()

	editDefault := false
	switch ideflt.(type) {
	case nil:
		// ignore
	case []byte, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, time.Time:
		editDefault = true
	default:
		if _, ok := ideflt.(interface {
			String() string
		}); ok {
			editDefault = true
		}
	}

	pos := 0
	var result []rune
	if editDefault {
		switch deflt := ideflt.(type) {
		case []byte:
			result = []rune(string(deflt))
		case string:
			result = []rune(deflt)
		default:
			result = []rune(fmt.Sprint(ideflt))
		}

		if poser, ok := ideflt.(interface {
			Pos() int
		}); ok {
			pos = poser.Pos()
		} else {
			pos = len(result)
		}
		if pos < 0 {
			pos = 0
		} else if len(result) < pos {
			pos = len(result)
		}
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
					fmt.Printf(escMoveLeft + string(result[pos:]) + " " + strings.Repeat(escMoveLeft, len(result)+1-pos))
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
								fmt.Printf(string(result[pos:]) + " " + strings.Repeat(escMoveLeft, len(result)+1-pos))
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
				fmt.Printf(string(result[pos:]) + strings.Repeat(" ", pos))
				fmt.Printf(strings.Repeat(escMoveLeft, len(result)))
				result = result[pos:]
				pos = 0
			} else if ' ' <= r {
				result = append(result[:pos], append([]rune{r}, result[pos:]...)...)
				fmt.Printf(string(result[pos:]) + strings.Repeat(escMoveLeft, len(result)-pos-1))
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
		} else if err == keyEscape {
			fmt.Printf("\n")
			return nil
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

func getSelected(iselected interface{}, options reflect.Value) (int, error) {
	var selected int
	if iselected == nil {
		// noop
	} else if reflect.TypeOf(iselected) == options.Type().Elem() {
		vselected := reflect.ValueOf(iselected)
		for i := 0; i < options.Len(); i++ {
			if options.Index(i).Equal(vselected) {
				selected = i
				break
			}
		}
	} else if i, ok := iselected.(int); ok {
		selected = i
	} else if i, ok := iselected.(int8); ok {
		selected = int(i)
	} else if i, ok := iselected.(int16); ok {
		selected = int(i)
	} else if i, ok := iselected.(int32); ok {
		selected = int(i)
	} else if i, ok := iselected.(int64); ok {
		selected = int(i)
	} else if u, ok := iselected.(uint); ok {
		selected = int(u)
	} else if u, ok := iselected.(uint8); ok {
		selected = int(u)
	} else if u, ok := iselected.(uint16); ok {
		selected = int(u)
	} else if u, ok := iselected.(uint32); ok {
		selected = int(u)
	} else if u, ok := iselected.(uint64); ok {
		selected = int(u)
	} else {
		return 0, fmt.Errorf("selected must be integer type or %v", options.Type().Elem())
	}
	if selected < 0 || options.Len() <= selected {
		return 0, fmt.Errorf("selected must be in range [0,%d]", options.Len()-1)
	}
	return selected, nil
}

func matchOption(query, option string) bool {
	return strings.Contains(strings.ToLower(option), strings.ToLower(query))
}

// Select is a list selection prompt that allows to select one of the list of possible values. The ioptions must be a slice of options. The idst must be a pointer to a variable and must of of the same type as the options (set the option value) or an integer (set the option index). Equally, iselected can be an integer (index) or of the same type as the options (value).
// Users can select an option using Up or W or K to move up, Down or S or J to move down, Tab and Shift+Tab to move down and up respectively and wrap around, Ctrl+C or Escape to quit, and Ctrl+Z or Enter to select an option.
func Select(idst interface{}, label string, ioptions, iselected interface{}) error {
	dst := reflect.ValueOf(idst)
	if dst.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be pointer to variable")
	}
	dst = dst.Elem()

	options := reflect.ValueOf(ioptions)
	if options.Kind() != reflect.Slice {
		return fmt.Errorf("options must be slice")
	} else if options.Len() == 0 {
		return fmt.Errorf("no options")
	}
	selected, err := getSelected(iselected, options)
	if err != nil {
		return err
	}

	// set constants
	maxLines := selectMaxLines
	if _, rows, err := TerminalSize(); err != nil {
		return err
	} else if rows-1 < maxLines {
		maxLines = rows - 1 // keep one for prompt row
	}
	search := maxLines < options.Len() || 25 < options.Len()
	numLines := Min(maxLines, options.Len())
	scrollOffset := selectScrollOffset
	if (numLines-1)/2 < scrollOffset {
		scrollOffset = (numLines - 1) / 2
	}

	// print label
	fmt.Printf("%v:", label)
	padding := ""
	if 2 < len(label) {
		padding = strings.Repeat(" ", len(label)-2)
	}
	preOption := escMoveStart + escClearLine + padding

	// print options
	windowStart := Clip(selected-(numLines-1)/2, 0, options.Len()-numLines)
	for i := 0; i < numLines; i++ {
		opt := options.Index(windowStart + i).Interface()
		if windowStart+i == selected {
			fmt.Printf("\n"+padding+optionSelected, opt)
		} else {
			fmt.Printf("\n"+padding+optionUnselected, opt)
		}
	}
	// go to selected option
	fmt.Printf(escMoveUpN+escMoveToCol, numLines, len(label)+3)

	optionsIndex := make([]int, options.Len())
	optionsString := make([]string, options.Len())
	for i := 0; i < options.Len(); i++ {
		optionsIndex[i] = i
		optionsString[i] = fmt.Sprint(options.Index(i).Interface())
	}

	// make raw and hide input
	restore, err := MakeRawTerminal(!search)
	if err != nil {
		return err
	}

	pos := 0
	var prevQuery, query []rune
	prevSelected := selected
	func() {
		defer restore()

		// read input
		input := bufio.NewReader(os.Stdin)
		for {
			// change search results
			if search && string(query) != string(prevQuery) {
				fmt.Printf(escMoveStart+escClearLine+"%v: %v"+escMoveToCol, label, string(query), len(label)+3+pos)
				i := 0
				hasSelected := false
				optionsIndex = optionsIndex[:0]
				optionsString = optionsString[:0]
				for i < options.Len() {
					v := fmt.Sprint(options.Index(i).Interface())
					if matchOption(string(query), v) {
						if i == selected {
							selected = len(optionsIndex)
							hasSelected = true
						}
						optionsIndex = append(optionsIndex, i)
						optionsString = append(optionsString, v)
					}
					i++
				}
				if !hasSelected {
					selected = 0
				}
				prevQuery = query

				fmt.Printf(escMoveStart + strings.Repeat(escMoveDown+escClearLine, numLines))
				if 0 < numLines {
					fmt.Printf(escMoveUpN, numLines)
				}
				numLines = Min(maxLines, len(optionsIndex))
				if numLines == 0 {
					fmt.Printf("\n" + padding + escRed + "No options found" + escReset)
					fmt.Printf(escMoveUp+escMoveToCol, len(label)+3+pos)
					prevSelected = 0
				} else {
					prevSelected = -1
				}
			}

			// change selection and move window
			if selected != prevSelected {
				prevWindowStart := windowStart
				if prevSelected == -1 {
					windowStart = Clip(selected-(numLines-1)/2, 0, len(optionsIndex)-numLines)
				} else if selected < prevSelected && Max(0, selected-scrollOffset) < windowStart {
					// move window up
					windowStart = Max(0, selected-scrollOffset)
				} else if prevSelected < selected && windowStart < Min(selected+scrollOffset+1-numLines, len(optionsIndex)-numLines) {
					// move window down
					windowStart = Min(selected+scrollOffset+1-numLines, len(optionsIndex)-numLines)
				}
				if windowStart != prevWindowStart || prevSelected == -1 {
					// print all options
					for i := 0; i < numLines; i++ {
						opt := optionsString[windowStart+i]
						if windowStart+i == selected {
							fmt.Printf(escMoveDown+preOption+optionSelected, opt)
						} else {
							fmt.Printf(escMoveDown+preOption+optionUnselected, opt)
						}
					}
					// go to selected option
					fmt.Printf(escMoveUpN+escMoveToCol, numLines, len(label)+3+pos)
				} else {
					fmt.Printf(escMoveDownN+preOption+optionUnselected, prevSelected-windowStart+1, optionsString[prevSelected])
					if selected < prevSelected {
						fmt.Printf(escMoveUpN, prevSelected-selected)
					} else {
						fmt.Printf(escMoveDownN, selected-prevSelected)
					}
					fmt.Printf(preOption+optionSelected, optionsString[selected])
					fmt.Printf(escMoveUpN+escMoveToCol, selected-windowStart+1, len(label)+3+pos)
				}
				prevSelected = selected
			}

			// read user input
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
					query = append(query[:pos-1], query[pos:]...)
					pos--
					fmt.Printf(escMoveLeft + string(query[pos:]) + " " + strings.Repeat(escMoveLeft, len(query)+1-pos))
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
						if pos != len(query) {
							fmt.Printf(escMoveRight)
							pos++
						}
					} else if r == 'H' { // home
						fmt.Printf(strings.Repeat(escMoveLeft, pos))
						pos = 0
					} else if r == 'F' { // end
						fmt.Printf(strings.Repeat(escMoveRight, len(query)-pos))
						pos = len(query)
					} else if r == 'A' || r == '\x5A' { // up or shift+tab
						selected--
						if selected < 0 {
							if len(optionsIndex) == 0 {
								selected = 0
							} else {
								selected = len(optionsIndex) - 1
							}
						}
					} else if r == 'B' { // down
						selected++
						if len(optionsIndex) <= selected {
							selected = 0
						}
					} else if r == '3' || r == '5' || r == '6' {
						if input.Buffered() == 0 {
							// ignore
						} else if tilde, _, err := input.ReadRune(); err != nil {
							break
						} else if tilde == '~' {
							if r == '3' { // delete
								if pos != len(query) {

									query = append(query[:pos], query[pos+1:]...)
									fmt.Printf(string(query[pos:]) + " " + strings.Repeat(escMoveLeft, len(query)+1-pos))
								}
							} else if r == '5' { // page up
								selected -= numLines
								if selected < 0 {
									selected = 0
								}
							} else if r == '6' { // page down
								selected += numLines
								if len(optionsIndex) <= selected {
									if len(optionsIndex) == 0 {
										selected = 0
									} else {
										selected = len(optionsIndex) - 1
									}
								}
							}
						}
					}
				}
			} else if r == '\t' { // tab
				selected++
				if len(optionsIndex) <= selected {
					selected = 0
				}
			} else if r == '\x01' { // Ctrl+A - move to start of line
				fmt.Printf(strings.Repeat(escMoveLeft, pos))
				pos = 0
			} else if r == '\x02' { // Ctrl+B - move back
				fmt.Printf(escMoveLeft)
				pos--
			} else if r == '\x05' { // Ctrl+E - move to end of line
				fmt.Printf(strings.Repeat(escMoveRight, len(query)-pos))
				pos = len(query)
			} else if r == '\x06' { // Ctrl+F - move forward
				fmt.Printf(escMoveRight)
				pos++
			} else if r == '\x0B' { // Ctrl+K - delete to end of line
				fmt.Printf(strings.Repeat(" ", len(query)-pos))
				fmt.Printf(strings.Repeat(escMoveLeft, len(query)-pos))
				query = query[:pos]
			} else if r == '\x15' { // Ctrl+U - delete to start of line
				fmt.Printf(strings.Repeat(escMoveLeft, pos))
				fmt.Printf(string(query[pos:]) + strings.Repeat(" ", pos))
				fmt.Printf(strings.Repeat(escMoveLeft, len(query)))
				query = query[pos:]
				pos = 0
			} else if search && ' ' <= r {
				query = append(query[:pos], append([]rune{r}, query[pos:]...)...)
				fmt.Printf(string(query[pos:]) + strings.Repeat(escMoveLeft, len(query)-pos-1))
				pos++
			}
		}
	}()

	// go to bottom and clear output
	fmt.Printf(escMoveStart + escClearLine + strings.Repeat(escMoveDown+escClearLine, numLines))
	fmt.Printf(escMoveUpN, numLines)

	fmt.Printf("%v: ", label)
	if err != nil {
		if err == keyInterrupt {
			fmt.Printf("^C")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		} else if err == keyEscape {
			fmt.Printf("\n")
			return nil
		}
		fmt.Printf("\n")
		return err
	}

	selected = optionsIndex[selected]
	fmt.Printf("%v\n", options.Index(selected).Interface())

	if dst.Type() == options.Type().Elem() {
		dst.Set(options.Index(selected))
	} else {
		switch kind := dst.Kind(); kind {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dst.SetUint(uint64(selected))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dst.SetInt(int64(selected))
		default:
			return fmt.Errorf("unsupported destination type: %v", kind)
		}
	}
	return nil
}
