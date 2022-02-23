package utils

type (
	Version struct {
		Major, Minor, Patch uint
	}

	VersionPattern struct {
		Low         Version
		High        Version
		LowExclude  bool
		HighExclude bool
	}
)

func (v Version) Equals(v2 Version) bool {
	return v.Major == v2.Major && v.Minor == v2.Minor && v.Patch == v2.Patch
}

func (v Version) CompareTo(v2 Version) int {
	if r := int(v.Major) - int(v2.Major); r != 0 {
		return r
	}
	if r := int(v.Minor) - int(v2.Minor); r != 0 {
		return r
	}
	return int(v.Patch) - int(v2.Patch)
}

func (v Version) GreaterThan(v2 Version) bool {
	return v.CompareTo(v2) > 0
}
func (v Version) GreaterEqualThan(v2 Version) bool {
	return v.CompareTo(v2) >= 0
}

func (v Version) LowerThan(v2 Version) bool {
	return v.CompareTo(v2) < 0
}
func (v Version) LowerEqualThan(v2 Version) bool {
	return v.CompareTo(v2) <= 0
}
