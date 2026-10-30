// Package cli contains the interactive terminal prompts of the lamp
// generator: the design mode picker (up/down keys) and the seed prompt.
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

	"lamp-3d/internal/design"
)

// SelectMode shows an interactive picker for the design mode. The list is
// navigated with the up/down arrow keys and confirmed with enter; 'q' or
// Ctrl+C quits the program. When stdin is not a terminal (e.g. piped input)
// ModeAny is returned directly.
func SelectMode() design.DesignMode {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return design.ModeAny
	}

	modes := append([]design.DesignMode{design.ModeAny}, design.Modes()...)
	idx := 0

	// Switch the terminal to raw mode so single key presses arrive
	// immediately, and restore it on the way out.
	old, err := term.MakeRaw(fd)
	if err != nil {
		return design.ModeAny
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
			fmt.Printf("\x1b[%dA\r", len(modes)+1)
		}
		first = false
		fmt.Printf("Select the design mode (up/down, enter to confirm):\r\n")
		for i, m := range modes {
			if i == idx {
				fmt.Printf("  \x1b[7m%s\x1b[0m\x1b[K\r\n", m)
			} else {
				fmt.Printf("    %s\x1b[K\r\n", m)
			}
		}
	}

	draw()

	// Read one byte at a time and parse escape sequences manually, so keys
	// still work if the terminal splits a sequence across reads.
	var b [1]byte
	for {
		if _, err := os.Stdin.Read(b[:]); err != nil || b[0] == 4 {
			return design.ModeAny
		}
		switch b[0] {
		case 3: // Ctrl+C
			abort()
		case '\r', '\n':
			return modes[idx]
		case 'q', 'Q':
			return design.ModeAny
		case 27: // escape sequence (e.g. arrow keys)
			if _, err := os.Stdin.Read(b[:]); err != nil {
				return design.ModeAny
			}
			if b[0] != '[' && b[0] != 'O' {
				continue // unknown sequence, ignore
			}
			if _, err := os.Stdin.Read(b[:]); err != nil {
				return design.ModeAny
			}
			switch b[0] {
			case 'A': // up
				if idx > 0 {
					idx--
					draw()
				}
			case 'B': // down
				if idx < len(modes)-1 {
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
