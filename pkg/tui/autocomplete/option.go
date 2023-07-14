package autocomplete

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type Option struct {
	Name        string
	Description string
}

func parseOptions(output string) (cobra.ShellCompDirective, []Option, error) {

	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return 0, nil, nil
	}

	var options = make([]Option, 0, len(lines)-1)

	directive := cobra.ShellCompDirectiveDefault

	// Loop over all the lines apart from the last two
	for _, line := range lines[:len(lines)-1] {
		if strings.HasPrefix(line, ":") {
			directiveLine := lines[len(lines)-2]
			if strings.HasPrefix(directiveLine, ":") {
				directiveLine = directiveLine[1:]

				value, err := strconv.ParseUint(strings.TrimSpace(directiveLine), 10, 64)
				if err != nil {
					directive = cobra.ShellCompDirective(value)
				}
			}
			break
		}

		cmd, description, _ := strings.Cut(line, "\t")

		if isShorthandFlag(cmd) {
			continue
		}

		options = append(options, Option{
			Name:        escapeSpecialCharacters(cmd),
			Description: description,
		})
	}

	sort.Slice(options, func(i, j int) bool {
		return options[i].Name < options[j].Name
	})

	return directive, options, nil
}

func escapeSpecialCharacters(val string) string {
	for _, c := range []string{"\\", "\"", "$", "`", "!"} {
		val = strings.ReplaceAll(val, c, "\\"+c)
	}

	if strings.ContainsAny(val, " #&*;<>?[]|~") {
		val = fmt.Sprintf(`"%s"`, val)
	}

	return val
}

func isFlag(arg string) bool {
	return strings.HasPrefix(arg, "-")
}

func isShorthandFlag(arg string) bool {
	return isFlag(arg) && !strings.HasPrefix(arg, "--")
}
