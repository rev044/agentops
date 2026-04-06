package rpi

import (
	"cmp"
	"slices"
)

// SelectHighestSeverityEntry picks the best item across all eligible entries.
// It returns a QueueSelection containing the winning item and its source entry
// index in next-work.jsonl. Items filtered out by repoFilter are skipped.
// Returns nil if no eligible items exist.
//
// Ranking order (descending priority):
//  1. repo affinity
//  2. freshness (not previously failed)
//  3. severity
//  4. work type rank
//  5. entry index (stable — earlier first)
//  6. item index (stable — earlier first)
func SelectHighestSeverityEntry(entries []NextWorkEntry, repoFilter string) *QueueSelection {
	type candidate struct {
		item       NextWorkItem
		entryIndex int
		itemIndex  int
		sourceEpic string
		severity   int
		affinity   int
		freshness  int
		typeRank   int
	}

	var candidates []candidate
	for _, entry := range entries {
		for itemIdx, item := range entry.Items {
			if !IsQueueItemSelectable(item) {
				continue
			}
			if repoFilter != "" && item.TargetRepo != "" && item.TargetRepo != "*" && item.TargetRepo != repoFilter {
				continue
			}
			candidates = append(candidates, candidate{
				item:       item,
				entryIndex: entry.QueueIndex,
				itemIndex:  itemIdx,
				sourceEpic: entry.SourceEpic,
				severity:   SeverityRank(item.Severity),
				affinity:   RepoAffinityRank(item, repoFilter),
				freshness:  FreshnessRank(item),
				typeRank:   WorkTypeRank(item),
			})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	slices.SortFunc(candidates, func(a, b candidate) int {
		if diff := cmp.Compare(b.affinity, a.affinity); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(b.freshness, a.freshness); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(b.severity, a.severity); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(b.typeRank, a.typeRank); diff != 0 {
			return diff
		}
		if diff := cmp.Compare(a.entryIndex, b.entryIndex); diff != 0 {
			return diff
		}
		return cmp.Compare(a.itemIndex, b.itemIndex)
	})

	best := candidates[0]
	return &QueueSelection{
		Item:       best.item,
		EntryIndex: best.entryIndex,
		ItemIndex:  best.itemIndex,
		SourceEpic: best.sourceEpic,
	}
}
