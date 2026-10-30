// Package citation reads a repository's CITATION.cff (Citation File Format
// 1.2.0, https://citation-file-format.github.io) and turns the works it
// lists into DataCite related works.
//
// CITATION.cff is where an app or repository keeps its citation metadata:
// who wrote it, how to cite it (preferred-citation) and the works it builds
// on (references). RelatedWorks maps those onto DataCite relatedIdentifiers:
// the preferred citation describes the software (IsDescribedBy), every
// reference is References, and a relations map (DOI → relation type, from
// app.yml) types the ones that are something else.
package citation

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

// File is a parsed CITATION.cff. Only the fields Pennsieve uses are kept;
// anything else in the file is ignored.
type File struct {
	CFFVersion        string       `json:"cffVersion,omitempty" yaml:"cff-version"`
	Message           string       `json:"message,omitempty" yaml:"message"`
	Type              string       `json:"type,omitempty" yaml:"type"`
	Title             string       `json:"title,omitempty" yaml:"title"`
	Version           string       `json:"version,omitempty" yaml:"version"`
	DOI               string       `json:"doi,omitempty" yaml:"doi"`
	DateReleased      string       `json:"dateReleased,omitempty" yaml:"date-released"`
	License           License      `json:"license,omitempty" yaml:"license"`
	Abstract          string       `json:"abstract,omitempty" yaml:"abstract"`
	Keywords          []string     `json:"keywords,omitempty" yaml:"keywords"`
	RepositoryCode    string       `json:"repositoryCode,omitempty" yaml:"repository-code"`
	URL               string       `json:"url,omitempty" yaml:"url"`
	Authors           []Author     `json:"authors,omitempty" yaml:"authors"`
	Identifiers       []Identifier `json:"identifiers,omitempty" yaml:"identifiers"`
	PreferredCitation *Reference   `json:"preferredCitation,omitempty" yaml:"preferred-citation"`
	References        []Reference  `json:"references,omitempty" yaml:"references"`
}

// Author is a CFF person (GivenNames/FamilyNames) or entity (Name: an
// organisation, a consortium, a project).
type Author struct {
	GivenNames   string `json:"givenNames,omitempty" yaml:"given-names"`
	FamilyNames  string `json:"familyNames,omitempty" yaml:"family-names"`
	NameParticle string `json:"nameParticle,omitempty" yaml:"name-particle"`
	NameSuffix   string `json:"nameSuffix,omitempty" yaml:"name-suffix"`
	Name         string `json:"name,omitempty" yaml:"name"`
	ORCID        string `json:"orcid,omitempty" yaml:"orcid"`
	Affiliation  string `json:"affiliation,omitempty" yaml:"affiliation"`
	Email        string `json:"email,omitempty" yaml:"email"`
}

// IsEntity is whether the author is an entity (an organisation) rather than
// a person.
func (a Author) IsEntity() bool {
	return a.GivenNames == "" && a.FamilyNames == "" && a.Name != ""
}

// DisplayName is the author as shown: "Given Particle Family, Suffix" for a
// person, the name for an entity.
func (a Author) DisplayName() string {
	if a.IsEntity() {
		return strings.TrimSpace(a.Name)
	}
	var parts []string
	for _, p := range []string{a.GivenNames, a.NameParticle, a.FamilyNames} {
		if p = strings.TrimSpace(p); p != "" {
			parts = append(parts, p)
		}
	}
	name := strings.Join(parts, " ")
	if s := strings.TrimSpace(a.NameSuffix); s != "" && name != "" {
		name += ", " + s
	}
	return name
}

// ORCIDID is the author's bare ORCID iD (0000-0002-1825-0097), without the
// https://orcid.org/ that CFF writes it with; empty when there's none.
func (a Author) ORCIDID() string {
	o := strings.TrimSpace(a.ORCID)
	for _, p := range []string{"https://orcid.org/", "http://orcid.org/", "orcid.org/"} {
		if len(o) >= len(p) && strings.EqualFold(o[:len(p)], p) {
			return o[len(p):]
		}
	}
	return o
}

// Identifier is one of a work's identifiers; Type is doi, url, swh or other.
type Identifier struct {
	Type        string `json:"type,omitempty" yaml:"type"`
	Value       string `json:"value,omitempty" yaml:"value"`
	Description string `json:"description,omitempty" yaml:"description"`
}

// Reference is a work in preferred-citation or references.
type Reference struct {
	Type          string       `json:"type,omitempty" yaml:"type"`
	Title         string       `json:"title,omitempty" yaml:"title"`
	Authors       []Author     `json:"authors,omitempty" yaml:"authors"`
	DOI           string       `json:"doi,omitempty" yaml:"doi"`
	Year          Year         `json:"year,omitempty" yaml:"year"`
	DatePublished string       `json:"datePublished,omitempty" yaml:"date-published"`
	Journal       string       `json:"journal,omitempty" yaml:"journal"`
	Volume        string       `json:"volume,omitempty" yaml:"volume"`
	Issue         string       `json:"issue,omitempty" yaml:"issue"`
	Start         string       `json:"start,omitempty" yaml:"start"`
	End           string       `json:"end,omitempty" yaml:"end"`
	URL           string       `json:"url,omitempty" yaml:"url"`
	Identifiers   []Identifier `json:"identifiers,omitempty" yaml:"identifiers"`
}

// DOIs is every DOI the work carries, normalised: its doi and its
// identifiers of type doi, without duplicates.
func (r Reference) DOIs() []string {
	var out []string
	seen := map[string]bool{}
	add := func(raw string) {
		if d := NormalizeDOI(raw); IsDOI(d) && !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	add(r.DOI)
	for _, id := range r.Identifiers {
		if strings.EqualFold(strings.TrimSpace(id.Type), "doi") {
			add(id.Value)
		}
	}
	return out
}

// PublicationYear is Year, or the year of DatePublished; 0 when neither.
func (r Reference) PublicationYear() int {
	if r.Year != 0 {
		return int(r.Year)
	}
	if len(r.DatePublished) >= 4 {
		if y, err := strconv.Atoi(r.DatePublished[:4]); err == nil {
			return y
		}
	}
	return 0
}

// Year is a reference's year, written as a number or a string.
type Year int

// UnmarshalYAML takes 2021 and "2021"; an empty string is no year.
func (y *Year) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: year must be a number", n.Line)
	}
	v := strings.TrimSpace(n.Value)
	if v == "" {
		*y = 0
		return nil
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("line %d: year %q is not a number", n.Line, n.Value)
	}
	*y = Year(i)
	return nil
}

// License is a CFF license: one SPDX id, or a list of them.
type License []string

// UnmarshalYAML takes a single id or a list.
func (l *License) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		if v := strings.TrimSpace(n.Value); v != "" {
			*l = License{v}
		}
		return nil
	case yaml.SequenceNode:
		var s []string
		if err := n.Decode(&s); err != nil {
			return err
		}
		*l = s
		return nil
	}
	return fmt.Errorf("line %d: license must be an SPDX id or a list of them", n.Line)
}

// Parse reads a CITATION.cff. It fails when the file isn't YAML, isn't a
// mapping, or a field has the wrong shape (keywords that aren't a list, a
// year that isn't a number). It doesn't check that required fields are
// there; Validate does.
func Parse(data []byte) (*File, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, errors.New("CITATION.cff is empty")
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("CITATION.cff is not valid YAML: %w", err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("CITATION.cff must be a YAML mapping")
	}
	var f File
	if err := root.Content[0].Decode(&f); err != nil {
		return nil, fmt.Errorf("CITATION.cff: %w", err)
	}
	return &f, nil
}

// Validate is how the file misses what CFF 1.2.0 requires: cff-version
// (1.x), message, title and authors, and a title for each reference. Empty
// when it doesn't.
func Validate(f *File) []string {
	if f == nil {
		return []string{"CITATION.cff is missing"}
	}
	var problems []string
	switch v := strings.TrimSpace(f.CFFVersion); {
	case v == "":
		problems = append(problems, "cff-version: required")
	case !strings.HasPrefix(v, "1."):
		problems = append(problems, fmt.Sprintf("cff-version: %q is not supported (expected 1.2.0)", v))
	}
	if strings.TrimSpace(f.Message) == "" {
		problems = append(problems, "message: required")
	}
	if strings.TrimSpace(f.Title) == "" {
		problems = append(problems, "title: required")
	}
	if len(f.Authors) == 0 {
		problems = append(problems, "authors: required")
	}
	for i, a := range f.Authors {
		if a.DisplayName() == "" {
			problems = append(problems, fmt.Sprintf("authors[%d]: needs family-names/given-names or name", i))
		}
	}
	if f.PreferredCitation != nil && strings.TrimSpace(f.PreferredCitation.Title) == "" {
		problems = append(problems, "preferred-citation.title: required")
	}
	for i, r := range f.References {
		if strings.TrimSpace(r.Title) == "" {
			problems = append(problems, fmt.Sprintf("references[%d].title: required", i))
		}
	}
	return problems
}

// NormalizeDOI is a DOI as Pennsieve keeps it: trimmed, lowercase, without
// https://doi.org/, https://dx.doi.org/ or doi:.
func NormalizeDOI(raw string) string {
	d := strings.ToLower(strings.TrimSpace(raw))
	for _, p := range []string{"https://doi.org/", "http://doi.org/", "https://dx.doi.org/", "http://dx.doi.org/", "doi.org/", "doi:"} {
		if strings.HasPrefix(d, p) {
			d = strings.TrimSpace(d[len(p):])
			break
		}
	}
	return d
}

// IsDOI is whether a normalised DOI looks like one: 10.<registrant>/<suffix>.
func IsDOI(d string) bool {
	if !strings.HasPrefix(d, "10.") {
		return false
	}
	slash := strings.Index(d, "/")
	return slash > len("10.") && slash < len(d)-1
}
