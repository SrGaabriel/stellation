package cli

import (
	"os"

	"github.com/pterm/pterm"
)

func Successf(msg string, args ...any) {
	pterm.Success.WithPrefix(pterm.Prefix{
		Text:  "DONE",
		Style: pterm.NewStyle(pterm.BgGreen, pterm.FgBlack),
	}).WithMessageStyle(pterm.NewStyle(pterm.FgWhite)).Printfln(msg, args...)
}

func Errorf(format string, args ...any) {
	pterm.Error.WithPrefix(pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgRed, pterm.FgBlack),
	}).WithMessageStyle(pterm.NewStyle(pterm.FgWhite)).WithWriter(os.Stderr).Printfln(format, args...)
}
