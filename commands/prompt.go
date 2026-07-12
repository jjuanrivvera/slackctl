package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// promptLine prints label to stderr and reads one line from stdin (trimmed). It reads one
// byte at a time rather than buffering, so successive prompts on the same reader don't lose
// input that a buffered reader would have read ahead and discarded.
func promptLine(cmd *cobra.Command, label string) (string, error) {
	fmt.Fprint(cmd.ErrOrStderr(), label)
	r := cmd.InOrStdin()
	var b strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				break
			}
			b.WriteByte(buf[0])
		}
		if err != nil {
			if b.Len() == 0 {
				return "", err
			}
			break
		}
	}
	return strings.TrimSpace(b.String()), nil
}

// promptSecret reads a secret without echoing when stdin is a terminal; on a pipe it falls
// back to a normal line read so scripts still work.
func promptSecret(cmd *cobra.Command, label string) (string, error) {
	fmt.Fprint(cmd.ErrOrStderr(), label)
	if f, ok := cmd.InOrStdin().(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", err
		}
		return sanitizeSecret(string(b)), nil
	}
	return promptLine(cmd, "")
}

// sanitizeSecret strips terminal bracketed-paste markers (ESC[200~ … ESC[201~) and trims
// surrounding whitespace. With bracketed paste enabled, a raw read (unlike the shell's line
// editor) receives those wrappers around pasted text; left in they corrupt a pasted key so it
// fails auth. Stripping them fixes the common "typing works, pasting fails".
func sanitizeSecret(s string) string {
	s = strings.ReplaceAll(s, "\x1b[200~", "")
	s = strings.ReplaceAll(s, "\x1b[201~", "")
	return strings.TrimSpace(s)
}
