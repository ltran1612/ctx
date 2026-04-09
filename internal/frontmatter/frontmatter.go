package frontmatter

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Meta holds the YAML frontmatter of a context.md file.
type Meta struct {
	ID      string    `yaml:"id"`
	Slug    string    `yaml:"slug"`
	Title   string    `yaml:"title"`
	Status  string    `yaml:"status"`
	Created time.Time `yaml:"created"`
	Updated time.Time `yaml:"updated"`
	Tags    []string  `yaml:"tags"`
	Ticket  string    `yaml:"ticket,omitempty"`
}

// File is the parsed representation of a context.md file.
type File struct {
	Meta Meta
	Body string // everything after the closing ---
}

const separator = "---"

// Parse splits a context.md file into its frontmatter and body.
func Parse(content string) (*File, error) {
	content = strings.TrimPrefix(content, "\xef\xbb\xbf") // strip BOM if present

	if !strings.HasPrefix(content, separator) {
		return &File{Body: content}, nil
	}

	// find closing ---
	rest := content[len(separator):]
	idx := strings.Index(rest, "\n"+separator)
	if idx == -1 {
		return nil, fmt.Errorf("frontmatter: missing closing ---")
	}

	rawYAML := rest[:idx]
	body := rest[idx+len("\n"+separator):]
	if strings.HasPrefix(body, "\n") {
		body = body[1:]
	}

	var meta Meta
	if err := yaml.Unmarshal([]byte(rawYAML), &meta); err != nil {
		return nil, fmt.Errorf("frontmatter: %w", err)
	}

	return &File{Meta: meta, Body: body}, nil
}

// Serialize encodes a File back to markdown text.
func Serialize(f *File) (string, error) {
	var buf bytes.Buffer

	buf.WriteString(separator + "\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(f.Meta); err != nil {
		return "", err
	}
	buf.WriteString(separator + "\n")
	if f.Body != "" {
		buf.WriteString("\n")
		buf.WriteString(f.Body)
	}

	return buf.String(), nil
}

// Template returns the default content for a new topic.
func Template(meta Meta) (string, error) {
	f := &File{
		Meta: meta,
		Body: `## Summary

<!-- 1-3 sentence description of what this work is about. AI agents read this first. -->

## Goal

<!-- The specific outcome you are trying to achieve. -->

## Context

<!-- Background, constraints, decisions already made. -->

## Current State

<!-- Where things stand. What's done, what's blocked. -->

## Next Steps

- [ ]

## Notes

<!-- Scratchpad: links, commands, observations. -->
`,
	}
	return Serialize(f)
}

// Section extracts the content of a named ## heading from the body.
func Section(body, name string) (string, bool) {
	heading := "## " + name
	lines := strings.Split(body, "\n")
	var inside bool
	var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			if inside {
				break
			}
			if strings.EqualFold(strings.TrimPrefix(line, "## "), name) {
				inside = true
			}
			continue
		}
		_ = heading
		if inside {
			result = append(result, line)
		}
	}
	if !inside {
		return "", false
	}
	return strings.TrimSpace(strings.Join(result, "\n")), true
}

// AppendToSection appends text to a named ## section (creates the section if missing).
func AppendToSection(body, section, text string) string {
	heading := "## " + section
	if !strings.Contains(body, heading) {
		return body + "\n" + heading + "\n\n" + text + "\n"
	}
	lines := strings.Split(body, "\n")
	var out []string
	var i int
	for i < len(lines) {
		line := lines[i]
		out = append(out, line)
		if strings.TrimSpace(line) == heading {
			// advance past blank lines after heading
			i++
			for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
				out = append(out, lines[i])
				i++
			}
			// find end of this section
			sectionEnd := i
			for sectionEnd < len(lines) && !strings.HasPrefix(lines[sectionEnd], "## ") {
				sectionEnd++
			}
			out = append(out, lines[i:sectionEnd]...)
			out = append(out, text)
			out = append(out, lines[sectionEnd:]...)
			return strings.Join(out, "\n")
		}
		i++
	}
	return strings.Join(out, "\n")
}

// PrependToSection prepends a timestamped note to a section.
func PrependToSection(body, section, text string) string {
	heading := "## " + section
	timestamp := time.Now().Format("2006-01-02 15:04")
	note := fmt.Sprintf("**%s** — %s", timestamp, text)

	if !strings.Contains(body, heading) {
		return body + "\n" + heading + "\n\n" + note + "\n"
	}
	lines := strings.Split(body, "\n")
	var out []string
	for i, line := range lines {
		out = append(out, line)
		if strings.TrimSpace(line) == heading {
			// insert note immediately after heading
			out = append(out, "")
			out = append(out, note)
			out = append(out, lines[i+1:]...)
			return strings.Join(out, "\n")
		}
	}
	return strings.Join(out, "\n")
}
