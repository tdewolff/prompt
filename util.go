package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func Clip(x, a, b int) int {
	if x < a {
		return a
	} else if b < x {
		return b
	}
	return x
}

func matchOption(query, option string) bool {
	return strings.Contains(strings.ToLower(option), strings.ToLower(query))
}

func terminalList(label string, options []string, selected, maxLines, scrollOffset int, withQuery bool, optionMarkup func(int, int) string, keyPress func(rune, int)) error {
	fmt.Printf("%v:", label)

	padding := ""
	if 2 < len(label) && len(label) < 20 {
		padding = strings.Repeat(" ", len(label)-2)
	}

	// print options
	numLines := Min(maxLines, len(options))
	if (numLines-1)/2 < scrollOffset {
		scrollOffset = (numLines - 1) / 2
	}
	windowStart := Clip(selected-(numLines-1)/2, 0, len(options)-numLines)
	for i := 0; i < numLines; i++ {
		fmt.Printf("\n"+padding+optionMarkup(windowStart+i, selected), options[windowStart+i])
	}
	// go to query
	fmt.Printf(escMoveUpN+escMoveToCol, numLines, len(label)+3)
	defer func() {
		// go to bottom and clear output
		fmt.Printf(escMoveStart + escClearLine + strings.Repeat(escMoveDown+escClearLine, numLines))
		fmt.Printf(escMoveUpN, numLines)
	}()

	// option index in current view to option index in options
	optionsIndex := make([]int, len(options))
	for i := 0; i < len(options); i++ {
		optionsIndex[i] = i
	}

	// make raw and hide input
	restore, err := MakeRawTerminal(!withQuery)
	if err != nil {
		return err
	}
	defer restore()

	pos := 0 // position in query
	var prevQuery, query []rune
	prevSelected := selected

	// read input
	input := bufio.NewReader(os.Stdin)
	for {
		// change query results
		if withQuery && string(query) != string(prevQuery) {
			fmt.Printf(escMoveStart+escClearLine+"%v: %v"+escMoveToCol, label, string(query), len(label)+3+pos)
			i := 0
			hasSelected := false
			optionsIndex = optionsIndex[:0]
			for i < len(options) {
				if matchOption(string(query), options[i]) {
					if i == selected {
						selected = len(optionsIndex)
						hasSelected = true
					}
					optionsIndex = append(optionsIndex, i)
				}
				i++
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
				prevSelected, selected = 0, 0
			} else {
				prevSelected = -1
				if !hasSelected {
					selected = 0
				}
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
					j := optionsIndex[windowStart+i]
					fmt.Printf(escMoveDown+escMoveStart+escClearLine+padding+optionMarkup(j, optionsIndex[selected]), options[j])
				}
				// go to query
				fmt.Printf(escMoveUpN+escMoveToCol, numLines, len(label)+3+pos)
			} else {
				jPrev, j := optionsIndex[prevSelected], optionsIndex[selected]
				fmt.Printf(escMoveDownN+escMoveStart+escClearLine+padding+optionMarkup(jPrev, j), prevSelected-windowStart+1, options[jPrev])
				if selected < prevSelected {
					fmt.Printf(escMoveUpN, prevSelected-selected)
				} else {
					fmt.Printf(escMoveDownN, selected-prevSelected)
				}
				j = optionsIndex[selected]
				fmt.Printf(escMoveStart+escClearLine+padding+optionMarkup(j, j), options[j])
				// go to query
				fmt.Printf(escMoveUpN+escMoveToCol, selected-windowStart+1, len(label)+3+pos)
			}
			prevSelected = selected
		} else if 0 < len(optionsIndex) {
			j := optionsIndex[selected]
			fmt.Printf(escMoveDownN+escMoveStart+escClearLine+padding+optionMarkup(j, j), selected-windowStart+1, options[j])
			// go to query
			fmt.Printf(escMoveUpN+escMoveToCol, selected-windowStart+1, len(label)+3+pos)
		}

		// read user input
		var r rune
		if r, _, err = input.ReadRune(); err != nil {
			return err
		}

		if r == '\x03' { // interrupt
			return keyInterrupt
		} else if r == '\x04' || r == '\x26' { // Ctrl+D, Ctrl-Z
			keyPress(r, optionsIndex[selected])
			return nil
		} else if r == ' ' { // space
			keyPress(r, optionsIndex[selected])
		} else if r == '\r' || r == '\n' { // return, enter
			keyPress(r, optionsIndex[selected])
			return nil
		} else if r == '\x7F' { // backspace
			if pos != 0 {
				query = append(query[:pos-1], query[pos:]...)
				pos--
				fmt.Printf(escMoveLeft+"%v "+strings.Repeat(escMoveLeft, len(query)+1-pos), string(query[pos:]))
			}
		} else if r == '\x1B' { // escape
			if input.Buffered() == 0 {
				return keyEscape
			} else if r, _, err = input.ReadRune(); err != nil {
				return err
			} else if r == '[' { // CSI
				if input.Buffered() == 0 {
					// ignore
				} else if r, _, err = input.ReadRune(); err != nil {
					return err
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
						return err
					} else if tilde == '~' {
						if r == '3' { // delete
							if pos != len(query) {

								query = append(query[:pos], query[pos+1:]...)
								fmt.Printf("%v "+strings.Repeat(escMoveLeft, len(query)+1-pos), string(query[pos:]))
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
			fmt.Printf("%v"+strings.Repeat(" ", pos), string(query[pos:]))
			fmt.Printf(strings.Repeat(escMoveLeft, len(query)))
			query = query[pos:]
			pos = 0
		} else if withQuery && ' ' <= r {
			query = append(query[:pos], append([]rune{r}, query[pos:]...)...)
			fmt.Printf("%v"+strings.Repeat(escMoveLeft, len(query)-pos-1), string(query[pos:]))
			pos++
		}
	}
}
