package citation

import (
	"fmt"
	"sort"
	"strings"
)

// The DataCite relation types a Pennsieve publication's related works may
// have (pennsieve-api RelationshipType). Cites is deliberately not one.
const (
	Describes        = "Describes"
	Documents        = "Documents"
	IsCitedBy        = "IsCitedBy"
	IsDerivedFrom    = "IsDerivedFrom"
	IsDescribedBy    = "IsDescribedBy"
	IsDocumentedBy   = "IsDocumentedBy"
	IsMetadataFor    = "IsMetadataFor"
	IsReferencedBy   = "IsReferencedBy"
	IsRequiredBy     = "IsRequiredBy"
	IsSourceOf       = "IsSourceOf"
	IsSupplementedBy = "IsSupplementedBy"
	IsSupplementTo   = "IsSupplementTo"
	IsOriginalFormOf = "IsOriginalFormOf"
	IsVariantFormOf  = "IsVariantFormOf"
	References       = "References"
	Requires         = "Requires"
)

// RelationTypes is every relation type a related work may have.
var RelationTypes = []string{
	Describes, Documents, IsCitedBy, IsDerivedFrom,
	IsDescribedBy, IsDocumentedBy, IsMetadataFor, IsReferencedBy,
	IsRequiredBy, IsSourceOf, IsSupplementedBy, IsSupplementTo,
	IsOriginalFormOf, IsVariantFormOf, References, Requires,
}

// IsValidRelationType is whether t is one of RelationTypes (exact case).
func IsValidRelationType(t string) bool {
	for _, v := range RelationTypes {
		if v == t {
			return true
		}
	}
	return false
}

// Where a related work comes from in CITATION.cff.
const (
	SourcePreferredCitation = "preferred-citation"
	SourceReferences        = "references"
)

// RelatedWork is a work CITATION.cff lists, typed for DataCite. DOI is
// normalised (10.1038/nbt.4314), or empty for a work without one, which is
// kept so it can be shown (by its URL) but can't go to DataCite; see
// WithDOI.
type RelatedWork struct {
	DOI          string   `json:"doi"`
	RelationType string   `json:"relationType"`
	Title        string   `json:"title,omitempty"`
	Authors      []string `json:"authors,omitempty"`
	Year         int      `json:"year,omitempty"`
	Journal      string   `json:"journal,omitempty"`
	URL          string   `json:"url,omitempty"`
	Source       string   `json:"source"`
}

// RelatedWorks is the works f lists: its preferred-citation first
// (IsDescribedBy), then its references (References), in file order. The
// relations map (DOI → relation type, DOIs in any form NormalizeDOI takes)
// types a work otherwise; a relation with an invalid type is ignored, and
// CheckRelations reports it. A work listed twice (the same DOI) is kept
// once, where it first appears, its missing fields filled from the repeats.
// Nil f gives nil.
func RelatedWorks(f *File, relations map[string]string) []RelatedWork {
	if f == nil {
		return nil
	}
	override := normalizeRelations(relations)

	var out []RelatedWork
	index := map[string]int{}
	add := func(r Reference, source, relation string) {
		w := RelatedWork{
			RelationType: relation,
			Title:        strings.TrimSpace(r.Title),
			Year:         r.PublicationYear(),
			Journal:      strings.TrimSpace(r.Journal),
			URL:          strings.TrimSpace(r.URL),
			Source:       source,
		}
		for _, a := range r.Authors {
			if n := a.DisplayName(); n != "" {
				w.Authors = append(w.Authors, n)
			}
		}
		dois := r.DOIs()
		if len(dois) == 0 {
			if w.Title == "" && w.URL == "" {
				return
			}
			out = append(out, w)
			return
		}
		w.DOI = dois[0]
		if t, ok := override[w.DOI]; ok {
			w.RelationType = t
		}
		if i, ok := index[w.DOI]; ok {
			out[i] = fill(out[i], w)
			return
		}
		index[w.DOI] = len(out)
		out = append(out, w)
	}

	if f.PreferredCitation != nil {
		add(*f.PreferredCitation, SourcePreferredCitation, IsDescribedBy)
	}
	for _, r := range f.References {
		add(r, SourceReferences, References)
	}
	return out
}

// fill is w with its empty fields taken from more.
func fill(w, more RelatedWork) RelatedWork {
	if w.Title == "" {
		w.Title = more.Title
	}
	if len(w.Authors) == 0 {
		w.Authors = more.Authors
	}
	if w.Year == 0 {
		w.Year = more.Year
	}
	if w.Journal == "" {
		w.Journal = more.Journal
	}
	if w.URL == "" {
		w.URL = more.URL
	}
	return w
}

// WithDOI is the works that have a DOI: the ones DataCite takes.
func WithDOI(works []RelatedWork) []RelatedWork {
	var out []RelatedWork
	for _, w := range works {
		if w.DOI != "" {
			out = append(out, w)
		}
	}
	return out
}

// CheckRelations is how a relations map (DOI → relation type) doesn't fit
// f: a type that isn't one of RelationTypes, a key that isn't a DOI, or a
// DOI that f doesn't list (in preferred-citation or references). One
// "<doi>: reason" each, sorted by DOI; empty when it fits. Nil f lists no
// DOIs.
func CheckRelations(f *File, relations map[string]string) []string {
	listed := map[string]bool{}
	if f != nil {
		if f.PreferredCitation != nil {
			for _, d := range f.PreferredCitation.DOIs() {
				listed[d] = true
			}
		}
		for _, r := range f.References {
			for _, d := range r.DOIs() {
				listed[d] = true
			}
		}
	}

	keys := make([]string, 0, len(relations))
	for k := range relations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var problems []string
	for _, k := range keys {
		t := relations[k]
		d := NormalizeDOI(k)
		if !IsValidRelationType(t) {
			problems = append(problems, fmt.Sprintf("%s: %q is not a relation type (one of %s)", k, t, strings.Join(RelationTypes, ", ")))
		}
		switch {
		case !IsDOI(d):
			problems = append(problems, fmt.Sprintf("%s: not a DOI", k))
		case !listed[d]:
			problems = append(problems, fmt.Sprintf("%s: not a DOI listed in CITATION.cff (preferred-citation or references)", k))
		}
	}
	return problems
}

// normalizeRelations is relations keyed by normalised DOI, keeping only
// valid types.
func normalizeRelations(relations map[string]string) map[string]string {
	out := map[string]string{}
	for k, t := range relations {
		if d := NormalizeDOI(k); IsDOI(d) && IsValidRelationType(t) {
			out[d] = t
		}
	}
	return out
}
