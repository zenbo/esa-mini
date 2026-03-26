package frontmatter

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

type Frontmatter struct {
	Number    int      `yaml:"number,omitempty"`
	Title     string   `yaml:"title"`
	URL       string   `yaml:"url,omitempty"`
	Category  string   `yaml:"category,omitempty"`
	Tags      []string `yaml:"tags,omitempty"`
	WIP       *bool    `yaml:"wip,omitempty"`
	UpdatedAt string   `yaml:"updated_at,omitempty"`
}

type Document struct {
	Frontmatter Frontmatter
	Body        string
}

func Parse(r io.Reader) (*Document, error) {
	scanner := bufio.NewScanner(r)

	// Expect opening ---
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return nil, fmt.Errorf("expected frontmatter opening '---'")
	}

	var yamlLines []string
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			found = true
			break
		}
		yamlLines = append(yamlLines, line)
	}
	if !found {
		return nil, fmt.Errorf("expected frontmatter closing '---'")
	}

	var fm Frontmatter
	yamlStr := strings.Join(yamlLines, "\n")
	if err := yaml.Unmarshal([]byte(yamlStr), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	var bodyLines []string
	// Skip one blank line after frontmatter if present
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		if firstLine && line == "" {
			firstLine = false
			continue
		}
		firstLine = false
		bodyLines = append(bodyLines, line)
	}

	return &Document{
		Frontmatter: fm,
		Body:        strings.Join(bodyLines, "\n"),
	}, nil
}

func Format(fm Frontmatter, body string) (string, error) {
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(yamlBytes)
	sb.WriteString("---\n\n")
	sb.WriteString(body)
	// Ensure trailing newline
	if !strings.HasSuffix(body, "\n") {
		sb.WriteString("\n")
	}
	return sb.String(), nil
}
