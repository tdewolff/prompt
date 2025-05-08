package prompt

import (
	"fmt"
	"reflect"
)

func getChecked(dst, options reflect.Value) ([]bool, error) {
	checked := make([]bool, options.Len())
	if dst.Type().Elem() == options.Type().Elem() {
		for j := 0; j < dst.Len(); j++ {
			for i := 0; i < len(checked); i++ {
				if options.Index(i).Equal(dst.Index(j)) {
					checked[i] = true
					break
				}
			}
		}
	} else if k := dst.Elem().Kind(); k == reflect.Bool {
		for j := 0; j < dst.Len(); j++ {
			checked[j] = dst.Index(j).Bool()
		}
	} else if k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 {
		for j := 0; j < dst.Len(); j++ {
			i := int(dst.Index(j).Int())
			checked[i] = true
		}
	} else if k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 {
		for j := 0; j < dst.Len(); j++ {
			i := int(dst.Index(j).Uint())
			checked[i] = true
		}
	} else {
		return nil, fmt.Errorf("destination must be a boolean or integer type or a slice of %v", options.Type().Elem())
	}
	return checked, nil
}

func Checklist(idst interface{}, label string, ioptions interface{}) error {
	dst := reflect.ValueOf(idst)
	options := reflect.ValueOf(ioptions)
	if dst.Kind() != reflect.Pointer || dst.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("destination must be a pointer to slice")
	} else if options.Kind() != reflect.Slice {
		return fmt.Errorf("options must be a slice")
	} else if options.Len() == 0 {
		return fmt.Errorf("no options")
	}
	dst = dst.Elem()

	checked, err := getChecked(dst, options)
	if err != nil {
		return err
	}

	optionStrings := make([]string, options.Len())
	for i := 0; i < options.Len(); i++ {
		optionStrings[i] = fmt.Sprint(options.Index(i).Interface())
	}

	// set constants
	selected := 0
	maxLines := selectMaxLines
	if _, rows, err := TerminalSize(); err != nil {
		return err
	} else if rows-1 < maxLines {
		maxLines = rows - 1 // keep one for prompt row
	}
	scrollOffset := selectScrollOffset
	withQuery := maxLines < options.Len() || 10 < options.Len()
	enterSelects := true

	label += " (space selects)"
	err = terminalList(label, optionStrings, selected, maxLines, scrollOffset, withQuery, enterSelects, func(i, selected int) string {
		s := "[ ] %v"
		if checked[i] {
			s = "[\u00D7] %v"
		}
		if i == selected {
			s = escBold + s + escReset
		}
		return s
	}, func(r rune, i int) {
		if r == ' ' || r == '\n' || r == '\r' {
			checked[i] = !checked[i]
		}
	})

	fmt.Printf("%v: ", label)
	if err != nil {
		if err == keyInterrupt {
			fmt.Printf("^C")
		}
		fmt.Printf("\n")
		return err
	}

	first := true
	for i := 0; i < len(optionStrings); i++ {
		if checked[i] {
			if !first {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", optionStrings[i])
			first = false
		}
	}
	fmt.Println()

	value := reflect.MakeSlice(dst.Type(), 0, options.Len())
	if dst.Type().Elem() == options.Type().Elem() {
		for i := 0; i < options.Len(); i++ {
			if checked[i] {
				value = reflect.Append(value, options.Index(i))
			}
		}
	} else {
		switch kind := dst.Elem().Kind(); kind {
		case reflect.Bool:
			for i := 0; i < options.Len(); i++ {
				value = reflect.Append(value, reflect.ValueOf(checked[i]))
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			for i := 0; i < options.Len(); i++ {
				if checked[i] {
					value = reflect.Append(value, reflect.ValueOf(i).Convert(dst.Type().Elem()))
				}
			}
		default:
			return fmt.Errorf("unsupported destination type: %v", kind)
		}
	}
	dst.Set(value)
	return nil
}
