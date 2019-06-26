package shared

var (
	ExpvarAnnounces int64
	ExpvarScrapes   int64
	ExpvarErrs      int64

	// !x test
	ExpvarSeeds   map[string]bool
	ExpvarLeeches map[string]bool
	ExpvarIPs     map[string]bool
	ExpvarPeers   map[PeerID]bool
)

// !x
func initExpvar() {
	// Might as well alloc capcity at start
	ExpvarSeeds = make(map[string]bool, 50000)
	ExpvarLeeches = make(map[string]bool, 50000)
	ExpvarIPs = make(map[string]bool, 30000)
	ExpvarPeers = make(map[PeerID]bool, 100000)
}
