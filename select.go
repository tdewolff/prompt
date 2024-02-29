package prompt

import (
	"fmt"
	"reflect"
)

func getSelected(dst, options reflect.Value) (int, error) {
	var selected int
	if dst.Type() == options.Type().Elem() {
		for i := 0; i < options.Len(); i++ {
			if options.Index(i).Equal(dst) {
				selected = i
				break
			}
		}
	} else if k := dst.Kind(); k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 {
		selected = int(dst.Int())
	} else if k == reflect.Uint || k == reflect.Uint8 || k == reflect.Uint16 || k == reflect.Uint32 || k == reflect.Uint64 {
		selected = int(dst.Uint())
	} else {
		return 0, fmt.Errorf("destination must be an integer type or %v", options.Type().Elem())
	}
	if selected < 0 {
		selected = 0
	} else if options.Len() <= selected {
		selected = options.Len() - 1
	}
	return selected, nil
}

// Select is a list selection prompt that allows to select one of the list of possible values. The ioptions must be a slice of options. The idst must be a pointer to a variable and must of the same type as the options (set the option value) or an integer (set the option index). The value od idst determines the initial selected value.
// Users can select an option using Up or W or K to move up, Down or S or J to move down, Tab and Shift+Tab to move down and up respectively and wrap around, Ctrl+C or Escape to quit, and Ctrl+Z or Enter to select an option.
func Select(idst interface{}, label string, ioptions interface{}) error {
	dst := reflect.ValueOf(idst)
	options := reflect.ValueOf(ioptions)
	if dst.Kind() != reflect.Pointer {
		return fmt.Errorf("destination must be a pointer to a variable")
	} else if options.Kind() != reflect.Slice {
		return fmt.Errorf("options must be a slice")
	} else if options.Len() == 0 {
		return fmt.Errorf("no options")
	}
	dst = dst.Elem()

	optionStrings := make([]string, options.Len())
	for i := 0; i < options.Len(); i++ {
		optionStrings[i] = fmt.Sprint(options.Index(i).Interface())
	}

	selected, err := getSelected(dst, options)
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
	scrollOffset := selectScrollOffset
	withQuery := maxLines < options.Len() || 10 < options.Len()
	exitEnter := true

	err = terminalList(label, optionStrings, selected, maxLines, scrollOffset, withQuery, exitEnter, func(i, selected int) string {
		if i == selected {
			return optionSelected
		}
		return optionUnselected
	}, func(r rune, i int) {
		if r == '\n' || r == '\r' {
			selected = i
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

	fmt.Printf("%v\n", optionStrings[selected])

	if dst.Type() == options.Type().Elem() {
		dst.Set(options.Index(selected))
	} else {
		switch kind := dst.Kind(); kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			dst.SetInt(int64(selected))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			dst.SetUint(uint64(selected))
		default:
			return fmt.Errorf("unsupported destination type: %v", kind)
		}
	}
	return nil
}
