# Command line prompter
Command line prompter for terminal user interaction with the following features:

- user input into `string`, `bool`, `int`, `float64`, ...
- yes/no prompt
- select prompt with options

*See also [github.com/tdewolff/argp](https://github.com/tdewolff/argp) for a command line argument parser.*

## Installation
Make sure you have [Git](https://git-scm.com/) and [Go](https://golang.org/dl/) (1.13 or higher) installed, run
```
mkdir Project
cd Project
go mod init
go get -u github.com/tdewolff/prompt
```

Then add the following import
``` go
import (
    "github.com/tdewolff/prompt"
)
```

## Examples
### Input prompt
A regular prompt requesting user input. When the target is a primary type (except boolean) or implements the `Stringer` interface, it will be editable in-place.

```go
package main

import "github.com/tdewolff/prompt"

func main() {
    // Validators verify the user input to match conditions.
    validators := []prompt.Validator{prompt.StrLength(5, 10), prompt.Suffix("suffix")}

    var val string
    deflt := prompt.Default("value", 3)  // set text caret to the 3rd character
    if err := prompt.Prompt(&val, "Label", deflt, validators...); err != nil {
        panic(err)
    }
    fmt.Println("Result:", val)
}
```

where `val` can be of any primary type, such as `string`, `bool`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, or `float64`.

When the value is editable it allowd users to use keys such as: <kbd>Left</kbd>, <kbd>Ctrl</kbd> + <kbd>B</kbd> to move left; <kbd>Right</kbd>, <kbd>Ctrl</kbd> + <kbd>F</kbd> to move right; <kbd>Home</kbd>, <kbd>Ctrl</kbd> + <kbd>A</kbd> to go to start; <kbd>End</kbd>, <kbd>Ctrl</kbd> + <kbd>E</kbd> to go to end; <kbd>Backspace</kbd> and <kbd>Delete</kbd> to delete a character; <kbd>Ctrl</kbd> + <kbd>K</kbd> and <kbd>Ctrl</kbd> + <kbd>U</kbd> to delete from the caret to the start and end of the input respectively; <kbd>Enter</kbd>, <kbd>Ctrl</kbd> + <kbd>D</kbd> to confirm input; and <kbd>Ctrl</kbd> + <kbd>C</kbd>, <kbd>Esc</kbd> to quit.

### Select prompt
A select prompt that allows the user to select amongst predetermined option.

```go
package main

import "github.com/tdewolff/prompt"

func main() {
    val, deflt := "", "Yellow"  // can be ints if you need the index into options
    options := []string{"Red", "Orange", "Green", "Yellow", "Blue", "Purple"}
    if err := prompt.Select(&val, "Label", options, deflt); err != nil {
        panic(err)
    }
    fmt.Println("Selected:", val)
}
```

The select prompt allows users to use keys such as: <kbd>Up</kbd>, <kbd>W</kbd>, <kbd>K</kbd> to go up; <kbd>Down</kbd>, <kbd>S</kbd>, <kbd>J</kbd> to go down; <kbd>Tab</kbd> and <kbd>Shift</kbd> + <kbd>Tab</kbd> to go down or up respectively with wrapping around; <kbd>Home</kbd> to go to first; <kbd>End</kbd> to go to last; <kbd>Enter</kbd>, <kbd>Ctrl</kbd> + <kbd>D</kbd> to select option; and <kbd>Ctrl</kbd> + <kbd>C</kbd>, <kbd>Esc</kbd> to quit.

### Yes/No prompt
A yes or no prompt.

```go
package main

import "github.com/tdewolff/prompt"

func main() {
    val := prompt.YesNo("Label", false)
    fmt.Println("Value:", val)
}
```

`true` is any of `1`, `y`, `yes`, `t`, `true` and is case-insensitive.

`false` is any of `0`, `n`, `no`, `f`, `false` and is case-insensitive.

### Enter prompt
A prompt that waits for Enter to be pressed.

```go
package main

import "github.com/tdewolff/prompt"

func main() {
    prompt.Enter("Are you done?")  // waits for enter
}
```

### Validators
```go
Not(Validator)     // logical NOT
And(Validator...)  // logical AND
Or(Validator...)   // logical OR

In([]any)     // in list
NotIn([]any)  // not in list

StrLength(min, max int)           // limit string length (inclusive)
NumRange(min, max float64)        // limit int/uint/float range (inclusive)
DateRange(min, max time.Time)     // limit time.Time range (inclusive)
Prefix(afix string)
Suffix(afix string)
Pattern(pattern, message string)  // pattern match and error message
EmailAddress()
IPAddress()                       // valid IPv4 or IPv6 address
IPv4Address()
IPv6Address()
Port()                            // server port
Path()                            // Unix path
AbsolutePath()                    // Unix absolute path
UserName()                        // valid Unix user name
TopDomainName()                   // such as example.com
DomainName()                      // such as sub.example.com
FQDN()                            // such as sub.example.com.
Dir()                             // existing directory
File()                            // existing file
```

## License
Released under the [MIT license](LICENSE.md).
