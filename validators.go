package prompt

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// Validator is a validator interface.
type Validator func(any) error

// StrLength matches if the input length is in the given range (inclusive). Use -1 for an open limit.
func StrLength(min, max int) Validator {
	return func(i any) error {
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
	return func(i any) error {
		var num float64
		switch v := i.(type) {
		case int:
			num = float64(v)
		case int8:
			num = float64(v)
		case int16:
			num = float64(v)
		case int32:
			num = float64(v)
		case int64:
			num = float64(v)
		case uint:
			num = float64(v)
		case uint8:
			num = float64(v)
		case uint16:
			num = float64(v)
		case uint32:
			num = float64(v)
		case uint64:
			num = float64(v)
		case float32:
			num = float64(v)
		case float64:
			num = v
		default:
			if inter, ok := i.(interface{ Int64() int64 }); ok {
				num = float64(inter.Int64())
			} else if floater, ok := i.(interface{ Float64() float64 }); ok {
				num = floater.Float64()
			} else {
				return fmt.Errorf("expected integer or floating point")
			}
		}
		if !math.IsNaN(min) && num < min || !math.IsNaN(max) && max < num {
			return fmt.Errorf("out of range [%v,%v]", min, max)
		}
		return nil
	}
}

// DateRange matches if the input is in the given time range (inclusive). Use time.Time's zero value for an open limit.
func DateRange(min, max time.Time) Validator {
	return func(i any) error {
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
	return func(i any) error {
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
	return func(i any) error {
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

// Before matches if the input is before the given number of date.
func Before(before any) Validator {
	return func(i any) error {
		if reflect.TypeOf(before) != reflect.TypeOf(i) {
			return fmt.Errorf("expected %v", reflect.TypeOf(before))
		}

		var ok bool
		switch v := i.(type) {
		case int:
			ok = v < before.(int)
		case int8:
			ok = v < before.(int8)
		case int16:
			ok = v < before.(int16)
		case int32:
			ok = v < before.(int32)
		case int64:
			ok = v < before.(int64)
		case uint:
			ok = v < before.(uint)
		case uint8:
			ok = v < before.(uint8)
		case uint16:
			ok = v < before.(uint16)
		case uint32:
			ok = v < before.(uint32)
		case uint64:
			ok = v < before.(uint64)
		case float32:
			ok = v < before.(float32)
		case float64:
			ok = v < before.(float64)
		case time.Time:
			ok = v.Before(before.(time.Time))
		default:
			return fmt.Errorf("expected integer, floating point, or timestamp")
		}
		if !ok {
			return fmt.Errorf("must be before %v", before)
		}
		return nil
	}
}

// After matches if the input is after the given number of date.
func After(after any) Validator {
	return func(i any) error {
		if reflect.TypeOf(after) != reflect.TypeOf(i) {
			return fmt.Errorf("expected %v", reflect.TypeOf(after))
		}

		var ok bool
		switch v := i.(type) {
		case int:
			ok = after.(int) < v
		case int8:
			ok = after.(int8) < v
		case int16:
			ok = after.(int16) < v
		case int32:
			ok = after.(int32) < v
		case int64:
			ok = after.(int64) < v
		case uint:
			ok = after.(uint) < v
		case uint8:
			ok = after.(uint8) < v
		case uint16:
			ok = after.(uint16) < v
		case uint32:
			ok = after.(uint32) < v
		case uint64:
			ok = after.(uint64) < v
		case float32:
			ok = after.(float32) < v
		case float64:
			ok = after.(float64) < v
		case time.Time:
			ok = v.After(after.(time.Time))
		default:
			return fmt.Errorf("expected integer, floating point, or timestamp")
		}
		if !ok {
			return fmt.Errorf("must be after %v", after)
		}
		return nil
	}
}

// Pattern matches the given pattern.
func Pattern(pattern, message string) Validator {
	re := regexp.MustCompile(pattern)
	return func(i any) error {
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
	return Pattern(`^[\w\.-]+@([a-z0-9][a-z0-9-]{0,61}[a-z0-9]\.)+[a-z0-9]{2,63}$`, "invalid e-mail address")
}

// TelephoneNumber matches a valid telephone number.
func TelephoneNumber() Validator {
	return Pattern(``, "invalid telephone number") // TODO
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
	return func(i any) error {
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
	return func(i any) error {
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
func Is(elem any) Validator {
	velem := reflect.ValueOf(elem)
	return func(i any) error {
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
func In(list any) Validator {
	vlist := reflect.ValueOf(list)
	if vlist.Kind() != reflect.Slice {
		panic("list must be a slice")
	}
	elemType := vlist.Type().Elem()
	return func(i any) error {
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
func NotIn(list any) Validator {
	vlist := reflect.ValueOf(list)
	if vlist.Kind() != reflect.Slice {
		panic("list must be a slice")
	}
	elemType := vlist.Type().Elem()
	return func(i any) error {
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
	return func(i any) error {
		if validator(i) != nil {
			return nil
		}
		return fmt.Errorf("not available")
	}
}

// And evaluates multiple validators using the logical AND operator, i.e. must satisfy all validators. This is only useful inside logical OR validators.
func And(validators ...Validator) Validator {
	return func(i any) error {
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
	return func(i any) error {
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
