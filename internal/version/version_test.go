package version

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			name:  "valid version without v prefix",
			input: "1.2.3",
			want:  &Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "valid version with v prefix",
			input: "v1.2.3",
			want:  &Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:  "zero version",
			input: "0.0.0",
			want:  &Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:  "large numbers",
			input: "100.200.300",
			want:  &Version{Major: 100, Minor: 200, Patch: 300},
		},
		{
			name:    "invalid format - missing patch",
			input:   "1.2",
			wantErr: true,
		},
		{
			name:    "invalid format - extra components",
			input:   "1.2.3.4",
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric",
			input:   "1.2.alpha",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid prefix",
			input:   "version1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch {
					t.Errorf("Parse() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version *Version
		want    string
	}{
		{
			name:    "standard version",
			version: &Version{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3",
		},
		{
			name:    "zero version",
			version: &Version{Major: 0, Minor: 0, Patch: 0},
			want:    "0.0.0",
		},
		{
			name:    "large numbers",
			version: &Version{Major: 100, Minor: 200, Patch: 300},
			want:    "100.200.300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		name  string
		v1    *Version
		v2    *Version
		want  int
	}{
		{
			name: "equal versions",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3},
			v2:   &Version{Major: 1, Minor: 2, Patch: 3},
			want: 0,
		},
		{
			name: "major version difference",
			v1:   &Version{Major: 2, Minor: 0, Patch: 0},
			v2:   &Version{Major: 1, Minor: 0, Patch: 0},
			want: 1,
		},
		{
			name: "minor version difference",
			v1:   &Version{Major: 1, Minor: 2, Patch: 0},
			v2:   &Version{Major: 1, Minor: 1, Patch: 0},
			want: 1,
		},
		{
			name: "patch version difference",
			v1:   &Version{Major: 1, Minor: 2, Patch: 3},
			v2:   &Version{Major: 1, Minor: 2, Patch: 2},
			want: 1,
		},
		{
			name: "v1 less than v2",
			v1:   &Version{Major: 1, Minor: 0, Patch: 0},
			v2:   &Version{Major: 2, Minor: 0, Patch: 0},
			want: -1,
		},
		{
			name: "v1 less than v2 - minor",
			v1:   &Version{Major: 1, Minor: 1, Patch: 0},
			v2:   &Version{Major: 1, Minor: 2, Patch: 0},
			want: -1,
		},
		{
			name: "v1 less than v2 - patch",
			v1:   &Version{Major: 1, Minor: 1, Patch: 1},
			v2:   &Version{Major: 1, Minor: 1, Patch: 2},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Compare(tt.v2)
			if (got > 0 && tt.want <= 0) || (got < 0 && tt.want >= 0) || (got == 0 && tt.want != 0) {
				t.Errorf("Version.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExactConstraint_Satisfies(t *testing.T) {
	tests := []struct {
		name       string
		constraint *ExactConstraint
		version    *Version
		want       bool
	}{
		{
			name:       "exact match",
			constraint: &ExactConstraint{version: &Version{Major: 1, Minor: 2, Patch: 3}},
			version:    &Version{Major: 1, Minor: 2, Patch: 3},
			want:       true,
		},
		{
			name:       "different major",
			constraint: &ExactConstraint{version: &Version{Major: 1, Minor: 2, Patch: 3}},
			version:    &Version{Major: 2, Minor: 2, Patch: 3},
			want:       false,
		},
		{
			name:       "different minor",
			constraint: &ExactConstraint{version: &Version{Major: 1, Minor: 2, Patch: 3}},
			version:    &Version{Major: 1, Minor: 3, Patch: 3},
			want:       false,
		},
		{
			name:       "different patch",
			constraint: &ExactConstraint{version: &Version{Major: 1, Minor: 2, Patch: 3}},
			version:    &Version{Major: 1, Minor: 2, Patch: 4},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.constraint.Satisfies(tt.version); got != tt.want {
				t.Errorf("ExactConstraint.Satisfies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExactConstraint_String(t *testing.T) {
	c := &ExactConstraint{version: &Version{Major: 1, Minor: 2, Patch: 3}}
	if got := c.String(); got != "1.2.3" {
		t.Errorf("ExactConstraint.String() = %v, want 1.2.3", got)
	}
}

func TestRangeConstraint_Satisfies(t *testing.T) {
	tests := []struct {
		name       string
		constraint *RangeConstraint
		version    *Version
		want       bool
	}{
		{
			name: "version within inclusive range",
			constraint: &RangeConstraint{
				min:          &Version{Major: 1, Minor: 0, Patch: 0},
				max:          &Version{Major: 2, Minor: 0, Patch: 0},
				minInclusive: true,
				maxInclusive: true,
			},
			version: &Version{Major: 1, Minor: 5, Patch: 0},
			want:    true,
		},
		{
			name: "version at min boundary (inclusive)",
			constraint: &RangeConstraint{
				min:          &Version{Major: 1, Minor: 0, Patch: 0},
				minInclusive: true,
			},
			version: &Version{Major: 1, Minor: 0, Patch: 0},
			want:    true,
		},
		{
			name: "version at min boundary (exclusive)",
			constraint: &RangeConstraint{
				min:          &Version{Major: 1, Minor: 0, Patch: 0},
				minInclusive: false,
			},
			version: &Version{Major: 1, Minor: 0, Patch: 0},
			want:    false,
		},
		{
			name: "version at max boundary (inclusive)",
			constraint: &RangeConstraint{
				max:          &Version{Major: 2, Minor: 0, Patch: 0},
				maxInclusive: true,
			},
			version: &Version{Major: 2, Minor: 0, Patch: 0},
			want:    true,
		},
		{
			name: "version at max boundary (exclusive)",
			constraint: &RangeConstraint{
				max:          &Version{Major: 2, Minor: 0, Patch: 0},
				maxInclusive: false,
			},
			version: &Version{Major: 2, Minor: 0, Patch: 0},
			want:    false,
		},
		{
			name: "version below min",
			constraint: &RangeConstraint{
				min:          &Version{Major: 1, Minor: 0, Patch: 0},
				minInclusive: true,
			},
			version: &Version{Major: 0, Minor: 9, Patch: 0},
			want:    false,
		},
		{
			name: "version above max",
			constraint: &RangeConstraint{
				max:          &Version{Major: 2, Minor: 0, Patch: 0},
				maxInclusive: false,
			},
			version: &Version{Major: 2, Minor: 1, Patch: 0},
			want:    false,
		},
		{
			name: "no constraints",
			constraint: &RangeConstraint{},
			version: &Version{Major: 999, Minor: 999, Patch: 999},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.constraint.Satisfies(tt.version); got != tt.want {
				t.Errorf("RangeConstraint.Satisfies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRangeConstraint_String(t *testing.T) {
	tests := []struct {
		name       string
		constraint *RangeConstraint
		want       string
	}{
		{
			name: "min only inclusive",
			constraint: &RangeConstraint{
				min:          &Version{Major: 1, Minor: 0, Patch: 0},
				minInclusive: true,
			},
			want: ">=1.0.0",
		},
		{
			name: "min only exclusive",
			constraint: &RangeConstraint{
				min:          &Version{Major: 1, Minor: 0, Patch: 0},
				minInclusive: false,
			},
			want: ">1.0.0",
		},
		{
			name: "max only inclusive",
			constraint: &RangeConstraint{
				max:          &Version{Major: 2, Minor: 0, Patch: 0},
				maxInclusive: true,
			},
			want: "<=2.0.0",
		},
		{
			name: "max only exclusive",
			constraint: &RangeConstraint{
				max:          &Version{Major: 2, Minor: 0, Patch: 0},
				maxInclusive: false,
			},
			want: "<2.0.0",
		},
		{
			name: "both min and max",
			constraint: &RangeConstraint{
				min:          &Version{Major: 1, Minor: 0, Patch: 0},
				max:          &Version{Major: 2, Minor: 0, Patch: 0},
				minInclusive: true,
				maxInclusive: false,
			},
			want: ">=1.0.0, <2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.constraint.String(); got != tt.want {
				t.Errorf("RangeConstraint.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, c Constraint)
	}{
		{
			name:  "exact version",
			input: "1.2.3",
			check: func(t *testing.T, c Constraint) {
				if _, ok := c.(*ExactConstraint); !ok {
					t.Errorf("expected ExactConstraint, got %T", c)
				}
				v := &Version{Major: 1, Minor: 2, Patch: 3}
				if !c.Satisfies(v) {
					t.Errorf("constraint should satisfy version 1.2.3")
				}
			},
		},
		{
			name:  "caret constraint",
			input: "^1.2.3",
			check: func(t *testing.T, c Constraint) {
				rc, ok := c.(*RangeConstraint)
				if !ok {
					t.Errorf("expected RangeConstraint, got %T", c)
					return
				}
				// Should allow 1.2.3 to 2.0.0 (exclusive)
				if !rc.Satisfies(&Version{Major: 1, Minor: 2, Patch: 3}) {
					t.Errorf("should satisfy 1.2.3")
				}
				if !rc.Satisfies(&Version{Major: 1, Minor: 9, Patch: 9}) {
					t.Errorf("should satisfy 1.9.9")
				}
				if rc.Satisfies(&Version{Major: 2, Minor: 0, Patch: 0}) {
					t.Errorf("should not satisfy 2.0.0")
				}
			},
		},
		{
			name:  "tilde constraint",
			input: "~1.2.3",
			check: func(t *testing.T, c Constraint) {
				rc, ok := c.(*RangeConstraint)
				if !ok {
					t.Errorf("expected RangeConstraint, got %T", c)
					return
				}
				// Should allow 1.2.3 to 1.3.0 (exclusive)
				if !rc.Satisfies(&Version{Major: 1, Minor: 2, Patch: 3}) {
					t.Errorf("should satisfy 1.2.3")
				}
				if !rc.Satisfies(&Version{Major: 1, Minor: 2, Patch: 9}) {
					t.Errorf("should satisfy 1.2.9")
				}
				if rc.Satisfies(&Version{Major: 1, Minor: 3, Patch: 0}) {
					t.Errorf("should not satisfy 1.3.0")
				}
			},
		},
		{
			name:  "range constraint",
			input: ">=1.0.0, <2.0.0",
			check: func(t *testing.T, c Constraint) {
				rc, ok := c.(*RangeConstraint)
				if !ok {
					t.Errorf("expected RangeConstraint, got %T", c)
					return
				}
				if !rc.Satisfies(&Version{Major: 1, Minor: 0, Patch: 0}) {
					t.Errorf("should satisfy 1.0.0")
				}
				if !rc.Satisfies(&Version{Major: 1, Minor: 9, Patch: 9}) {
					t.Errorf("should satisfy 1.9.9")
				}
				if rc.Satisfies(&Version{Major: 2, Minor: 0, Patch: 0}) {
					t.Errorf("should not satisfy 2.0.0")
				}
			},
		},
		{
			name:  "single comparison",
			input: ">=1.0.0",
			check: func(t *testing.T, c Constraint) {
				rc, ok := c.(*RangeConstraint)
				if !ok {
					t.Errorf("expected RangeConstraint, got %T", c)
					return
				}
				if !rc.Satisfies(&Version{Major: 1, Minor: 0, Patch: 0}) {
					t.Errorf("should satisfy 1.0.0")
				}
				if !rc.Satisfies(&Version{Major: 999, Minor: 0, Patch: 0}) {
					t.Errorf("should satisfy 999.0.0")
				}
				if rc.Satisfies(&Version{Major: 0, Minor: 9, Patch: 9}) {
					t.Errorf("should not satisfy 0.9.9")
				}
			},
		},
		{
			name:  "whitespace handling",
			input: "  >=1.0.0  ,  <2.0.0  ",
			check: func(t *testing.T, c Constraint) {
				rc, ok := c.(*RangeConstraint)
				if !ok {
					t.Errorf("expected RangeConstraint, got %T", c)
					return
				}
				if !rc.Satisfies(&Version{Major: 1, Minor: 5, Patch: 0}) {
					t.Errorf("should satisfy 1.5.0")
				}
			},
		},
		{
			name:    "invalid version in constraint",
			input:   "^1.2.invalid",
			wantErr: true,
		},
		{
			name:    "invalid range format",
			input:   ">=1.0.0, <2.0.0, <3.0.0",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConstraint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}