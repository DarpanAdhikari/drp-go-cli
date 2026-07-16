package generator

import (
	"strings"
	"unicode"
)

// Names holds all the derived name variants for a resource, pre-computed
// so templates don't need to call functions themselves.
type Names struct {
	// e.g. input "product_category"
	Name       string // PascalCase:   "ProductCategory"
	LowerName  string // lowercase:    "productcategory"
	PluralName string // naive plural: "ProductCategories"
	TableName  string // snake_case:   "product_categories"
	RouteName  string // kebab-case:   "product-categories"
	DomainName string // snake_case singular: "product_category"
	ModuleName string // from go.mod:  "github.com/yourorg/myapp"
	DBDriver   string // "postgres" or "mysql"
}

// NewNames derives all name variants from a raw resource name (e.g. "product"
// or "product_category") and the project's Go module name.
func NewNames(raw, moduleName string) Names {
	return NewNamesWithDriver(raw, moduleName, "postgres")
}

// NewNamesWithDriver is like NewNames but also accepts a DB driver name.
func NewNamesWithDriver(raw, moduleName, dbDriver string) Names {
	snake := toSnake(raw)
	pascal := toPascal(snake)
	plural := naivePlural(pascal)
	tableSnake := toSnake(naivePluralSnake(snake))

	if dbDriver == "" {
		dbDriver = "postgres"
	}

	return Names{
		Name:       pascal,
		LowerName:  strings.ToLower(pascal),
		PluralName: plural,
		TableName:  tableSnake,
		RouteName:  strings.ReplaceAll(tableSnake, "_", "-"),
		DomainName: snake,
		ModuleName: moduleName,
		DBDriver:   dbDriver,
	}
}

// toSnake converts a raw name to snake_case (spaces and hyphens → underscores,
// CamelCase → snake_case).
func toSnake(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ToLower(s)

	// Handle CamelCase input: insert underscores before uppercase runs.
	var out []rune
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(r))
	}
	return strings.Trim(string(out), "_")
}

// toPascal converts a snake_case string to PascalCase.
func toPascal(snake string) string {
	parts := strings.Split(snake, "_")
	var sb strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		sb.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return sb.String()
}

// naivePlural returns a naive plural of a PascalCase name.
// Handles: -y → -ies, -s/-x/-z/-ch/-sh → -es, otherwise append -s.
func naivePlural(pascal string) string {
	l := strings.ToLower(pascal)
	switch {
	case strings.HasSuffix(l, "y") && len(l) > 1 && !isVowel(rune(l[len(l)-2])):
		return pascal[:len(pascal)-1] + "ies"
	case strings.HasSuffix(l, "s") || strings.HasSuffix(l, "x") ||
		strings.HasSuffix(l, "z") || strings.HasSuffix(l, "ch") ||
		strings.HasSuffix(l, "sh"):
		return pascal + "es"
	default:
		return pascal + "s"
	}
}

// naivePluralSnake applies naive pluralisation to a snake_case word's last segment.
func naivePluralSnake(snake string) string {
	parts := strings.Split(snake, "_")
	last := parts[len(parts)-1]
	parts[len(parts)-1] = strings.ToLower(naivePlural(toPascal(last)))
	return strings.Join(parts, "_")
}

func isVowel(r rune) bool {
	return strings.ContainsRune("aeiou", unicode.ToLower(r))
}
