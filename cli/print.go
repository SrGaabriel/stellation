package cli

import "github.com/pterm/pterm"

func Success(msg string) {
	pterm.Success.WithPrefix(pterm.Prefix{
		Text:  "DONE",
		Style: pterm.NewStyle(pterm.BgGreen, pterm.FgBlack),
	}).WithMessageStyle(pterm.NewStyle(pterm.FgWhite)).Println(msg)
}
