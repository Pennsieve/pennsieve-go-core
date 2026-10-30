package citation

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readFixture(t *testing.T) *File {
	t.Helper()
	data, err := os.ReadFile("testdata/CITATION.cff")
	require.NoError(t, err)
	f, err := Parse(data)
	require.NoError(t, err)
	return f
}

func TestParse(t *testing.T) {
	f := readFixture(t)

	assert.Equal(t, "1.2.0", f.CFFVersion)
	assert.Equal(t, "software", f.Type)
	assert.Equal(t, "Immune cytometry: batch-corrected clustering", f.Title)
	assert.Equal(t, "1.4.0", f.Version)
	assert.Equal(t, "https://doi.org/10.5281/ZENODO.1234567", f.DOI)
	assert.Equal(t, "2026-10-01", f.DateReleased, "an unquoted date stays as written")
	assert.Equal(t, License{"Apache-2.0"}, f.License)
	assert.Contains(t, f.Abstract, "high-dimensional flow cytometry")
	assert.Equal(t, []string{"cytometry", "FlowSOM", "CytoNorm"}, f.Keywords)
	assert.Equal(t, "https://github.com/pennsieve/workflow-immune-cytometry", f.RepositoryCode)
	assert.Equal(t, "https://app.pennsieve.io", f.URL)
	assert.Len(t, f.Identifiers, 2)
	assert.Equal(t, Identifier{Type: "doi", Value: "10.5281/zenodo.1234567", Description: "The concept DOI"}, f.Identifiers[0])

	require.NotNil(t, f.PreferredCitation)
	assert.Equal(t, "article", f.PreferredCitation.Type)
	assert.Equal(t, "Science", f.PreferredCitation.Journal)
	assert.Equal(t, Year(2020), f.PreferredCitation.Year)
	require.Len(t, f.References, 5)
	assert.Equal(t, Year(2015), f.References[0].Year, "a quoted year is a year")
	assert.Equal(t, 2025, f.References[2].PublicationYear(), "the year of date-published")

	assert.Empty(t, Validate(f))
}

func TestAuthors_PersonsAndEntities(t *testing.T) {
	f := readFixture(t)
	require.Len(t, f.Authors, 3)

	joost := f.Authors[0]
	assert.False(t, joost.IsEntity())
	assert.Equal(t, "Joost Wagenaar", joost.DisplayName())
	assert.Equal(t, "https://orcid.org/0000-0002-1825-0097", joost.ORCID)
	assert.Equal(t, "0000-0002-1825-0097", joost.ORCIDID())
	assert.Equal(t, "University of Pennsylvania", joost.Affiliation)

	assert.Equal(t, "Sofie Van Gassen", f.Authors[1].DisplayName(), "name-particle between given and family")
	assert.Equal(t, "", f.Authors[1].ORCIDID())

	team := f.Authors[2]
	assert.True(t, team.IsEntity())
	assert.Equal(t, "Pennsieve Team", team.DisplayName())

	assert.Equal(t, "Ann Smith, Jr.", Author{GivenNames: "Ann", FamilyNames: "Smith", NameSuffix: "Jr."}.DisplayName())
	assert.Equal(t, "0000-0001-0000-0000", Author{ORCID: "0000-0001-0000-0000"}.ORCIDID())
}

func TestParse_Invalid(t *testing.T) {
	for name, data := range map[string]string{
		"empty":            "   \n",
		"not yaml":         "title: [unclosed",
		"not a mapping":    "- a\n- b\n",
		"keywords string":  "cff-version: 1.2.0\nkeywords: {a: b}\n",
		"year not number":  "references:\n  - title: x\n    year: soon\n",
		"authors not list": "authors: Joost\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Parse([]byte(data))
			assert.Error(t, err)
		})
	}
}

func TestParse_LicenseList(t *testing.T) {
	f, err := Parse([]byte("license:\n  - MIT\n  - Apache-2.0\n"))
	require.NoError(t, err)
	assert.Equal(t, License{"MIT", "Apache-2.0"}, f.License)
}

func TestValidate(t *testing.T) {
	f, err := Parse([]byte("cff-version: 2.0.0\nreferences:\n  - doi: 10.1/x\nauthors:\n  - affiliation: Penn\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{
		`cff-version: "2.0.0" is not supported (expected 1.2.0)`,
		"message: required",
		"title: required",
		"authors[0]: needs family-names/given-names or name",
		"references[0].title: required",
	}, Validate(f))
	assert.Equal(t, []string{"CITATION.cff is missing"}, Validate(nil))
}

func TestNormalizeDOI(t *testing.T) {
	for in, want := range map[string]string{
		"10.1038/NBT.4314":                     "10.1038/nbt.4314",
		" https://doi.org/10.1038/nbt.4314 ":   "10.1038/nbt.4314",
		"http://dx.doi.org/10.1038/nbt.4314":   "10.1038/nbt.4314",
		"HTTPS://DOI.ORG/10.1038/nbt.4314":     "10.1038/nbt.4314",
		"doi:10.1038/nbt.4314":                 "10.1038/nbt.4314",
		"DOI: 10.1038/nbt.4314":                "10.1038/nbt.4314",
		"doi.org/10.12688/f1000research.11622": "10.12688/f1000research.11622",
	} {
		assert.Equal(t, want, NormalizeDOI(in), in)
	}
	assert.True(t, IsDOI("10.1038/nbt.4314"))
	for _, bad := range []string{"", "nbt.4314", "10./x", "10.1038/", "https://example.org"} {
		assert.False(t, IsDOI(bad), bad)
	}
}

func TestRelationTypes(t *testing.T) {
	assert.Len(t, RelationTypes, 16)
	assert.True(t, IsValidRelationType("IsSupplementTo"))
	assert.False(t, IsValidRelationType("Cites"), "Cites is deliberately not allowed")
	assert.False(t, IsValidRelationType("references"), "case matters")
	assert.False(t, IsValidRelationType(""))
}

func TestRelatedWorks(t *testing.T) {
	f := readFixture(t)
	works := RelatedWorks(f, nil)

	require.Len(t, works, 5, "the repeated FlowSOM collapses")
	assert.Equal(t, RelatedWork{
		DOI:          "10.1126/science.abc8511",
		RelationType: IsDescribedBy,
		Title:        "Deep immune profiling of COVID-19 patients reveals distinct immunotypes with therapeutic implications",
		Authors:      []string{"Divij Mathew", "Josephine R. Giles"},
		Year:         2020,
		Journal:      "Science",
		Source:       SourcePreferredCitation,
	}, works[0])

	flowsom := works[1]
	assert.Equal(t, "10.1002/cyto.a.22625", flowsom.DOI, "doi: prefix and case normalised")
	assert.Equal(t, References, flowsom.RelationType)
	assert.Equal(t, SourceReferences, flowsom.Source)
	assert.Equal(t, 2015, flowsom.Year)
	assert.Equal(t, "FlowSOM: Using self-organizing maps for visualization and interpretation of cytometry data", flowsom.Title, "the first listing wins")
	assert.Equal(t, "https://onlinelibrary.wiley.com/doi/10.1002/cyto.a.22625", flowsom.URL, "missing fields filled from the repeat")

	assert.Equal(t, "10.1002/cyto.a.23904", works[2].DOI, "a DOI from identifiers")

	assert.Equal(t, "10.5281/zenodo.15388817", works[3].DOI)
	assert.Equal(t, []string{"Zenodo contributors"}, works[3].Authors)

	flowkit := works[4]
	assert.Equal(t, "", flowkit.DOI, "a work without a DOI is kept")
	assert.Equal(t, "https://github.com/whitews/FlowKit", flowkit.URL)
	assert.Equal(t, References, flowkit.RelationType)

	withDOI := WithDOI(works)
	assert.Len(t, withDOI, 4)
	for _, w := range withDOI {
		assert.NotEmpty(t, w.DOI)
	}

	assert.Nil(t, RelatedWorks(nil, nil))
}

func TestRelatedWorks_Overrides(t *testing.T) {
	f := readFixture(t)
	works := RelatedWorks(f, map[string]string{
		"https://doi.org/10.5281/ZENODO.15388817": IsDerivedFrom,  // any DOI form
		"10.1126/science.abc8511":                 IsSupplementTo, // overrides the preferred citation too
		"10.1002/cyto.a.23904":                    "Cites",        // invalid: ignored
		"10.9999/not.listed":                      Requires,       // not in the file: no work
	})
	byDOI := map[string]RelatedWork{}
	for _, w := range works {
		byDOI[w.DOI] = w
	}
	assert.Equal(t, IsSupplementTo, byDOI["10.1126/science.abc8511"].RelationType)
	assert.Equal(t, IsDerivedFrom, byDOI["10.5281/zenodo.15388817"].RelationType)
	assert.Equal(t, References, byDOI["10.1002/cyto.a.23904"].RelationType, "an invalid type leaves the default")
	assert.NotContains(t, byDOI, "10.9999/not.listed")
	assert.Len(t, works, 5)
}

func TestCheckRelations(t *testing.T) {
	f := readFixture(t)
	assert.Empty(t, CheckRelations(f, map[string]string{
		"doi:10.5281/zenodo.15388817": IsDerivedFrom,
		"10.1002/CYTO.A.23904":        Requires,
	}))
	assert.Empty(t, CheckRelations(f, nil))

	assert.Equal(t, []string{
		`10.1002/cyto.a.23904: "Cites" is not a relation type (one of Describes, Documents, IsCitedBy, IsDerivedFrom, IsDescribedBy, IsDocumentedBy, IsMetadataFor, IsReferencedBy, IsRequiredBy, IsSourceOf, IsSupplementedBy, IsSupplementTo, IsOriginalFormOf, IsVariantFormOf, References, Requires)`,
		"10.9999/not.listed: not a DOI listed in CITATION.cff (preferred-citation or references)",
		"zenodo: not a DOI",
	}, CheckRelations(f, map[string]string{
		"10.1002/cyto.a.23904": "Cites",
		"10.9999/not.listed":   Requires,
		"zenodo":               References,
	}))

	assert.Equal(t, []string{
		"10.1038/nbt.4314: not a DOI listed in CITATION.cff (preferred-citation or references)",
	}, CheckRelations(nil, map[string]string{"10.1038/nbt.4314": References}), "no file lists no DOIs")
}
