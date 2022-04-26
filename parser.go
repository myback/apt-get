package apt

import (
	"bufio"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

func parse(file io.Reader) (map[string]Package, error) {
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 131072), 0)

	onDoubleNewLine := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if i > 0 && data[i-1] == '\n' && data[i] == '\n' {
				return i + 1, data[:i], nil
			}
		}
		if !atEOF {
			return 0, nil, nil
		}

		return 0, data, bufio.ErrFinalToken
	}
	scanner.Split(onDoubleNewLine)

	packages := map[string]Package{}
	for scanner.Scan() {
		b := scanner.Bytes()
		s := scanner.Text()
		if s == "" {
			continue
		}

		var pkg Package
		if err := yaml.Unmarshal(b, &pkg); err != nil {
			b, err = yamlNormalizer(s)
			if err != nil {
				return nil, err
			}

			if err := yaml.Unmarshal(b, &pkg); err != nil {
				return nil, err
			}
		}
		packages[pkg.Package] = pkg
	}

	return packages, scanner.Err()
}

func yamlNormalizer(s string) ([]byte, error) {
	var prevKey string

	m := map[string]string{}
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(line, " ") {
			m[prevKey] += " " + strings.TrimSpace(line)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		idx := strings.IndexByte(line, ':')
		prevKey = line[:idx]
		m[prevKey] = line[idx+2:]
	}

	return yaml.Marshal(m)
}
