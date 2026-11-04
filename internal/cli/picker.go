// Package cli contains the interactive terminal prompts of the lamp
// generator: the generator and design mode pickers (up/down keys) and the
// seed prompt.
package cli

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"lamp3d-cli/internal/design"
	"lamp3d-cli/internal/fluid"
)

// generatorOptions are the entries of the generator picker and the
// generator names they resolve to.
var generatorOptions = []struct{ label, name string }{
	{"any (random)", "any"},
	{"classic (wave lamp)", "classic"},
	{"fluid (flowing patterns)", "fluid"},
}

// SelectGenerator shows an interactive picker for the lamp generator. The
// list is navigated with the up/down arrow keys and confirmed with enter;
// 'q' or Ctrl+C quits the program. When stdin is not a terminal (e.g. piped
// input) "any" is returned directly.
func SelectGenerator() string {
	options := make([]string, len(generatorOptions))
	for i, o := range generatorOptions {
		options[i] = o.label
	}
	idx := selectOption("Select the generator (up/down, enter to confirm):", options)
	return generatorOptions[idx].name
}

// classicModeHints are the short explanations shown in parentheses next to
// each classic mode in the picker.
var classicModeHints = map[design.DesignMode]string{
	design.ModeAny:    "random pick",
	design.ModeWave:   "fine spiral ribs",
	design.ModeFluted: "deep twisted flutes",
	design.ModeSpiral: "bold helical ribs",
	design.ModeRipple: "horizontal rings",
	design.ModeDune:   "wavy vertical lobes",
}

// fluidModeHints are the short explanations shown in parentheses next to
// each fluid mode in the picker.
var fluidModeHints = map[design.DesignMode]string{
	design.ModeAny:   "random pick",
	fluid.ModeFlow:   "drifting harmonics",
	fluid.ModeBloom:  "breathing petals",
	fluid.ModeVortex: "slow swirl + ring",
	fluid.ModeWeave:  "flowing moiré",
	fluid.ModeRipple: "smooth helical rings",
}

// SelectMode shows an interactive picker for the classic design mode. The
// list is navigated with the up/down arrow keys and confirmed with enter;
// 'q' or Ctrl+C quits the program. When stdin is not a terminal (e.g. piped
// input) ModeAny is returned directly.
func SelectMode() design.DesignMode {
	modes := append([]design.DesignMode{design.ModeAny}, design.Modes()...)
	return modes[selectOption("Select the design mode (up/down, enter to confirm):", modeLabels(modes, classicModeHints))]
}

// SelectFluidMode shows an interactive picker for the fluid design modes.
// It behaves exactly like SelectMode but lists the fluid generator's
// styles.
func SelectFluidMode() design.DesignMode {
	modes := append([]design.DesignMode{design.ModeAny}, fluid.Modes()...)
	return modes[selectOption("Select the fluid design mode (up/down, enter to confirm):", modeLabels(modes, fluidModeHints))]
}

// modeLabels converts design modes into the strings shown in the picker:
// the mode name followed by a short explanation in parentheses.
func modeLabels(modes []design.DesignMode, hints map[design.DesignMode]string) []string {
	labels := make([]string, len(modes))
	for i, m := range modes {
		if hint, ok := hints[m]; ok {
			labels[i] = fmt.Sprintf("%s (%s)", m, hint)
		} else {
			labels[i] = string(m)
		}
	}
	return labels
}

// selectOption shows a generic interactive list picker and returns the
// index of the chosen entry. The list is navigated with the up/down arrow
// keys and confirmed with enter; 'q' quits by accepting the first entry and
// Ctrl+C aborts the program. When stdin is not a terminal (e.g. piped
// input) index 0 is returned directly.
func selectOption(title string, options []string) int {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return 0
	}

	idx := 0

	// Switch the terminal to raw mode so single key presses arrive
	// immediately, and restore it on the way out.
	old, err := term.MakeRaw(fd)
	if err != nil {
		return 0
	}
	defer term.Restore(fd, old)

	// Abort restores the terminal and exits like a Ctrl+C in a normal shell.
	abort := func() {
		term.Restore(fd, old)
		fmt.Print("\r\n")
		os.Exit(130)
	}

	fmt.Print("\x1b[?25l") // hide cursor
	defer fmt.Print("\x1b[?25h\r\n")

	first := true
	draw := func() {
		if !first {
			// Move the cursor back up to the header line: one line per
			// menu entry plus the header itself.
			fmt.Printf("\x1b[%dA\r", len(options)+1)
		}
		first = false
		fmt.Printf("%s\r\n", title)
		for i, o := range options {
			if i == idx {
				fmt.Printf("  \x1b[7m%s\x1b[0m\x1b[K\r\n", o)
			} else {
				fmt.Printf("    %s\x1b[K\r\n", o)
			}
		}
	}

	draw()

	// Read one byte at a time and parse escape sequences manually, so keys
	// still work if the terminal splits a sequence across reads.
	var b [1]byte
	for {
		if _, err := os.Stdin.Read(b[:]); err != nil || b[0] == 4 {
			return 0
		}
		switch b[0] {
		case 3: // Ctrl+C
			abort()
		case '\r', '\n':
			return idx
		case 'q', 'Q':
			return 0
		case 27: // escape sequence (e.g. arrow keys)
			if _, err := os.Stdin.Read(b[:]); err != nil {
				return 0
			}
			if b[0] != '[' && b[0] != 'O' {
				continue // unknown sequence, ignore
			}
			if _, err := os.Stdin.Read(b[:]); err != nil {
				return 0
			}
			switch b[0] {
			case 'A': // up
				if idx > 0 {
					idx--
					draw()
				}
			case 'B': // down
				if idx < len(options)-1 {
					idx++
					draw()
				}
			}
		}
	}
}

// ResolveSeed returns the seed to use. A nonzero flag value wins. Otherwise
// the user is asked on the terminal; pressing enter there picks a random
// seed from the current time.
func ResolveSeed(flagValue int64) int64 {
	if flagValue != 0 {
		return flagValue
	}
	fmt.Print("Enter a seed (press enter for a random one): ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err == nil {
		line = strings.TrimSpace(line)
		if line != "" {
			if v, err := strconv.ParseInt(line, 10, 64); err == nil {
				return v
			}
			log.Fatalf("Invalid seed %q: must be an integer", line)
		}
	}
	return time.Now().UnixNano()
}
