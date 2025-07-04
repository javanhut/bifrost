package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func Parse(v string) (*Version, error) {
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

type Constraint interface {
	Satisfies(v *Version) bool
	String() string
}

type ExactConstraint struct {
	version *Version
}

func (c *ExactConstraint) Satisfies(v *Version) bool {
	return c.version.Compare(v) == 0
}

func (c *ExactConstraint) String() string {
	return c.version.String()
}

type RangeConstraint struct {
	min          *Version
	max          *Version
	minInclusive bool
	maxInclusive bool
}

func (c *RangeConstraint) Satisfies(v *Version) bool {
	if c.min != nil {
		cmp := v.Compare(c.min)
		if cmp < 0 || (cmp == 0 && !c.minInclusive) {
			return false
		}
	}
	if c.max != nil {
		cmp := v.Compare(c.max)
		if cmp > 0 || (cmp == 0 && !c.maxInclusive) {
			return false
		}
	}
	return true
}

func (c *RangeConstraint) String() string {
	var parts []string
	if c.min != nil {
		op := ">="
		if !c.minInclusive {
			op = ">"
		}
		parts = append(parts, op+c.min.String())
	}
	if c.max != nil {
		op := "<="
		if !c.maxInclusive {
			op = "<"
		}
		parts = append(parts, op+c.max.String())
	}
	return strings.Join(parts, ", ")
}

func ParseConstraint(s string) (Constraint, error) {
	s = strings.TrimSpace(s)

	// Caret constraint (^1.2.3)
	if strings.HasPrefix(s, "^") {
		v, err := Parse(s[1:])
		if err != nil {
			return nil, err
		}
		max := &Version{Major: v.Major + 1, Minor: 0, Patch: 0}
		return &RangeConstraint{
			min:          v,
			max:          max,
			minInclusive: true,
			maxInclusive: false,
		}, nil
	}

	// Tilde constraint (~1.2.3)
	if strings.HasPrefix(s, "~") {
		v, err := Parse(s[1:])
		if err != nil {
			return nil, err
		}
		max := &Version{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
		return &RangeConstraint{
			min:          v,
			max:          max,
			minInclusive: true,
			maxInclusive: false,
		}, nil
	}

	// Range constraint (>=1.0.0, <2.0.0)
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range constraint: %s", s)
		}

		var constraint RangeConstraint
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, ">=") {
				v, err := Parse(part[2:])
				if err != nil {
					return nil, err
				}
				constraint.min = v
				constraint.minInclusive = true
			} else if strings.HasPrefix(part, ">") {
				v, err := Parse(part[1:])
				if err != nil {
					return nil, err
				}
				constraint.min = v
				constraint.minInclusive = false
			} else if strings.HasPrefix(part, "<=") {
				v, err := Parse(part[2:])
				if err != nil {
					return nil, err
				}
				constraint.max = v
				constraint.maxInclusive = true
			} else if strings.HasPrefix(part, "<") {
				v, err := Parse(part[1:])
				if err != nil {
					return nil, err
				}
				constraint.max = v
				constraint.maxInclusive = false
			}
		}
		return &constraint, nil
	}

	// Single comparison (>=1.0.0)
	if strings.HasPrefix(s, ">=") {
		v, err := Parse(s[2:])
		if err != nil {
			return nil, err
		}
		return &RangeConstraint{
			min:          v,
			minInclusive: true,
		}, nil
	}

	// Exact version
	v, err := Parse(s)
	if err != nil {
		return nil, err
	}
	return &ExactConstraint{version: v}, nil
}
