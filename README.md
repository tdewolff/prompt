# Command line prompter
Command line prompter for terminal user interaction.

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
### Regular prompt
A regular prompt requesting user input.

```go
package main

import "github.com/tdewolff/prompt"

func main() {
    // Validators verify the user input to match conditions.
    validators := []prompt.Validator{prompt.StrLength(5,10), prompt.Suffix("suffix")}

    var val string
    if err := prompt.Prompt(&val, "Label", "default value", validators...); err != nil {
        panic(err)
    }
    fmt.Println("Value:", val)
}
```

where `val` can be of any primary type, such as `string`, `bool`, `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `float32`, or `float64`.

### Select prompt
A select prompt that allows the user to select amongst predetermined option.

```go
package main

import "github.com/tdewolff/prompt"

func main() {
    var val string // can be int if you need the index into options
    deflt := "Yellow" // can be int if you need the index into options
    options := []string{"Red", "Orange", "Green", "Yellow", "Blue", "Purple"}
    if err := prompt.Select(&val, "Label", options, deflt); err != nil {
        panic(err)
    }
    fmt.Println("Value:", val)
}
```

The select prompt allows users to use keys such as: <kbd>Up</kbd>, <kbd>W</kbd>, <kbd>K</kbd> to go up; <kbd>Down</kbd>, <kbd>S</kbd>, <kbd>J</kbd> to go down; <kbd>Tab</kbd> to go down with wrapping to first; <kbd>Home</kbd> to go to first; <kbd>End</kbd> to go to last; <kbd>Enter</kbd>, <kbd>Ctrl</kbd> + <kbd>D</kbd> to select option; and <kbd>Ctrl</kbd> + <kbd>C</kbd> to interrupt.

### Yes/no prompt
A yes or no prompt.

```go
package main

import "github.com/tdewolff/prompt"

func main() {
    val := prompt.YesNo("Label", false)
    fmt.Println("Value:", val)
}
```

`true` is any of `1`, `y`, `Y`, `yes`, `YES`, `t`, `T`, `true`, `TRUE`.

`false` is any of `0`, `n`, `N`, `no`, `NO`, `f`, `F`, `false`, `FALSE`.

### Validators
```go
StrLength(min, max int)           // limit string length (-1 is no limit)
IntRange(min, max int64)          // limit int64/uint64 range
FloatRange(min, max float64)      // limit float32/float64 range
Prefix(afix string)
Suffix(afix string)
Pattern(pattern, message string)  // pattern match and error message
EmailAddress()
IPAddress()
IPv4Address()
IPv6Address()
Port()                            // server port
Path()                            // Linux path
AbsPath()                         // Linux absolute path
UserName()
TopDomainName()                   // such as example.com
DomainName()                      // such as sub.example.com
StrNot(items []string)            // exclude items as valid input
Dir()                             // existing directory
File()                            // existing file
```

## License
Released under the [MIT license](LICENSE.md).
