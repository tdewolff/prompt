package prompt

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/araddon/dateparse"
)

var optionSelected = fmt.Sprintf("  %v[\u00D7] %%v%v", escBold, escReset)
var optionUnselected = "  [ ] %v"
var keyInterrupt = fmt.Errorf("interrupt")
var keyEscape = fmt.Errorf("escape")

func clearlines(n int) {
	fmt.Printf(escMoveStart + escClearLine + strings.Repeat(escMoveUp+escClearLine, n-1))
}

// Validator is a validator interface.
type Validator func(interface{}) error

// StrLength matches if the input length is in the given range (inclusive). Use -1 for an open limit.
func StrLength(min, max int) Validator {
	return func(i interface{}) error {
		var str string
		if s, ok := i.(string); ok {
			str = s
		} else if stringer, ok := i.(interface{ String() string }); ok {
			str = stringer.String()
		} else {
			return fmt.Errorf("expected string")
		}
		if len(str) < min {
			return fmt.Errorf("too short, minimum is %v", min)
		} else if max != -1 && max < len(str) {
			return fmt.Errorf("too long, maximum is %v", max)
		}
		return nil
	}
}

// NumRange matches if the input is in the given number range (inclusive). Use NaN or +/-Inf for an open limit.
func NumRange(min, max float64) Validator {
	return func(i interface{}) error {
		var num float64
		if n, ok := i.(int); ok {
			num = float64(n)
		} else if n, ok := i.(int8); ok {
			num = float64(n)
		} else if n, ok := i.(int16); ok {
			num = float64(n)
		} else if n, ok := i.(int32); ok {
			num = float64(n)
		} else if n, ok := i.(int64); ok {
			num = float64(n)
		} else if n, ok := i.(uint); ok {
			num = float64(n)
		} else if n, ok := i.(uint8); ok {
			num = float64(n)
		} else if n, ok := i.(uint16); ok {
			num = float64(n)
		} else if n, ok := i.(uint32); ok {
			num = float64(n)
		} else if n, ok := i.(uint64); ok {
			num = float64(n)
		} else if n, ok := i.(float32); ok {
			num = float64(n)
		} else if n, ok := i.(float64); ok {
			num = n
		} else if inter, ok := i.(interface{ Int64() int64 }); ok {
			num = float64(inter.Int64())
		} else if floater, ok := i.(interface{ Float64() float64 }); ok {
			num = floater.Float64()
		} else {
			return fmt.Errorf("expected integer or floating point")
		}
		if !math.IsNaN(min) && num < min || !math.IsNaN(max) && max < num {
			return fmt.Errorf("out of range [%v,%v]", min, max)
		}
		return nil
	}
}

// DateRange matches if the input is in the given time range (inclusive). Use time.Time's zero value for an open limit.
func DateRange(min, max time.Time) Validator {
	return func(i interface{}) error {
		if t, ok := i.(time.Time); ok {
			if !min.IsZero() && t.Before(min) || !max.IsZero() && t.After(max) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else {
			return fmt.Errorf("expected timestamp")
		}
		return nil
	}
}

// Prefix matches if the input has the given prefix.
func Prefix(afix string) Validator {
	return func(i interface{}) error {
		var str string
		if s, ok := i.(string); ok {
			str = s
		} else if stringer, ok := i.(interface{ String() string }); ok {
			str = stringer.String()
		} else {
			return fmt.Errorf("expected string")
		}
		if !strings.HasPrefix(str, afix) {
			return fmt.Errorf("expected prefix '%v'", afix)
		}
		return nil
	}
}

// Suffix matches if the input has the given suffix.
func Suffix(afix string) Validator {
	return func(i interface{}) error {
		var str string
		if s, ok := i.(string); ok {
			str = s
		} else if stringer, ok := i.(interface{ String() string }); ok {
			str = stringer.String()
		} else {
			return fmt.Errorf("expected string")
		}
		if !strings.HasSuffix(str, afix) {
			return fmt.Errorf("expected suffix '%v'", afix)
		}
		return nil
	}
}

// Pattern matches the given pattern.
func Pattern(pattern, message string) Validator {
	re := regexp.MustCompile(pattern)
	return func(i interface{}) error {
		var str string
		if s, ok := i.(string); ok {
			str = s
		} else if stringer, ok := i.(interface{ String() string }); ok {
			str = stringer.String()
		} else {
			return fmt.Errorf("expected string")
		}
		if !re.MatchString(str) {
			return fmt.Errorf(message)
		}
		return nil
	}
}

// EmailAddress matches a valid e-mail address.
func EmailAddress() Validator {
	return Pattern(`^[\w\.-]+@[\w\.-]+\.\w{2,4}$`, "invalid e-mail address")
}

// IPAddress matches an IPv4 or IPv6 address.
func IPAddress() Validator {
	return Pattern(`^([0-9]{1,3}\.){3}[0-9]{1,3}$|^(([a-fA-F0-9]{1,4}|):){1,7}([a-fA-F0-9]{1,4}|:)$`, "invalid IP address")
}

// IPv4Address matches an IPv4 address.
func IPv4Address() Validator {
	return Pattern(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`, "invalid IPv4 address")
}

// IPv6Address matches an IPv6 address.
func IPv6Address() Validator {
	return Pattern(`^(([a-fA-F0-9]{1,4}|):){1,7}([a-fA-F0-9]{1,4}|:)$`, "invalid IPv6 address")
}

// Port matches a valid port number.
func Port() Validator {
	return NumRange(1, 65535)
}

// Path matches any file path.
func Path() Validator {
	return Pattern(`^([^\/]+)?\/([^\/]+\/)*([^\/]+)?$`, "invalid path")
}

// AbsolutePath matches an absolute file path.
func AbsolutePath() Validator {
	return Pattern(`^\/([^\/]+\/)*([^\/]+)?$`, "invalid absolute path")
}

// UserName matches a valid Unix user name.
func UserName() Validator {
	return Pattern(`^[a-z_]([a-z0-9_-]{1,31}|[a-z0-9_-]{1,30}\$)$`, "invalid user name")
}

// TopDomainName matches a top-level domain name.
func TopDomainName() Validator {
	return Pattern(`^[a-z0-9][a-z0-9-]{0,61}[a-z0-9]\.[a-z0-9]{2,63}$`, "invalid top-level domain name")
}

// DomainName matches a domain name.
func DomainName() Validator {
	return Pattern(`^([a-z0-9][a-z0-9-]{0,61}[a-z0-9]\.)+[a-z0-9]{2,63}$`, "invalid domain name")
}

// FQDN matches a fully qualified domain name.
func FQDN() Validator {
	return Pattern(`^([a-z0-9][a-z0-9-]{0,61}[a-z0-9]\.)+[a-z0-9]{2,63}\.$`, "invalid fully qualified domain name")
}

// Dir matches a path to an existing directory on the system.
func Dir() Validator {
	return func(i interface{}) error {
		var str string
		if s, ok := i.(string); ok {
			str = s
		} else if stringer, ok := i.(interface{ String() string }); ok {
			str = stringer.String()
		} else {
			return fmt.Errorf("expected string")
		}
		if info, err := os.Stat(str); err != nil {
			return fmt.Errorf("file not found: %v", str)
		} else if !info.Mode().IsDir() {
			return fmt.Errorf("path is not regular file: %v", str)
		}
		return nil
	}
}

// File matches a path to an existing file on the system.
func File() Validator {
	return func(i interface{}) error {
		var str string
		if s, ok := i.(string); ok {
			str = s
		} else if stringer, ok := i.(interface{ String() string }); ok {
			str = stringer.String()
		} else {
			return fmt.Errorf("expected string")
		}
		if info, err := os.Stat(str); err != nil {
			return fmt.Errorf("file not found: %v", str)
		} else if !info.Mode().IsRegular() {
			return fmt.Errorf("path is not regular file: %v", str)
		}
		return nil
	}
}

// Is matches if the input matches the given value.
func Is(elem interface{}) Validator {
	velem := reflect.ValueOf(elem)
	return func(i interface{}) error {
		v := reflect.ValueOf(i)
		if v.Type() != velem.Type() {
			return fmt.Errorf("expected %v", velem.Type().Name())
		} else if !velem.Equal(v) {
			return fmt.Errorf("expected '%v'", elem)
		}
		return nil
	}
}

// In matches if the input matches any element of the list.
func In(list interface{}) Validator {
	vlist := reflect.ValueOf(list)
	if vlist.Kind() != reflect.Slice {
		panic("list must be a slice")
	}
	elemType := vlist.Type().Elem()
	return func(i interface{}) error {
		v := reflect.ValueOf(i)
		if v.Type() != elemType {
			return fmt.Errorf("expected %v", elemType.Name())
		}
		for j := 0; j < vlist.Len(); j++ {
			if vlist.Index(j).Equal(v) {
				return nil
			}
		}
		return fmt.Errorf("not available")
	}
}

// NotIn matches if the input does not match any element of the list.
func NotIn(list interface{}) Validator {
	vlist := reflect.ValueOf(list)
	if vlist.Kind() != reflect.Slice {
		panic("list must be a slice")
	}
	elemType := vlist.Type().Elem()
	return func(i interface{}) error {
		v := reflect.ValueOf(i)
		if v.Type() != elemType {
			return fmt.Errorf("expected %v", elemType.Name())
		}
		for j := 0; j < vlist.Len(); j++ {
			if vlist.Index(j).Equal(v) {
				return fmt.Errorf("not available")
			}
		}
		return nil
	}
}

// Not evaluates the validator using the logical NOT operator, i.e. satisfies if the validator does not satisfy.
func Not(validator Validator) Validator {
	return func(i interface{}) error {
		if validator(i) != nil {
			return nil
		}
		return fmt.Errorf("not available")
	}
}

// And evaluates multiple validators using the logical AND operator, i.e. must satisfy all validators. This is only useful inside logical OR validators.
func And(validators ...Validator) Validator {
	return func(i interface{}) error {
		if len(validators) == 0 {
			return nil
		}
		for _, val := range validators {
			if err := val(i); err != nil {
				return err
			}
		}
		return nil
	}
}

// Or evaluates multiple validators using the logical OR operator, i.e. at least one validator must be satisfied.
func Or(validators ...Validator) Validator {
	return func(i interface{}) error {
		if len(validators) == 0 {
			return nil
		}
		for _, val := range validators {
			if err := val(i); err == nil {
				return nil
			}
		}
		return fmt.Errorf("not available")
	}
}

// Enter is a prompt that requires the Enter key to continue.
func Enter(label string) {
	fmt.Printf("%v [enter]: ", label)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
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

	var res string
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		res = strings.TrimSpace(scanner.Text())
	}
	if res == "" {
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
		clearlines(1)
		goto Prompt
	} else if !first {
		fmt.Printf(escClearLine)
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
	} else {
		fmt.Printf("%v: %v", label, string(result))
		fmt.Printf(strings.Repeat(escMoveLeft, len(result)-pos))
	}

	// make raw and hide input
	restore, err := MakeRaw(false)
	if err != nil {
		return err
	}

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
	restore()

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
			ival = res
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
		clearlines(1)
		goto Prompt
	} else if !first {
		fmt.Printf(escClearLine)
	}
	dst.Elem().Set(reflect.ValueOf(ival))
	return nil
}

func getSelected(iselected interface{}, options reflect.Value) (int, error) {
	var selected int
	if reflect.TypeOf(iselected) == options.Type().Elem() {
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
	if selected < 0 {
		selected = 0
	} else if options.Len() <= selected {
		selected = options.Len() - 1
	}
	return selected, nil
}

// Select is a list selection prompt that allows to select one of the list of possible values. The ioptions must be a slice of options. The idst must be a pointer to a variable and must of of the same type as the options (set the option value) or an integer (set the option index). Equally, iselected can be an integer (index) or of the same type as the options (value).
// Users can select an option using Up or W or K to move up, Down or S or J to move down, Tab and Shift+Tab to move down and up respectively and wrap around, Ctrl+C or Escape to quit, and Ctrl+Z or Enter to select an option.
func Select(idst interface{}, label string, ioptions interface{}, iselected interface{}) error {
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
	} else if 256 <= options.Len() {
		return fmt.Errorf("too many options")
	}

	selected, err := getSelected(iselected, options)

	// print options
	fmt.Printf("%v:", label)
	for i := 0; i < options.Len(); i++ {
		opt := options.Index(i).Interface()
		if i == selected {
			fmt.Printf("\n"+optionSelected, opt)
		} else {
			fmt.Printf("\n"+optionUnselected, opt)
		}
	}

	// go to selected option
	fmt.Printf(escMoveStart + strings.Repeat(escMoveUp, options.Len()-1-selected))

	// make raw and hide input
	restore, err := MakeRaw(true)
	if err != nil {
		return err
	}

	// read input
	input := bufio.NewReader(os.Stdin)
	for {
		var r, csi rune
		if r, _, err = input.ReadRune(); err != nil {
			break
		} else if r == '\x1B' { // escape
			if input.Buffered() == 0 {
				err = keyEscape
				break
			} else if r, _, err = input.ReadRune(); err != nil {
				break
			} else if r == '[' {
				csi, _, err = input.ReadRune()
				if err != nil {
					break
				}
			}
		}

		if r == '\x03' { // interrupt
			err = keyInterrupt
			break
		} else if r == '\x04' || r == '\r' || r == '\n' { // select
			break
		} else if selected != 0 && (csi == 'A' || r == 'w' || r == 'k' || csi == '\x5A') { // up
			fmt.Printf(escMoveStart+escClearLine+optionUnselected, options.Index(selected).Interface())
			fmt.Printf(escMoveUp)
			selected--
			fmt.Printf(escMoveStart+escClearLine+optionSelected, options.Index(selected).Interface())
		} else if selected != options.Len()-1 && (csi == 'B' || r == 's' || r == 'j' || r == '\t') { // down
			fmt.Printf(escMoveStart+escClearLine+optionUnselected, options.Index(selected).Interface())
			fmt.Printf(escMoveDown)
			selected++
			fmt.Printf(escMoveStart+escClearLine+optionSelected, options.Index(selected).Interface())
		} else if csi == 'H' || selected == options.Len()-1 && r == '\t' { // home
			fmt.Printf(escMoveStart+escClearLine+optionUnselected, options.Index(selected).Interface())
			fmt.Printf(strings.Repeat(escMoveUp, selected))
			selected = 0
			fmt.Printf(escMoveStart+escClearLine+optionSelected, options.Index(selected).Interface())
		} else if csi == 'F' || selected == 0 && csi == '\x5A' { // end
			fmt.Printf(escMoveStart+escClearLine+optionUnselected, options.Index(selected).Interface())
			fmt.Printf(strings.Repeat(escMoveDown, options.Len()-1-selected))
			selected = options.Len() - 1
			fmt.Printf(escMoveStart+escClearLine+optionSelected, options.Index(selected).Interface())
		}
	}
	restore()

	// go to bottom and clear output
	fmt.Printf(strings.Repeat(escMoveDown, options.Len()-1-selected))
	clearlines(options.Len() + 1)

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
