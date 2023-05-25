package main

import (
	"fmt"

	"github.com/tdewolff/prompt"
	"golang.org/x/text/language"
)

type Language language.Tag

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

	if s == "" {
		return fmt.Errorf("expected language tag")
	}
	tag, err := language.Parse(s)
	if err != nil {
		return err
	}
	*lang = Language(tag)
	return nil
}

func (lang Language) String() string {
	return language.Tag(lang).String()
}

func main() {
	var name string
	var age uint
	var language Language
	var smoker bool
	var car string
	if err := prompt.Prompt(&name, "Name", prompt.Default("Juan", 2), prompt.StrLength(3, -1)); err != nil {
		panic(err)
	}
	if err := prompt.Prompt(&age, "Age (18-65)", nil, prompt.NumRange(18, 65)); err != nil {
		panic(err)
	}
	if err := prompt.Prompt(&language, "Language", language); err != nil {
		panic(err)
	}
	if err := prompt.Prompt(&smoker, "Smoker", nil); err != nil {
		panic(err)
	}
	cars := []string{"Chevrolet", "Kia", "Peugeot", "Subaru", "Volvo"}
	if err := prompt.Select(&car, "Car brand", cars, "Subaru"); err != nil {
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
