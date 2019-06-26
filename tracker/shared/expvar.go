package shared

var (
	ExpvarAnnounces int64
	ExpvarScrapes   int64
	ExpvarErrs      int64

	// !x test
	ExpvarSeeds   map[string]bool
	ExpvarLeeches map[string]bool
	ExpvarIPs     map[string]bool
)
