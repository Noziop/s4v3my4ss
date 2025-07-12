package input

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Noziop/s4v3my4ss/internal/ui/display"
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
		fmt.Printf("%s%s%s\n", display.ColorRed(), errorMessage, display.ColorReset())
	}
}

// ReadStringInput reads and validates a string input, returning the new value or current if empty.
func ReadStringInput(prompt, currentVal string, validator func(string) bool, errorMsg string) string {
	val := ReadAndValidateInput(prompt, validator, errorMsg)
	if val == "" {
		return currentVal
	}
	return val
}

// ReadBoolInput reads a boolean (o/n) input, returning the new value or current if empty.
func ReadBoolInput(prompt string, currentVal bool) bool {
	str := "n"
	if currentVal {
		str = "o"
	}
	inputStr := ReadInput(fmt.Sprintf("%s (actuel: %s, o/n): ", prompt, str))
	if inputStr == "" {
		return currentVal
	}
	return strings.ToLower(inputStr) == "o"
}

// ReadIntInput reads and validates an integer input, returning the new value or current if empty.
func ReadIntInput(prompt string, currentVal int) int {
	inputStr := ReadInput(fmt.Sprintf("%s (actuel: %d): ", prompt, currentVal))
	if inputStr == "" {
		return currentVal
	}
	if i, err := strconv.Atoi(inputStr); err == nil && i >= 0 {
		return i
	}
	common.LogWarning("Valeur numérique invalide: %s. Garde l'actuel.", inputStr)
	return currentVal
}

// ConfirmAction asks for confirmation for an action.
func ConfirmAction(prompt string) bool {
	confirm := ReadInput(prompt + " (o/n): ")
	return strings.ToLower(confirm) == "o"
}

// DisplayMessage shows a success or error message.
func DisplayMessage(isError bool, format string, args ...interface{}) {
	if isError {
		common.LogError(format, args...)
		fmt.Printf("%sErreur: %s%s\n", display.ColorRed(), fmt.Sprintf(format, args...), display.ColorReset())
	} else {
		common.LogInfo(format, args...)
		fmt.Printf("%s%s%s\n", display.ColorGreen(), fmt.Sprintf(format, args...), display.ColorReset())
	}
}
