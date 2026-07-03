package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type CRUDSelections struct {
	Model      bool
	Repository bool
	Service    bool
	Handler    bool
	Routes     bool
}

type layerOption struct {
	key  string
	path string
}

var layers = []layerOption{
	{"Model", "internal/models"},
	{"Repository", "internal/repositories"},
	{"Service", "internal/services"},
	{"Handler", "internal/handlers"},
	{"Routes", "internal/routes"},
}

func CRUDLayerSelection(resourceName string) (CRUDSelections, error) {
	fmt.Println()
	fmt.Printf("Generate CRUD for %q — select layers by number:\n", resourceName)
	fmt.Println()
	for i, l := range layers {
		fmt.Printf("  [%d] %s\n", i+1, l.key)
	}
	fmt.Println()
	fmt.Print("Enter numbers (comma/space separated, e.g. 1,2,3 or 1-5): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()
	if err := scanner.Err(); err != nil {
		return CRUDSelections{}, fmt.Errorf("read input: %w", err)
	}

	selected := parseRange(input)

	sel := CRUDSelections{}
	for _, n := range selected {
		switch n {
		case 1:
			sel.Model = true
		case 2:
			sel.Repository = true
		case 3:
			sel.Service = true
		case 4:
			sel.Handler = true
		case 5:
			sel.Routes = true
		}
	}

	if !sel.Model && !sel.Repository && !sel.Service && !sel.Handler && !sel.Routes {
		fmt.Fprintln(os.Stderr, "No layers selected. Aborting.")
		os.Exit(2)
	}

	return sel, nil
}

func PromptModule() (string, error) {
	fmt.Print("Go module name (e.g. github.com/yourorg/myapp): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	module := strings.TrimSpace(scanner.Text())
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}

	if module == "" {
		return "", fmt.Errorf("module name cannot be empty")
	}

	return module, nil
}

func ConfirmGeneration(files []string) (bool, error) {
	fmt.Println()
	fmt.Println("The following files will be created:")
	for _, f := range files {
		fmt.Printf("  \u2713 %s\n", f)
	}
	fmt.Println()

	fmt.Print("Proceed with generation? (Y/n): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.TrimSpace(scanner.Text())
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("read input: %w", err)
	}

	return response == "" || strings.EqualFold(response, "y") || strings.EqualFold(response, "yes"), nil
}

func parseRange(input string) []int {
	input = strings.ReplaceAll(input, ",", " ")
	fields := strings.Fields(input)

	var result []int
	seen := make(map[int]bool)

	for _, f := range fields {
		if strings.Contains(f, "-") {
			parts := strings.SplitN(f, "-", 2)
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil && start <= end {
				for i := start; i <= end; i++ {
					if i >= 1 && i <= len(layers) && !seen[i] {
						result = append(result, i)
						seen[i] = true
					}
				}
			}
		} else {
			n, err := strconv.Atoi(f)
			if err == nil && n >= 1 && n <= len(layers) && !seen[n] {
				result = append(result, n)
				seen[n] = true
			}
		}
	}

	return result
}
