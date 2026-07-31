package reservation

// staff_affinity.go — TASK-021 Stage B inverse mapping helpers.
//
// Universe = clinic-scoped active + non-deleted reservation types.
// capable  = affirmative affinity SoT (staff_reservation_capabilities)
// excluded = universe \ capable  (compatibility facade shape only)

// capableIDsFromExcluded maps an exclusion set onto the capability set:
//
//	capable = universe \ excluded
//
// IDs in excluded that are not in universe are ignored (inactive/deleted types).
func capableIDsFromExcluded(universe, excluded []uint64) []uint64 {
	if len(universe) == 0 {
		return nil
	}
	excl := make(map[uint64]struct{}, len(excluded))
	for _, id := range excluded {
		excl[id] = struct{}{}
	}
	out := make([]uint64, 0, len(universe))
	for _, id := range universe {
		if _, hit := excl[id]; !hit {
			out = append(out, id)
		}
	}
	return out
}

// excludedIDsFromCapable maps a capability set onto the exclusion facade set:
//
//	excluded = universe \ capable
func excludedIDsFromCapable(universe, capable []uint64) []uint64 {
	if len(universe) == 0 {
		return nil
	}
	capSet := make(map[uint64]struct{}, len(capable))
	for _, id := range capable {
		capSet[id] = struct{}{}
	}
	out := make([]uint64, 0, len(universe))
	for _, id := range universe {
		if _, hit := capSet[id]; !hit {
			out = append(out, id)
		}
	}
	return out
}
