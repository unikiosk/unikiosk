package shell

import (
	"bytes"
	"os/exec"
)

const ShellToUse = "sh"

// Exec executes shell command provided using specific shell
// returns: stdout, stderr, err
func Exec(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}
