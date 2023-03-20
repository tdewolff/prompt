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
			} else if max < len(s) {
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
		if n, ok := i.(int64); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(uint64); ok {
			if float64(n) < min || max < float64(n) {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else if n, ok := i.(float64); ok {
			if n < min || max < n {
				return fmt.Errorf("out of range [%v,%v]", min, max)
			}
		} else {
			return fmt.Errorf("expected int or uint")
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

func YesNo(label string, deflt bool) bool {
	first := true

Prompt:
	if deflt {
		fmt.Printf("%v (Y/n): ", label)
	} else {
		fmt.Printf("%v (y/N): ", label)
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
		fmt.Printf("%v%vERROR: %v%v%v", escRed, escBold, err, escReset, escMoveUp)
		clearlines(1)
		goto Prompt
	} else if !first {
		fmt.Printf(escClearLine)
	}
	return b
}

func Prompt(idst interface{}, label string, ideflt interface{}, validators ...Validator) error {
	first := true

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

	// get destination
	dst := reflect.ValueOf(idst)
	if dst.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be pointer to variable")
	}
	dst = dst.Elem()
	kind := dst.Kind()

	// fill destination
	var err error
	ival := ideflt
	if res != "" || ival == nil {
		switch kind {
		case reflect.String:
			ival = res
		case reflect.Bool:
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
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var max int64 = math.MaxInt64
			if kind == reflect.Int {
				max = math.MaxInt
			} else if kind == reflect.Int8 {
				max = math.MaxInt8
			} else if kind == reflect.Int16 {
				max = math.MaxInt16
			} else if kind == reflect.Int32 {
				max = math.MaxInt32
			}

			i, perr := strconv.ParseInt(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid integer")
			} else if max < i {
				err = fmt.Errorf("integer overflow")
			}
			ival = i
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var max uint64 = math.MaxUint64
			if kind == reflect.Uint {
				max = math.MaxUint
			} else if kind == reflect.Uint8 {
				max = math.MaxUint8
			} else if kind == reflect.Uint16 {
				max = math.MaxUint16
			} else if kind == reflect.Uint32 {
				max = math.MaxUint32
			}

			u, perr := strconv.ParseUint(res, 10, 64)
			if perr != nil {
				err = fmt.Errorf("invalid positive integer")
			} else if max < u {
				err = fmt.Errorf("unsigned integer overflow")
			}
			ival = u
		case reflect.Float32, reflect.Float64:
			bitsize := 64
			if kind == reflect.Float32 {
				bitsize = 32
			}
			f, perr := strconv.ParseFloat(res, bitsize)
			if perr.(*strconv.NumError).Err == strconv.ErrRange {
				err = fmt.Errorf("floating point overflow")
			} else if perr != nil {
				err = fmt.Errorf("invalid floating point")
			}
			ival = f
		default:
			return fmt.Errorf("unsupported destination type: %v", kind)
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
		fmt.Printf("%v%vERROR: %v%v%v", escRed, escBold, err, escReset, escMoveUp)
		clearlines(1)
		goto Prompt
	} else if !first {
		fmt.Printf(escClearLine)
	}

	switch kind {
	case reflect.Int:
		ival = int(ival.(int64))
	case reflect.Int8:
		ival = int8(ival.(int64))
	case reflect.Int16:
		ival = int16(ival.(int64))
	case reflect.Int32:
		ival = int32(ival.(int64))
	case reflect.Uint:
		ival = uint(ival.(uint64))
	case reflect.Uint8:
		ival = uint8(ival.(uint64))
	case reflect.Uint16:
		ival = uint16(ival.(uint64))
	case reflect.Uint32:
		ival = uint32(ival.(uint64))
	case reflect.Float32:
		ival = float32(ival.(float64))
	}
	dst.Set(reflect.ValueOf(ival))
	return nil
}

func Select(idst interface{}, label string, options []string, iselected interface{}) error {
	var selected int
	if i, ok := iselected.(int); ok {
		selected = i
	} else if s, ok := iselected.(string); ok {
		for i, option := range options {
			if option == s {
				selected = i
				break
			}
		}
	} else {
		return fmt.Errorf("selected must be int or string")
	}

	if len(options) == 0 {
		return fmt.Errorf("no options")
	} else if 256 <= len(options) {
		return fmt.Errorf("too many options")
	} else if selected < 0 {
		selected = 0
	} else if len(options) <= selected {
		selected = len(options) - 1
	}

	// print options
	fmt.Printf("%v:", label)
	for i, opt := range options {
		if i == selected {
			fmt.Printf("\n"+OptionSelected, opt)
		} else {
			fmt.Printf("\n"+OptionUnselected, opt)
		}
	}

	// go to selected option
	fmt.Printf(escMoveStart + strings.Repeat(escMoveUp, len(options)-1-selected))

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
			fmt.Printf(escMoveStart+escClearLine+OptionUnselected, options[selected])
			fmt.Printf(escMoveUp)
			selected--
			fmt.Printf(escMoveStart+escClearLine+OptionSelected, options[selected])
		} else if selected != len(options)-1 && (key == 'B' || r == 's' || r == 'j' || r == '\t') { // down
			fmt.Printf(escMoveStart+escClearLine+OptionUnselected, options[selected])
			fmt.Printf(escMoveDown)
			selected++
			fmt.Printf(escMoveStart+escClearLine+OptionSelected, options[selected])
		} else if key == 'H' || selected == len(options)-1 && r == '\t' { // home
			fmt.Printf(escMoveStart+escClearLine+OptionUnselected, options[selected])
			fmt.Printf(strings.Repeat(escMoveUp, selected))
			selected = 0
			fmt.Printf(escMoveStart+escClearLine+OptionSelected, options[selected])
		} else if key == 'F' { // end
			fmt.Printf(escMoveStart+escClearLine+OptionUnselected, options[selected])
			fmt.Printf(strings.Repeat(escMoveDown, len(options)-1-selected))
			selected = len(options) - 1
			fmt.Printf(escMoveStart+escClearLine+OptionSelected, options[selected])
		}
	}
	restore()

	// go to bottom and clear output
	fmt.Printf(strings.Repeat(escMoveDown, len(options)-1-selected))
	clearlines(len(options) + 1)

	fmt.Printf("%v: ", label)
	if err != nil {
		if err == KeyInterrupt {
			fmt.Printf("^C")
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}
		return err
	}
	fmt.Printf("%v\n", options[selected])

	dst := reflect.ValueOf(idst)
	if dst.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be pointer to variable")
	}
	dst = dst.Elem()
	switch kind := dst.Kind(); kind {
	case reflect.String:
		dst.SetString(options[selected])
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dst.SetUint(uint64(selected))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dst.SetInt(int64(selected))
	default:
		return fmt.Errorf("unsupported destination type: %v", kind)
	}
	return nil
}
