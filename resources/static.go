package resource

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	Reset        = "\033[0m"
	Bold         = "\033[1m"
	BlueBright   = "\033[94m"
	CyanBright   = "\033[96m"
	YellowBright = "\033[93m"
	WhiteBright  = "\033[97m"
	DimWhite     = "\033[37m"
)

func PrintBanner() {
	enableANSI()

	lines := []string{
		` █████╗  ██████╗ ███████╗███╗   ██╗████████╗    ██████╗  ██████╗  ██╗ █████╗ `,
		`██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝   ██╔═████╗██╔═████╗███║██╔══██╗`,
		`███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║      ██║██╔██║██║██╔██║╚██║╚█████╔╝`,
		`██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║      ████╔╝██║████╔╝██║ ██║██╔══██╗`,
		`██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║      ╚██████╔╝╚██████╔╝ ██║╚█████╔╝`,
		`╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝       ╚═════╝  ╚═════╝ ╚═╝ ╚════╝ `,
	}

	maxLen := 0
	for _, line := range lines {
		if len([]rune(line)) > maxLen {
			maxLen = len([]rune(line))
		}
	}

	width := maxLen + 4
	border := strings.Repeat("═", width)

	fmt.Println()
	fmt.Printf("%s╔%s╗%s\n", BlueBright, border, Reset)
	fmt.Printf("%s║%s%s║%s\n", BlueBright, strings.Repeat(" ", width), BlueBright, Reset)

	for _, line := range lines {
		runes := []rune(line)
		padCount := width - len(runes)
		if padCount < 0 {
			padCount = 0
		}
		leftPad := padCount / 2
		rightPad := padCount - leftPad
		fmt.Printf("%s║%s%s%s%s%s║%s\n",
			BlueBright,
			strings.Repeat(" ", leftPad),
			BlueBright+Bold,
			string(runes),
			strings.Repeat(" ", rightPad),
			BlueBright,
			Reset,
		)
	}

	fmt.Printf("%s║%s%s║%s\n", BlueBright, strings.Repeat(" ", width), BlueBright, Reset)
	fmt.Printf("%s╚%s╝%s\n", BlueBright, border, Reset)
	fmt.Println()
}

func centerText(text string, width int) string {
	runes := []rune(text)
	textLen := len(runes)
	if textLen >= width {
		return text
	}
	totalPad := width - textLen
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	return fmt.Sprintf("%s%s%s%s%s",
		strings.Repeat(" ", leftPad),
		BlueBright+Bold,
		text,
		Reset,
		strings.Repeat(" ", rightPad),
	)
}

func YouPrompt() {
	fmt.Printf("%s%s  You › %s", Bold, YellowBright, Reset)
}

func AgentPrompt() {
	fmt.Printf("\n%s%s AGENT 0018 › %s", Bold, CyanBright, Reset)
}

func GoodBye() {
	fmt.Printf("\n%s  Goodbye. AGENT 0018 signing off.%s\n\n", DimWhite, Reset)
}

func ErrorPrompt(err error) {
	fmt.Printf("\n\033[91m✗ Error: %v\033[0m\n\n", err)
}

func Separator() {
	width := 95
	fmt.Printf("%s%s%s\n", DimWhite, strings.Repeat("─", width), Reset)
}

func enableANSI() {
	cmd := exec.Command("cmd", "/c", "color")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
