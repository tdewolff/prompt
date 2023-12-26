package main

import (
	"fmt"
	"strings"

	"github.com/tdewolff/prompt"
)

type Language int

const (
	English Language = iota
	French
	Dutch
)

func (lang *Language) Scan(i interface{}) error {
	s := ""
	switch v := i.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("incompatible type for Language: %T", i)
	}

	switch strings.ToLower(s) {
	case "en":
		*lang = English
	case "fr":
		*lang = French
	case "nl":
		*lang = Dutch
	default:
		return fmt.Errorf("expected language")
	}
	return nil
}

func (lang Language) String() string {
	switch lang {
	case English:
		return "en"
	case French:
		return "fr"
	case Dutch:
		return "nl"
	}
	return ""
}

func main() {
	var age uint
	var language Language
	var smoker bool
	smokerBrands := []string{"Camel"}
	name := "Juan"
	car := "Subaru"

	if err := prompt.Prompt(prompt.Default(&name, name, 2), "Name", prompt.StrLength(3, -1)); err != nil {
		panic(err)
	}
	if err := prompt.Prompt(&age, "Age (18-65)", prompt.NumRange(18, 65)); err != nil {
		panic(err)
	}
	if err := prompt.Prompt(&language, "Language"); err != nil {
		panic(err)
	}
	if err := prompt.Prompt(&smoker, "Smoker"); err != nil {
		panic(err)
	}
	if smoker {
		brands := []string{"Marlboro", "Newport", "Camel", "Pall Mall"}
		if err := prompt.Checklist(&smokerBrands, "Cigarette brands", brands); err != nil {
			panic(err)
		}
	}
	cars := []string{"Chevrolet", "Kia", "Peugeot", "Subaru", "Volvo"}
	if err := prompt.Select(&car, "Car brand", cars); err != nil {
		panic(err)
	}
	smokerMsg := ""
	if !smoker {
		smokerMsg = "not "
	}
	fmt.Printf("\nYou are %v, %v years old, speak %v, %va smoker, and you drive a %v.\n", name, age, language, smokerMsg, car)
	if prompt.YesNo("Is that correct?", false) {
		fmt.Println("Done")
	} else {
		fmt.Println("Aborted")
	}
}
