package generator

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DetectModuleName reads the module name from a go.mod file in the current
// working directory (or the path given). Returns an error if go.mod is not
// found or has no module declaration — callers should surface this clearly.
func DetectModuleName(gomodPath string) (string, error) {
	if gomodPath == "" {
		gomodPath = "go.mod"
	}
	f, err := os.Open(gomodPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(
				"go.mod not found at %q — run `drp init` first or specify --module manually",
				gomodPath,
			)
		}
		return "", fmt.Errorf("reading go.mod: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			if mod == "" {
				return "", fmt.Errorf("go.mod has empty module declaration")
			}
			return mod, nil
		}
	}
	return "", fmt.Errorf("go.mod has no module declaration")
}
