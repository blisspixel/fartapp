package cli

import (
	"fmt"
	"strings"
)

// parseValuedOptions extends presentation options with named, single-use
// values. Values belonging to --format are never reinterpreted as options.
func parseValuedOptions(args []string, names ...string) (outputOptions, map[string]string, error) {
	values := make(map[string]string, len(names))
	filtered := make([]string, 0, len(args))
	ended := false
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if ended || arg == "--" {
			ended = true
			filtered = append(filtered, arg)
			continue
		}
		if arg == "--format" {
			filtered = append(filtered, arg)
			if index+1 < len(args) {
				index++
				filtered = append(filtered, args[index])
			}
			continue
		}
		name, value, inline := strings.Cut(arg, "=")
		recognized := false
		for _, candidate := range names {
			if name == candidate {
				recognized = true
				break
			}
		}
		if !recognized {
			filtered = append(filtered, arg)
			continue
		}
		if _, exists := values[name]; exists {
			return outputOptions{}, nil, fmt.Errorf("%s may be specified only once", name)
		}
		if !inline {
			index++
			if index == len(args) || strings.HasPrefix(args[index], "--") || args[index] == "-h" {
				return outputOptions{}, nil, fmt.Errorf("%s requires a value", name)
			}
			value = args[index]
		}
		if value == "" {
			return outputOptions{}, nil, fmt.Errorf("%s requires a nonempty value", name)
		}
		values[name] = value
	}
	options, err := parseOutputOptions(filtered)
	return options, values, err
}
