package nextinsight
import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// normalize
// ---------------------------------------------------------------------------

func TestNormalize(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		// Basic lowercasing
		{"Production", "production"},
		// Spaces become hyphens
		{"In House", "in-house"},
		// Underscores become hyphens
		{"in_house", "in-house"},
		// Slashes become hyphens
		{"buy/build", "buy-build"},
		// Mixed spaces, underscores, slashes
		{"My App/Name_v2.0", "my-app-name-v2.0"},
		// Leading/trailing hyphens stripped
		{"-leading", "leading"},
		{"trailing-", "trailing"},
		// Special characters dropped
		{"hello@world!", "helloworld"},
		// Multiple consecutive hyphens collapsed
		{"a--b---c", "a-b-c"},
		// Dot preserved
		{"v2.0", "v2.0"},
		// Whitespace trimmed
		{"  trimmed  ", "trimmed"},
		// Empty string
		{"", ""},
		// Already normalised
		{"agile-team", "agile-team"},
		// Exactly 63 chars — unchanged
		{strings.Repeat("a", 63), strings.Repeat("a", 63)},
		// 64 chars — truncated to 63
		{strings.Repeat("a", 64), strings.Repeat("a", 63)},
		// Truncation should not leave trailing hyphen
		{strings.Repeat("a", 62) + "-x", strings.Repeat("a", 62)},
	}

	for _, tc := range cases {
		got := normalize(tc.in)
		if got != tc.want {
			t.Errorf("normalize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// AppMetadata.Labels
// ---------------------------------------------------------------------------

func TestLabels_PopulatesAllFields(t *testing.T) {
	m := &AppMetadata{
		ApplicationID:     "123",
		ApplicationName:   "Platform App",
		Lifecycle:         "Production",
		LifecycleDecision: "Keep",
		Criticality:       "High",
		Complexity:        "Medium",
		Category:          "Infrastructure",
		DevelopmentType:   "In House",
		SourcingType:      "Internal",
		FacingInternet:    "true",
		AgileReleaseTrain: "ART Platform",
		AgileTeam:         "Team Falcon",
	}

	labels := m.Labels()

	expected := map[string]string{
		"next-insight.io/app-id":              "123",
		"next-insight.io/app-name":            "platform-app",
		"next-insight.io/lifecycle":           "production",
		"next-insight.io/lifecycle-decision":  "keep",
		"next-insight.io/criticality":         "high",
		"next-insight.io/complexity":          "medium",
		"next-insight.io/category":            "infrastructure",
		"next-insight.io/development-type":    "in-house",
		"next-insight.io/sourcing-type":       "internal",
		"next-insight.io/facing-internet":     "true",
		"next-insight.io/agile-release-train": "art-platform",
		"next-insight.io/agile-team":          "team-falcon",
	}

	for key, want := range expected {
		if got := labels[key]; got != want {
			t.Errorf("label %q = %q, want %q", key, got, want)
		}
	}
}

func TestLabels_OmitsEmptyFields(t *testing.T) {
	// Only ApplicationID set — all other fields empty.
	m := &AppMetadata{ApplicationID: "42"}
	labels := m.Labels()

	if len(labels) != 1 {
		t.Errorf("expected 1 label for a mostly-empty AppMetadata, got %d: %v", len(labels), labels)
	}
	if labels["next-insight.io/app-id"] != "42" {
		t.Errorf("expected app-id label to be '42', got %q", labels["next-insight.io/app-id"])
	}
}

func TestLabels_EmptyMetadataReturnsEmptyMap(t *testing.T) {
	labels := (&AppMetadata{}).Labels()
	if len(labels) != 0 {
		t.Errorf("expected empty label map, got %v", labels)
	}
}

// ---------------------------------------------------------------------------
// buildMetadata — first-wins for ART and Agile Team
// ---------------------------------------------------------------------------

func TestBuildMetadata_FirstGroupWins(t *testing.T) {
	app := &applicationResponse{}
	app.Data.Name = "My App"

	groups := &groupsResponse{
		Data: []groupItem{
			{Name: "ART-One", Type: groupTypeART},
			{Name: "ART-Two", Type: groupTypeART},
			{Name: "Team-Alpha", Type: groupTypeAgileTeam},
			{Name: "Team-Beta", Type: groupTypeAgileTeam},
		},
	}

	meta := buildMetadata("1", app, groups)

	if meta.AgileReleaseTrain != "ART-One" {
		t.Errorf("AgileReleaseTrain = %q, want %q", meta.AgileReleaseTrain, "ART-One")
	}
	if meta.AgileTeam != "Team-Alpha" {
		t.Errorf("AgileTeam = %q, want %q", meta.AgileTeam, "Team-Alpha")
	}
}

func TestBuildMetadata_UnknownGroupTypeIgnored(t *testing.T) {
	app := &applicationResponse{}
	groups := &groupsResponse{
		Data: []groupItem{
			{Name: "Some Portfolio", Type: "Portfolio"},
		},
	}

	meta := buildMetadata("5", app, groups)

	if meta.AgileReleaseTrain != "" || meta.AgileTeam != "" {
		t.Errorf("expected empty ART and AgileTeam for unknown group type, got ART=%q Team=%q",
			meta.AgileReleaseTrain, meta.AgileTeam)
	}
}

func TestBuildMetadata_FacingInternetLowercased(t *testing.T) {
	app := &applicationResponse{}
	app.Data.FacingInternet = "TRUE"
	groups := &groupsResponse{}

	meta := buildMetadata("3", app, groups)

	if meta.FacingInternet != "true" {
		t.Errorf("FacingInternet = %q, want %q", meta.FacingInternet, "true")
	}
}
