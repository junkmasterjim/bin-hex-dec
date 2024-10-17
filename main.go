package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type converter struct {
	inputTypeChoices  []string
	outputTypeChoices []string
	cursorPosition    int
	inputType         string
	outputType        string
	input             string
	output            string
	step              int
	error             string
}

func initConverter() converter {
	return converter{
		inputTypeChoices:  []string{"binary", "hexadecimal", "decimal"},
		outputTypeChoices: []string{"binary", "hexadecimal", "decimal"},
		step:              0,
	}
}

func (c converter) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (c converter) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return c, tea.Quit
		case "r":
			if c.step > 0 {
				return initConverter(), nil
			}
		case "up", "k":
			if c.cursorPosition > 0 {
				c.cursorPosition--
			}
		case "down", "j":
			if c.step == 0 && c.cursorPosition < len(c.inputTypeChoices)-1 {
				c.cursorPosition++
			} else if c.step == 1 && c.cursorPosition < len(c.outputTypeChoices)-1 {
				c.cursorPosition++
			}
		case "enter":
			if c.step == 0 {
				c.inputType = c.inputTypeChoices[c.cursorPosition]
				c.step++
				c.cursorPosition = 0
				c.outputTypeChoices = removeChoice(c.outputTypeChoices, c.inputType)
			} else if c.step == 1 {
				c.outputType = c.outputTypeChoices[c.cursorPosition]
				if c.outputType != c.inputType {
					c.step++
				}
			} else if c.step == 2 {
				c.output, c.error = convert(c.input, c.inputType, c.outputType)
				if c.error == "" {
					c.step++
				} else {
					c.input = ""
				}
			} else if c.step == 3 {
				c = initConverter()
			}
		case "backspace":
			if c.step == 2 && len(c.input) > 0 {
				c.input = c.input[:len(c.input)-1]
			}
		default:
			if c.step == 2 {
				c.input += msg.String()
			}
		}
	}
	return c, nil
}

func (c converter) View() string {
	s := ""
	switch c.step {
	case 0:
		s = "What type of input would you like to convert?\n\n"
		for i, choice := range c.inputTypeChoices {
			cursor := " "
			if c.cursorPosition == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	case 1:
		s = "What type of output would you like?\n\n"
		for i, choice := range c.outputTypeChoices {
			cursor := " "
			if c.cursorPosition == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	case 2:
		s = fmt.Sprintf("Enter your %s input:\n", c.inputType)
		activeCursor := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> ")
		s += activeCursor + c.input
		if c.error != "" {
			s += fmt.Sprintf("\n\nError: %s. Press Enter to try again.", c.error)
		}
	case 3:
		s = fmt.Sprintf("Input (%s): %s\nOutput (%s): %s\n", c.inputType, c.input, c.outputType, c.output)
		s += "\nPress Enter to start over."
	}
	s += "\nPress q to quit."
	if c.step > 0 {
		s += "\nPress r to reset."
	}
	s += "\n"

	return s
}

func removeChoice(choices []string, choice string) []string {
	for i, c := range choices {
		if c == choice {
			return append(choices[:i], choices[i+1:]...)
		}
	}
	return choices
}

func convert(input, inputType, outputType string) (string, string) {
	var decimal int64
	var err error

	switch inputType {
	case "binary":
		decimal, err = parseBinary(input)
	case "hexadecimal":
		decimal, err = parseHexadecimal(input)
	case "decimal":
		decimal, err = parseDecimal(input)
	}

	if err != nil {
		return "", "Invalid input"
	}

	switch outputType {
	case "binary":
		return formatBinary(decimal), ""
	case "hexadecimal":
		return formatHexadecimal(decimal), ""
	case "decimal":
		return formatDecimal(decimal), ""
	}

	return "", "Invalid output type"
}

func parseBinary(input string) (int64, error) {
	parts := strings.Split(input, ".")
	result := int64(0)
	for _, part := range parts {
		value, err := strconv.ParseInt(part, 2, 64)
		if err != nil {
			return 0, err
		}
		result = (result << 8) | value
	}
	return result, nil
}

func parseHexadecimal(input string) (int64, error) {
	parts := strings.Split(input, ".")
	result := int64(0)
	for _, part := range parts {
		value, err := strconv.ParseInt(part, 16, 64)
		if err != nil {
			return 0, err
		}
		result = (result << 8) | value
	}
	return result, nil
}

func parseDecimal(input string) (int64, error) {
	parts := strings.Split(input, ".")
	result := int64(0)
	for _, part := range parts {
		value, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return 0, err
		}
		result = (result << 8) | value
	}
	return result, nil
}

func formatBinary(decimal int64) string {
	binary := strconv.FormatInt(decimal, 2)
	numChunks := (len(binary) + 7) / 8
	padded := fmt.Sprintf("%0*s", numChunks*8, binary)
	chunks := splitByN(padded, 8)
	return strings.Join(chunks, ".")
}

func formatHexadecimal(decimal int64) string {
	hex := strconv.FormatInt(decimal, 16)
	numChunks := (len(hex) + 1) / 2
	padded := fmt.Sprintf("%0*s", numChunks*2, hex)
	chunked := strings.Join(splitByN(padded, 2), ".")
	return chunked
}

func formatDecimal(decimal int64) string {
	return strconv.FormatInt(decimal, 10)
}

func splitByN(s string, n int) []string {
	var chunks []string
	runes := []rune(s)

	if len(runes) == 0 {
		return []string{"00000000"}
	}

	for i := 0; i < len(runes); i += n {
		nn := n
		if i+nn > len(runes) {
			nn = len(runes) - i
		}
		chunks = append(chunks, string(runes[i:i+nn]))
	}
	return chunks
}

func printHelp() {
	fmt.Println("Binary-Hex-Decimal Converter")
	fmt.Println("Usage: bhd [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help    Show this help message")
	fmt.Println("\nDescription:")
	fmt.Println("  Allows you to convert between binary, hexadecimal, and decimal numbers.")
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			printHelp()
			return
		}
	}

	p := tea.NewProgram(initConverter(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something went wrong: %v", err)
		os.Exit(1)
	}
}
