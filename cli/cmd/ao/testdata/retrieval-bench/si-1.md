---
type: learning
maturity: candidate
confidence: high
utility: 0.85
---
# Session Intelligence Context Assembly

Ranked context assembly for session intelligence must weight recency, relevance, and source trust independently before combining scores. Session intelligence context assembly that collapses all signals into a single scalar too early loses the ability to explain why a particular artifact ranked above another. Keeping the ranked context assembly pipeline composable — with per-signal weights exposed as config — makes session intelligence tunable without code changes.
