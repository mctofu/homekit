package gen

import "strings"

func pascalCase(name string) string {
	return strings.ReplaceAll(name, " ", "")
}
