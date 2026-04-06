package knowledge

// TopicState captures the minimal topic metadata used across knowledge
// builders and gap reporting.
type TopicState struct {
	ID       string
	Title    string
	Health   string
	Path     string
	OpenGaps []string
}

// TopicDetail extends TopicState with richer content parsed from topic
// packets (summary, consumers, query seeds, decisions, patterns, counts).
type TopicDetail struct {
	TopicState
	Summary          string
	Consumers        []string
	Aliases          []string
	QuerySeeds       []string
	KeyDecisions     []string
	RepeatedPatterns []string
	Conversations    int
	Artifacts        int
	VerifiedHits     int
}

// BuilderInvocation describes a single knowledge builder step (script or
// ao-native implementation) as scheduled by `ao knowledge activate`.
type BuilderInvocation struct {
	Step           string   `json:"step"`
	Script         string   `json:"script,omitempty"`
	Implementation string   `json:"implementation,omitempty"`
	Args           []string `json:"args,omitempty"`
}

// PromotedPacketState captures data loaded from a promoted packet markdown
// file for a given topic.
type PromotedPacketState struct {
	TopicID       string
	Path          string
	PrimaryClaims []string
}

// ChunkState represents a single knowledge chunk parsed from a chunk
// bundle markdown file.
type ChunkState struct {
	ID         string
	Type       string
	Confidence string
	Claim      string
}

// ChunkBundleState captures all chunks associated with a single topic.
type ChunkBundleState struct {
	TopicID            string
	Title              string
	Path               string
	PromotedPacketPath string
	Chunks             []ChunkState
}

// NativeBuildResult is returned by ao-native knowledge builders.
type NativeBuildResult struct {
	OutputPath string
	Metadata   map[string]string
	Output     string
}

// BriefEvidence is a single chunk-level citation used in briefings.
type BriefEvidence struct {
	TopicID string
	ChunkID string
	Claim   string
}

// PlaybookRow is a single row in the playbook index.
type PlaybookRow struct {
	Topic     string
	Path      string
	Health    string
	Canonical bool
}
