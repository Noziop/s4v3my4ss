package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Noziop/s4v3my4ss/pkg/common"
)

var reader *bufio.Reader

func init() {
	reader = bufio.NewReader(os.Stdin)
}

// ReadInput lit une entrée utilisateur depuis la console.
func ReadInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// ReadAndValidateInput lit une entrée utilisateur et la valide en utilisant une fonction de validation.
// La lecture se répète jusqu'à ce que l'entrée soit valide.
func ReadAndValidateInput(prompt string, validator func(string) bool, errorMessage string) string {
	for {
		input := ReadInput(prompt)
		if validator(input) {
			return input
		}
		fmt.Printf("%s%s%s\n", common.ColorRed, errorMessage, common.ColorReset)
	}
}
