package main

import (
	"fmt"

	"github.com/tdewolff/prompt"
)

func main() {
	var name, car string
	var age uint
	var smoker bool
	if err := prompt.Prompt(&name, "Name", "Juan"); err != nil {
		panic(err)
	}
	if err := prompt.Prompt(&age, "Age (18-65)", nil, prompt.NumRange(18, 65)); err != nil {
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
	fmt.Printf("\nYou are %v, %v years old, %va smoker, and you drive a %v.\n", name, age, smokerMsg, car)
	if prompt.YesNo("Is that correct?", false) {
		fmt.Println("Done")
	} else {
		fmt.Println("Aborted")
	}
}
