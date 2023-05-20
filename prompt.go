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

var OptionSelected = fmt.Sprintf("  %v[\u00D7] %%v%v", escBold, escReset)
var OptionUnselected = "  [ ] %v"
var KeyInterrupt = fmt.Errorf("interrupt")

func clearlines(n int) {
	fmt.Printf(escMoveStart + escClearLine + strings.Repeat(escMoveUp+escClearLine, n-1))
}

type Validator func(interface{}) error

func StrLength(min, max int) Validator {
	return func(i interface{}) error {
		if s, ok := i.(string); ok {
			if len(s) < min {
				return fmt.Errorf("too short, minimum is %v", min)
			} else if max != -1 && max < len(s) {
				return fmt.Errorf("too long, maximum is %v", max)
			}
		} else {
			return fmt.Errorf("expected string")
		}
		return nil
	}
}

func NumRange(min, max float64) Validator {
	return func(i interface{}) error {
		if n, ok := i.(int); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(int8); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(int16); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(int32); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(int64); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(uint); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(uint8); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(uint16); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(uint32); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(uint64); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(float32); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(float64); ok {
			if n < min || max < n {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else {
			return fmt.Errorf("expected integer or floating point")
		}
		return nil
	}
}

func DateRange(min, max time.Time) Validator {
	return func(i interface{}) error {
		if t, ok := i.(time.Time); ok {
			if t.Before(min) || t.After(max) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else {
			return fmt.Errorf("expected timestamp")
		}
		return nil
	}
}

func Prefix(afix string) Validator {
	return func(i interface{}) error {
		if s, ok := i.(string); ok {
			if !strings.HasPrefix(s, afix) {
				return fmt.Errorf("expected prefix '%v'", afix)
			}
		} else {
			return fmt.Errorf("expected string")
		}
		return nil
	}
}

func Suffix(afix string) Validator {
	return func(i interface{}) error {
		if s, ok := i.(string); ok {
			if !strings.HasSuffix(s, afix) {
				return fmt.Errorf("expected suffix '%v'", afix)
			}
		} else {
			return fmt.Errorf("expected string")
		}
		return nil
	}
}

func Pattern(pattern, message string) Validator {
	re := regexp.MustCompile(pattern)
	return func(i interface{}) error {
		if s, ok := i.(string); ok {
			if !re.MatchString(s) {
				return fmt.Errorf(message)
			}
		} else {
			return fmt.Errorf("expected string")
		}
		return nil
	}
}

func EmailAddress() Validator {
	return Pattern(`^[\w\.-]+@[\w\.-]+\.\w{2,4}$`, "invalid e-mail address")
}

func IPAddress() Validator {
	return Pattern(`^([0-9]{1,3}\.){3}[0-9]{1,3}$|^(([a-fA-F0-9]{1,4}|):){1,7}([a-fA-F0-9]{1,4}|:)$`, "invalid IP address")
}

func IPv4Address() Validator {
	return Pattern(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`, "invalid IPv4 address")
}

func IPv6Address() Validator {
	return Pattern(`^(([a-fA-F0-9]{1,4}|):){1,7}([a-fA-F0-9]{1,4}|:)$`, "invalid IPv6 address")
}

func Port() Validator {
	return NumRange(1, 65535)
}

func Path() Validator {
	return Pattern(`^([^\/]+)?\/([^\/]+\/)*([^\/]+)?$`, "invalid path")
}

func AbsPath() Validator {
	return Pattern(`^\/([^\/]+\/)*([^\/]+)?$`, "invalid absolute path")
}

func UserName() Validator {
	return Pattern(`^[a-z_]([a-z0-9_-]{1,31}|[a-z0-9_-]{1,30}\$)$`, "invalid user name")
}

func TopDomainName() Validator {
	return Pattern(`^[a-z0-9][a-z0-9-]{0,61}[a-z0-9]\.[a-z0-9]{2,63}$`, "invalid top-level domain name")
}

func DomainName() Validator {
	return Pattern(`^([a-z0-9][a-z0-9-]{0,61}[a-z0-9]\.)+[a-z0-9]{2,63}$`, "invalid domain name")
}

func StrNot(items []string) Validator {
	return func(i interface{}) error {
		if s, ok := i.(string); ok {
			for _, item := range items {
				if item == s {
					return fmt.Errorf("unavailable: %v", s)
				}
			}
		} else {
			return fmt.Errorf("expected string")
		}
		return nil
	}
}

func Dir() Validator {
	return func(i interface{}) error {
		if s, ok := i.(string); ok {
			if info, err := os.Stat(s); err != nil {
				return fmt.Errorf("file not found: %v", s)
			} else if !info.Mode().IsDir() {
				return fmt.Errorf("path is not regular file: %v", s)
			}
		} else {
			return fmt.Errorf("expected string")
		}
		return nil
	}
}

func File() Validator {
	return func(i interface{}) error {
		if s, ok := i.(string); ok {
			if info, err := os.Stat(s); err != nil {
				return fmt.Errorf("file not found: %v", s)
			} else if !info.Mode().IsRegular() {
				return fmt.Errorf("path is not regular file: %v", s)
			}
		} else {
			return fmt.Errorf("expected string")
		}
		return nil
	}
}

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

func EmptyOr(validator Validator) Validator {
	return func(i interface{}) error {
		if s, ok := i.(string); ok && s == "" {
			return nil
		}
		return validator(i)
	}
}

func Enter(label string) {
	fmt.Printf("%v [enter]: ", label)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}

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
	}

	var b bool
	var err error
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

func Prompt(idst interface{}, label string, ideflt interface{}, validators ...Validator) error {
	first := true

	// get destination
	dst := reflect.ValueOf(idst)
	if dst.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be pointer to variable")
	}
	idst = dst.Elem().Interface()

Prompt:
	// prompt input
	if ideflt == nil || reflect.ValueOf(ideflt).IsZero() {
		fmt.Printf("%v: ", label)
	} else {
		fmt.Printf("%v (%v): ", label, ideflt)
	}
	var res string
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		res = strings.TrimSpace(scanner.Text())
	}

	// fill destination
	var err error
	ival := ideflt
	if res != "" || ival == nil {
		switch idst.(type) {
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
				if perr := scanner.Scan(res); perr != nil {
					err = fmt.Errorf("invalid %T: %w", idst, perr)
				}
				ival = nil
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

	if label == "Price" {
		return fmt.Errorf("test")
	}

	if err != nil {
		first = false
		fmt.Printf("%v%v%vERROR: %v%v%v", escClearLine, escRed, escBold, err, escReset, escMoveUp)
		clearlines(1)
		goto Prompt
	} else if !first {
		fmt.Printf(escClearLine)
	}
	if ival != nil {
		// otherwise already set using the Scanner interface
		dst.Elem().Set(reflect.ValueOf(ival))
	}
	return nil
}

func Select(idst interface{}, label string, ioptions interface{}, iselected interface{}) error {
	dst := reflect.ValueOf(idst)
	if dst.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be pointer to variable")
	}
	dst = dst.Elem()

	options := reflect.ValueOf(ioptions)
	if options.Kind() != reflect.Slice {
		return fmt.Errorf("options must be slice")
	}

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
		return fmt.Errorf("selected must be integer type or %v", options.Type().Elem().Name())
	}

	if options.Len() == 0 {
		return fmt.Errorf("no options")
	} else if 256 <= options.Len() {
		return fmt.Errorf("too many options")
	} else if selected < 0 {
		selected = 0
	} else if options.Len() <= selected {
		selected = options.Len() - 1
	}

	// print options
	fmt.Printf("%v:", label)
	for i := 0; i < options.Len(); i++ {
		opt := options.Index(i).Interface()
		if i == selected {
			fmt.Printf("\n"+OptionSelected, opt)
		} else {
			fmt.Printf("\n"+OptionUnselected, opt)
		}
	}

	// go to selected option
	fmt.Printf(escMoveStart + strings.Repeat(escMoveUp, options.Len()-1-selected))

	// hide input
	restore, err := MakeRaw()
	if err != nil {
		return err
	}

	// read input
	input := bufio.NewReader(os.Stdin)
	for {
		var r, key rune
		r, _, err = input.ReadRune()
		if err != nil {
			break
		} else if r == '\x1B' { // escape
			key, _, err = input.ReadRune()
			if err != nil {
				break
			} else if key == '[' {
				key, _, err = input.ReadRune()
				if err != nil {
					break
				}
			}
		}

		if r == '\x03' { // interrupt
			err = KeyInterrupt
			break
		} else if r == '\x04' || r == '\r' || r == '\n' { // select
			break
		} else if selected != 0 && (key == 'A' || r == 'w' || r == 'k') { // up
			fmt.Printf(escMoveStart+escClearLine+OptionUnselected, options.Index(selected).Interface())
			fmt.Printf(escMoveUp)
			selected--
			fmt.Printf(escMoveStart+escClearLine+OptionSelected, options.Index(selected).Interface())
		} else if selected != options.Len()-1 && (key == 'B' || r == 's' || r == 'j' || r == '\t') { // down
			fmt.Printf(escMoveStart+escClearLine+OptionUnselected, options.Index(selected).Interface())
			fmt.Printf(escMoveDown)
			selected++
			fmt.Printf(escMoveStart+escClearLine+OptionSelected, options.Index(selected).Interface())
		} else if key == 'H' || selected == options.Len()-1 && r == '\t' { // home
			fmt.Printf(escMoveStart+escClearLine+OptionUnselected, options.Index(selected).Interface())
			fmt.Printf(strings.Repeat(escMoveUp, selected))
			selected = 0
			fmt.Printf(escMoveStart+escClearLine+OptionSelected, options.Index(selected).Interface())
		} else if key == 'F' { // end
			fmt.Printf(escMoveStart+escClearLine+OptionUnselected, options.Index(selected).Interface())
			fmt.Printf(strings.Repeat(escMoveDown, options.Len()-1-selected))
			selected = options.Len() - 1
			fmt.Printf(escMoveStart+escClearLine+OptionSelected, options.Index(selected).Interface())
		}
	}
	restore()

	// go to bottom and clear output
	fmt.Printf(strings.Repeat(escMoveDown, options.Len()-1-selected))
	clearlines(options.Len() + 1)

	fmt.Printf("%v: ", label)
	if err != nil {
		if err == KeyInterrupt {
			fmt.Printf("^C")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
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
