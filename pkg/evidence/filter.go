package evidence

import "time"

// FilterFunc is a predicate that returns true if the given Item should be
// included in a result set.
type FilterFunc func(item Item) bool

// FilterOptions groups the common query parameters used when selecting items
// from a Collector.
type FilterOptions struct {
	// Tags, if non-empty, restricts results to items that carry ALL of the
	// listed tags.
	Tags []string

	// Since, if non-zero, restricts results to items collected at or after
	// this time.
	Since time.Time

	// Until, if non-zero, restricts results to items collected before this
	// time.
	Until time.Time

	// Kind, if non-empty, restricts results to items whose Kind field matches
	// exactly.
	Kind string
}

// BuildFilter converts a FilterOptions value into a single FilterFunc that
// can be passed to Collector.Filter.
func BuildFilter(opts FilterOptions) FilterFunc {
	return func(item Item) bool {
		// Kind match
		if opts.Kind != "" && item.Kind != opts.Kind {
			return false
		}

		// Time-window checks
		if !opts.Since.IsZero() && item.CollectedAt.Before(opts.Since) {
			return false
		}
		// Until is exclusive: items collected exactly at Until are excluded.
		if !opts.Until.IsZero() && !item.CollectedAt.Before(opts.Until) {
			return false
		}

		// Tag subset check — every required tag must be present on the item.
		if len(opts.Tags) > 0 {
			tagSet := make(map[string]struct{}, len(item.Tags))
			for _, t := range item.Tags {
				tagSet[t] = struct{}{}
			}
			for _, required := range opts.Tags {
				if _, ok := tagSet[required]; !ok {
					return false
				}
			}
		}

		return true
	}
}

// Filter returns the subset of items for which fn returns true.
// It operates on a snapshot so callers do not need to hold any lock.
func (c *Collector) Filter(fn FilterFunc) []Item {
	snap := c.All()
	// Pre-allocate with a smaller capacity hint since most filters are selective.
	result := make([]Item, 0, len(snap)/2+1)
	for _, item := range snap {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

// FilterByOptions is a convenience wrapper around BuildFilter + Filter.
func (c *Collector) FilterByOptions(opts FilterOptions) []Item {
	return c.Filter(BuildFilter(opts))
}
