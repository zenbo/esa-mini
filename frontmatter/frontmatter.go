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

// NormalizeLF は \r\n を \n に統一する。
// ファイル保存時に使用し、ローカルの改行コードを LF に揃える。
func NormalizeLF(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// NormalizeCRLF は \n を \r\n に統一する。
// esa API への送信時に使用し、WebUI でのコピペ時に改行が保持されるようにする。
func NormalizeCRLF(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}

func Format(fm Frontmatter, body string) (string, error) {
	body = NormalizeLF(body)

	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(fm); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	yamlStr := buf.String()

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(yamlStr)
	sb.WriteString("---\n\n")
	sb.WriteString(body)
	// Ensure trailing newline
	if !strings.HasSuffix(body, "\n") {
		sb.WriteString("\n")
	}
	return sb.String(), nil
}
