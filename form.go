package prompt

import (
	"fmt"
	"strings"
)

type Form struct {
	labels []string
	inputs []func() error
}

func NewForm() *Form {
	return &Form{}
}

func (f *Form) Print(label string, ival interface{}) {
	i := len(f.labels)
	f.labels = append(f.labels, label)
	f.inputs = append(f.inputs, func() error {
		fmt.Printf("%v: %v\n", f.labels[i], ival)
		return nil
	})
}

func (f *Form) Prompt(idst interface{}, label string, validators ...Validator) {
	i := len(f.labels)
	f.labels = append(f.labels, label)
	f.inputs = append(f.inputs, func() error {
		return Prompt(idst, f.labels[i], validators...)
	})
}

func (f *Form) Select(idst interface{}, label string, ioptions interface{}) {
	i := len(f.labels)
	f.labels = append(f.labels, label)
	f.inputs = append(f.inputs, func() error {
		return Select(idst, f.labels[i], ioptions)
	})
}

func (f *Form) Send() error {
	n := 0
	for _, label := range f.labels {
		if n < len(label) {
			n = len(label)
		}
	}
	for i, label := range f.labels {
		if len(label) < n {
			f.labels[i] = strings.Repeat(" ", n-len(label)) + label
		}
	}
	for _, input := range f.inputs {
		if err := input(); err != nil {
			return err
		}
	}
	return nil
}
