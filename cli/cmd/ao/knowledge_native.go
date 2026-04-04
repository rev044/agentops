package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type knowledgeTopicDetail struct {
	knowledgeTopicState
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

type knowledgePromotedPacketState struct {
	TopicID       string
	Path          string
	PrimaryClaims []string
}

type knowledgeChunkState struct {
	ID         string
	Type       string
	Confidence string
	Claim      string
}

type knowledgeChunkBundleState struct {
	TopicID            string
	Title              string
	Path               string
	PromotedPacketPath string
	Chunks             []knowledgeChunkState
}

type knowledgeNativeBuildResult struct {
	OutputPath string
	Metadata   map[string]string
	Output     string
}

type knowledgeBriefEvidence struct {
	TopicID string
	ChunkID string
	Claim   string
}

func runKnowledgeNativeBuilder(workspace, agentsRoot string, step knowledgeBuilderInvocation) (knowledgeBuilderRun, error) {
	run := knowledgeBuilderRun{knowledgeBuilderInvocation: step}

	var (
		result knowledgeNativeBuildResult
		err    error
	)

	switch step.Step {
	case "belief-book":
		result, err = buildKnowledgeBeliefBook(agentsRoot)
	case "playbooks":
		result, err = buildKnowledgePlaybooks(agentsRoot, slicesContain(step.Args, "--include-thin"))
	case "briefing":
		goal := knowledgeBuilderGoal(step.Args)
		if goal == "" {
			return run, fmt.Errorf("briefing builder requires --goal")
		}
		result, err = buildKnowledgeBriefing(agentsRoot, goal)
	default:
		return run, fmt.Errorf("unsupported native knowledge builder step: %s", step.Step)
	}
	if err != nil {
		return run, err
	}

	run.Path = result.OutputPath
	run.Metadata = result.Metadata
	run.Output = strings.TrimSpace(result.Output)
	if run.Output == "" && result.OutputPath != "" {
		run.Output = result.OutputPath
	}
	return run, nil
}

func buildKnowledgeBeliefBook(agentsRoot string) (knowledgeNativeBuildResult, error) {
	topics := loadKnowledgeTopicDetails(agentsRoot)
	if len(topics) == 0 {
		return knowledgeNativeBuildResult{}, fmt.Errorf("knowledge beliefs requires topic packets under %s", filepath.Join(agentsRoot, "topics"))
	}

	coreBeliefs, operatingPrinciples := collectKnowledgeBeliefSections(topics, agentsRoot)
	thinTopics := make([]string, 0, len(topics))
	for _, topic := range topics {
		if topic.Health != "healthy" {
			thinTopics = append(thinTopics, topic.Title)
		}
	}

	outputPath := filepath.Join(agentsRoot, "knowledge", "book-of-beliefs.md")
	sources := knowledgeExistingPaths(
		filepath.Join(agentsRoot, "packets", "index.md"),
		filepath.Join(agentsRoot, "packets", "chunks", "index.md"),
		filepath.Join(agentsRoot, "topics", "index.md"),
	)
	sourcePath := ""
	if len(sources) > 0 {
		sourcePath = sources[0]
	}
	content := renderKnowledgeBeliefBook(outputPath, sourcePath, coreBeliefs, operatingPrinciples, thinTopics, sources)
	if err := writeKnowledgeOutput(outputPath, content); err != nil {
		return knowledgeNativeBuildResult{}, err
	}

	return knowledgeNativeBuildResult{
		OutputPath: outputPath,
		Metadata:   map[string]string{"belief_book": outputPath},
		Output:     fmt.Sprintf("belief_book=%s", outputPath),
	}, nil
}

func buildKnowledgePlaybooks(agentsRoot string, includeThin bool) (knowledgeNativeBuildResult, error) {
	topics := loadKnowledgeTopicDetails(agentsRoot)
	if len(topics) == 0 {
		return knowledgeNativeBuildResult{}, fmt.Errorf("knowledge playbooks requires topic packets under %s", filepath.Join(agentsRoot, "topics"))
	}

	selected := make([]knowledgeTopicDetail, 0, len(topics))
	for _, topic := range topics {
		if includeThin || topic.Health == "healthy" {
			selected = append(selected, topic)
		}
	}
	if len(selected) == 0 {
		return knowledgeNativeBuildResult{}, fmt.Errorf("knowledge playbooks found no topics eligible for promotion")
	}

	playbooksDir := filepath.Join(agentsRoot, "playbooks")
	if err := pruneKnowledgeMarkdown(playbooksDir, "index.md", "README.md"); err != nil {
		return knowledgeNativeBuildResult{}, err
	}

	rows := make([]knowledgePlaybookRow, 0, len(selected))
	for _, topic := range selected {
		outputPath := filepath.Join(playbooksDir, topic.ID+".md")
		content := renderKnowledgePlaybook(topic, agentsRoot)
		if err := writeKnowledgeOutput(outputPath, content); err != nil {
			return knowledgeNativeBuildResult{}, err
		}
		rows = append(rows, knowledgePlaybookRow{
			Topic:     topic.Title,
			Path:      filepath.Base(outputPath),
			Health:    topic.Health,
			Canonical: topic.Health == "healthy" && knowledgePathExists(filepath.Join(agentsRoot, "packets", "promoted", topic.ID+".md")),
		})
	}

	indexPath := filepath.Join(playbooksDir, "index.md")
	if err := writeKnowledgeOutput(indexPath, renderKnowledgePlaybooksIndex(rows)); err != nil {
		return knowledgeNativeBuildResult{}, err
	}

	return knowledgeNativeBuildResult{
		OutputPath: indexPath,
		Metadata:   map[string]string{"playbooks": fmt.Sprintf("%d", len(rows))},
		Output:     fmt.Sprintf("playbooks=%d", len(rows)),
	}, nil
}

func buildKnowledgeBriefing(agentsRoot, goal string) (knowledgeNativeBuildResult, error) {
	topics := loadKnowledgeTopicDetails(agentsRoot)
	if len(topics) == 0 {
		return knowledgeNativeBuildResult{}, fmt.Errorf("knowledge brief requires topic packets under %s", filepath.Join(agentsRoot, "topics"))
	}

	candidates := selectRelevantKnowledgeTopics(goal, topics, agentsRoot, 6)
	selected := make([]knowledgeTopicDetail, 0, 3)
	suppressedThin := make([]string, 0, len(candidates))
	for _, topic := range candidates {
		if topic.Health != "healthy" {
			suppressedThin = append(suppressedThin, topic.ID)
			continue
		}
		selected = append(selected, topic)
		if len(selected) == 3 {
			break
		}
	}
	if len(selected) == 0 {
		for _, topic := range topics {
			if topic.Health != "healthy" || containsKnowledgeTopic(selected, topic.ID) {
				continue
			}
			selected = append(selected, topic)
			if len(selected) == 3 {
				break
			}
		}
	}
	if len(selected) == 0 {
		return knowledgeNativeBuildResult{}, fmt.Errorf("knowledge brief could not select relevant topics for %q", goal)
	}

	coreBeliefs, _ := collectKnowledgeBeliefSections(topics, agentsRoot)
	if len(coreBeliefs) > 3 {
		coreBeliefs = coreBeliefs[:3]
	}

	evidence := collectKnowledgeBriefEvidence(agentsRoot, selected, 6)
	warnings := make([]string, 0, len(selected)+2)
	sourceSurfaces := make([]string, 0, len(selected)*3)
	playbookPath := ""
	for _, topic := range selected {
		sourceSurfaces = append(sourceSurfaces, topic.Path)
		chunksPath := filepath.Join(agentsRoot, "packets", "chunks", topic.ID+".md")
		promotedPath := filepath.Join(agentsRoot, "packets", "promoted", topic.ID+".md")
		playbookCandidate := filepath.Join(agentsRoot, "playbooks", topic.ID+".md")
		if knowledgePathExists(chunksPath) {
			sourceSurfaces = append(sourceSurfaces, chunksPath)
		}
		if knowledgePathExists(promotedPath) {
			sourceSurfaces = append(sourceSurfaces, promotedPath)
		}
		if playbookPath == "" && knowledgePathExists(playbookCandidate) {
			playbookPath = playbookCandidate
			sourceSurfaces = append(sourceSurfaces, playbookCandidate)
		}
	}
	if knowledgePathExists(filepath.Join(agentsRoot, "knowledge", "operator-model.md")) {
		sourceSurfaces = append(sourceSurfaces, filepath.Join(agentsRoot, "knowledge", "operator-model.md"))
	}
	if knowledgePathExists(filepath.Join(agentsRoot, "knowledge", "book-of-beliefs.md")) {
		sourceSurfaces = append(sourceSurfaces, filepath.Join(agentsRoot, "knowledge", "book-of-beliefs.md"))
	}
	for _, topicID := range dedupeKnowledgeStrings(suppressedThin) {
		warnings = append(warnings, fmt.Sprintf("`%s` matched the goal but is still thin; it stays discovery-only and was not injected into the control plane.", topicID))
	}
	if warning := knowledgeSourceManifestWarning(agentsRoot); warning != "" {
		warnings = append(warnings, warning)
	}
	sourceSurfaces = dedupeKnowledgeStrings(sourceSurfaces)

	datePrefix := time.Now().Format("2006-01-02")
	slug := slugify(goal)
	if slug == "" {
		slug = "briefing"
	}
	outputPath := filepath.Join(agentsRoot, "briefings", fmt.Sprintf("%s-%s.md", datePrefix, slug))
	content := renderKnowledgeBriefing(goal, selected, coreBeliefs, evidence, warnings, sourceSurfaces, playbookPath, filepath.Join(agentsRoot, "knowledge", "operator-model.md"))
	if err := writeKnowledgeOutput(outputPath, content); err != nil {
		return knowledgeNativeBuildResult{}, err
	}

	return knowledgeNativeBuildResult{
		OutputPath: outputPath,
		Metadata:   map[string]string{"briefing": outputPath},
		Output:     fmt.Sprintf("briefing=%s", outputPath),
	}, nil
}

func loadKnowledgeTopicDetails(agentsRoot string) []knowledgeTopicDetail {
	topicsDir := filepath.Join(agentsRoot, "topics")
	entries, err := os.ReadDir(topicsDir)
	if err != nil {
		return nil
	}

	topics := make([]knowledgeTopicDetail, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		if entry.Name() == "index.md" || entry.Name() == "README.md" {
			continue
		}

		path := filepath.Join(topicsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		text := string(data)
		frontmatter := parseKnowledgeFrontmatter(text)

		topic := knowledgeTopicDetail{
			knowledgeTopicState: knowledgeTopicState{
				ID:     knowledgeFrontmatterString(frontmatter, "topic_id", strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))),
				Title:  knowledgeFrontmatterString(frontmatter, "title", strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))),
				Health: knowledgeFrontmatterString(frontmatter, "health_state", "thin"),
				Path:   path,
			},
			Summary:          knowledgeSectionText(text, "## Summary"),
			Consumers:        dedupeKnowledgeStrings(append(knowledgeFrontmatterStringList(frontmatter, "consumer_surfaces"), extractKnowledgeBullets(text, "## Consumers")...)),
			Aliases:          knowledgeFrontmatterStringList(frontmatter, "aliases"),
			QuerySeeds:       knowledgeFrontmatterStringList(frontmatter, "query_seeds"),
			KeyDecisions:     extractKnowledgeBullets(text, "## Key Decisions"),
			RepeatedPatterns: extractKnowledgeBullets(text, "## Repeated Patterns"),
			Conversations:    knowledgeFrontmatterNestedInt(frontmatter, "evidence_counts", "conversations"),
			Artifacts:        knowledgeFrontmatterNestedInt(frontmatter, "evidence_counts", "artifacts"),
			VerifiedHits:     knowledgeFrontmatterNestedInt(frontmatter, "evidence_counts", "verified_hits"),
		}
		topic.OpenGaps = filterKnowledgeOpenGaps(extractKnowledgeBullets(text, "## Open Gaps"))
		topics = append(topics, topic)
	}

	sort.Slice(topics, func(i, j int) bool {
		if topics[i].Health != topics[j].Health {
			return knowledgeHealthRank(topics[i].Health) > knowledgeHealthRank(topics[j].Health)
		}
		return topics[i].Title < topics[j].Title
	})
	return topics
}

func collectKnowledgeBeliefSections(topics []knowledgeTopicDetail, agentsRoot string) ([]string, []string) {
	coreBeliefs := []string{
		"Build the environment, not the actor; durable intelligence lives in state, traces, and gates.",
		"Replaceable actors should coordinate through stigmergic traces and provenance, not hidden conversational state.",
		"Promote only what survives explicit selection gates and improves future execution.",
	}
	operatingPrinciples := []string{
		"Startup context is budgeted: a few beliefs, one healthy playbook, concrete warnings, and source links.",
		"Thin topics stay discovery-only until they survive verification and promotion.",
		"When CAS freshness is unhealthy, file-backed artifacts and lexical probes are authoritative.",
	}
	for _, topic := range topics {
		if topic.Health != "healthy" {
			continue
		}

		promoted := loadKnowledgePromotedPacket(agentsRoot, topic.ID)
		for _, claim := range promoted.PrimaryClaims {
			coreBeliefs = appendKnowledgeCandidate(coreBeliefs, claim)
		}
		for _, decision := range topic.KeyDecisions {
			coreBeliefs = appendKnowledgeCandidate(coreBeliefs, decision)
		}
		for _, pattern := range topic.RepeatedPatterns {
			operatingPrinciples = appendKnowledgeCandidate(operatingPrinciples, pattern)
		}
	}

	if len(coreBeliefs) == 0 {
		coreBeliefs = appendKnowledgeCandidate(coreBeliefs, "Promoted topic packets have not surfaced stable beliefs yet.")
	}
	if len(operatingPrinciples) == 0 {
		operatingPrinciples = appendKnowledgeCandidate(operatingPrinciples, "Generated operator surfaces should stay citation-backed and grounded in topic packets.")
	}

	if len(coreBeliefs) > 8 {
		coreBeliefs = coreBeliefs[:8]
	}
	if len(operatingPrinciples) > 7 {
		operatingPrinciples = operatingPrinciples[:7]
	}
	return coreBeliefs, operatingPrinciples
}

func loadKnowledgePromotedPacket(agentsRoot, topicID string) knowledgePromotedPacketState {
	path := filepath.Join(agentsRoot, "packets", "promoted", topicID+".md")
	if !knowledgePathExists(path) {
		return knowledgePromotedPacketState{}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return knowledgePromotedPacketState{}
	}
	text := string(data)
	frontmatter := parseKnowledgeFrontmatter(text)
	return knowledgePromotedPacketState{
		TopicID:       knowledgeFrontmatterString(frontmatter, "source_topic", topicID),
		Path:          path,
		PrimaryClaims: extractKnowledgeBullets(text, "## Primary Claims"),
	}
}

func selectRelevantKnowledgeTopics(goal string, topics []knowledgeTopicDetail, agentsRoot string, limit int) []knowledgeTopicDetail {
	type scoredTopic struct {
		topic knowledgeTopicDetail
		score int
	}

	goalLower := strings.ToLower(goal)
	goalTokens := knowledgeTokens(goal)
	scored := make([]scoredTopic, 0, len(topics))
	for _, topic := range topics {
		score := 0
		if strings.Contains(goalLower, strings.ToLower(topic.ID)) || strings.Contains(goalLower, strings.ToLower(topic.Title)) {
			score += 8
		}

		combined := strings.ToLower(strings.Join(append(append([]string{topic.Title, topic.Summary}, topic.Aliases...), topic.QuerySeeds...), " "))
		for _, token := range goalTokens {
			if strings.Contains(combined, token) {
				score += 2
			}
		}
		if topic.Health == "healthy" {
			score += 2
		}
		if knowledgePathExists(filepath.Join(agentsRoot, "packets", "promoted", topic.ID+".md")) {
			score++
		}
		scored = append(scored, scoredTopic{topic: topic, score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		if scored[i].topic.Health != scored[j].topic.Health {
			return knowledgeHealthRank(scored[i].topic.Health) > knowledgeHealthRank(scored[j].topic.Health)
		}
		return scored[i].topic.Title < scored[j].topic.Title
	})

	selected := make([]knowledgeTopicDetail, 0, limit)
	for _, candidate := range scored {
		if candidate.score == 0 && len(selected) > 0 {
			break
		}
		selected = append(selected, candidate.topic)
		if len(selected) == limit {
			return selected
		}
	}

	for _, candidate := range scored {
		if len(selected) == limit {
			break
		}
		if containsKnowledgeTopic(selected, candidate.topic.ID) {
			continue
		}
		selected = append(selected, candidate.topic)
	}
	return selected
}

func collectKnowledgeBriefEvidence(agentsRoot string, topics []knowledgeTopicDetail, maxItems int) []knowledgeBriefEvidence {
	evidence := make([]knowledgeBriefEvidence, 0, maxItems)
	for _, topic := range topics {
		if len(evidence) == maxItems {
			break
		}

		bundle := loadKnowledgeChunkBundle(agentsRoot, topic.ID)
		if len(bundle.Chunks) == 0 {
			fallbackID := topic.ID + "-overview"
			evidence = append(evidence, knowledgeBriefEvidence{TopicID: topic.ID, ChunkID: fallbackID, Claim: topic.Summary})
			continue
		}

		chunks := bundle.Chunks
		sort.SliceStable(chunks, func(i, j int) bool {
			if knowledgeChunkRank(chunks[i].Type) != knowledgeChunkRank(chunks[j].Type) {
				return knowledgeChunkRank(chunks[i].Type) > knowledgeChunkRank(chunks[j].Type)
			}
			return chunks[i].ID < chunks[j].ID
		})

		added := 0
		for _, chunk := range chunks {
			if strings.TrimSpace(chunk.Claim) == "" {
				continue
			}
			evidence = append(evidence, knowledgeBriefEvidence{
				TopicID: topic.ID,
				ChunkID: chunk.ID,
				Claim:   chunk.Claim,
			})
			added++
			if len(evidence) == maxItems || added == 2 {
				break
			}
		}
	}
	return evidence
}

func loadKnowledgeChunkBundle(agentsRoot, topicID string) knowledgeChunkBundleState {
	path := filepath.Join(agentsRoot, "packets", "chunks", topicID+".md")
	if !knowledgePathExists(path) {
		return knowledgeChunkBundleState{}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return knowledgeChunkBundleState{}
	}
	text := string(data)
	frontmatter := parseKnowledgeFrontmatter(text)
	return knowledgeChunkBundleState{
		TopicID:            knowledgeFrontmatterString(frontmatter, "topic_id", topicID),
		Title:              knowledgeFrontmatterString(frontmatter, "title", topicID),
		Path:               path,
		PromotedPacketPath: knowledgeFrontmatterString(frontmatter, "promoted_packet_path", ""),
		Chunks:             parseKnowledgeChunks(text),
	}
}

func parseKnowledgeChunks(text string) []knowledgeChunkState {
	lines := strings.Split(text, "\n")
	chunks := make([]knowledgeChunkState, 0, 8)
	inSection := false
	var current *knowledgeChunkState

	flush := func() {
		if current == nil {
			return
		}
		if strings.TrimSpace(current.ID) != "" || strings.TrimSpace(current.Claim) != "" {
			chunks = append(chunks, *current)
		}
		current = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "## Knowledge Chunks":
			inSection = true
		case !inSection:
			continue
		case strings.HasPrefix(trimmed, "## "):
			flush()
			return chunks
		case strings.HasPrefix(trimmed, "### "):
			flush()
			current = &knowledgeChunkState{}
		case current == nil:
			continue
		case strings.HasPrefix(trimmed, "- Chunk ID:"):
			current.ID = knowledgeFieldValue(trimmed)
		case strings.HasPrefix(trimmed, "- Type:"):
			current.Type = knowledgeFieldValue(trimmed)
		case strings.HasPrefix(trimmed, "- Confidence:"):
			current.Confidence = knowledgeFieldValue(trimmed)
		case strings.HasPrefix(trimmed, "- Claim:"):
			current.Claim = knowledgeFieldValue(trimmed)
		}
	}
	flush()
	return chunks
}

func renderKnowledgeBeliefBook(outputPath, sourcePath string, coreBeliefs, operatingPrinciples, thinTopics, sourceSurfaces []string) string {
	datePrefix := time.Now().Format("2006-01-02")
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("id: knowledge-book-of-beliefs-%s\n", datePrefix))
	b.WriteString("type: principle-book\n")
	b.WriteString(fmt.Sprintf("date: %s\n", datePrefix))
	if sourcePath != "" {
		b.WriteString(fmt.Sprintf("source: %s\n", sourcePath))
	}
	b.WriteString("status: generated\n")
	b.WriteString("---\n\n")
	b.WriteString("# Book Of Beliefs\n\n")
	b.WriteString("Cross-domain operating beliefs promoted from the `.agents` corpus.\n\n")
	b.WriteString("## Canonical Vocabulary\n\n")
	for _, primitive := range []string{
		"fitness gradient",
		"stateful environment",
		"replaceable actors",
		"stigmergic traces",
		"selection gates",
		"evolutionary promotion",
		"governance",
	} {
		b.WriteString(fmt.Sprintf("- %s\n", primitive))
	}
	b.WriteString("## Core Beliefs\n\n")
	for idx, belief := range coreBeliefs {
		b.WriteString(fmt.Sprintf("%d. %s\n", idx+1, belief))
	}
	if len(coreBeliefs) == 0 {
		b.WriteString("1. Promoted topic packets have not surfaced stable beliefs yet.\n")
	}
	b.WriteString("\n## Operating Principles\n\n")
	for _, principle := range operatingPrinciples {
		b.WriteString(fmt.Sprintf("- %s\n", principle))
	}
	if len(operatingPrinciples) == 0 {
		b.WriteString("- No operating principles surfaced from the current topic packets.\n")
	}
	b.WriteString("\n## Translation Map\n\n")
	for _, item := range []string{
		"context is the control plane -> stateful environment + governance",
		"distributed cognition -> replaceable actors coordinating through a shared environment",
		"stigmergy -> stigmergic traces in the environment",
		"flywheel -> evolutionary promotion after selection gates",
	} {
		b.WriteString(fmt.Sprintf("- %s\n", item))
	}
	b.WriteString("\n## Thin-Topic Cautions\n\n")
	if len(thinTopics) == 0 {
		b.WriteString("- None surfaced\n")
	} else {
		for _, title := range thinTopics {
			b.WriteString(fmt.Sprintf("- %s\n", title))
		}
	}
	b.WriteString("\n## Source Surfaces\n\n")
	for _, source := range sourceSurfaces {
		b.WriteString(fmt.Sprintf("- `%s`\n", source))
	}
	if len(sourceSurfaces) == 0 {
		b.WriteString("- No index surfaces were found.\n")
	}
	b.WriteString("\n## Refresh Command\n\n")
	b.WriteString("`ao knowledge beliefs`\n")
	return b.String()
}

func knowledgeSourceManifestWarning(agentsRoot string) string {
	indexPath := filepath.Join(agentsRoot, "packets", "source-manifests", "index.md")
	if !knowledgePathExists(indexPath) {
		return ""
	}
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return ""
	}
	text := string(data)
	if detail := knowledgeSectionText(text, "## Refresh Warning"); strings.TrimSpace(detail) != "" {
		return "CAS freshness warning: " + detail
	}
	if strings.Contains(text, "CASS stale at run start: `true`") || strings.Contains(text, "CASS refresh succeeded: `false`") {
		return "CAS freshness was unhealthy during packet refresh; treat file-backed artifacts and lexical probes as authoritative."
	}
	return ""
}

type knowledgePlaybookRow struct {
	Topic     string
	Path      string
	Health    string
	Canonical bool
}

func knowledgePrimitiveDescriptions() map[string]string {
	return map[string]string{
		"fitness gradient":       "Defines what better and worse look like for this topic.",
		"stateful environment":   "Captures the durable context, artifacts, and rules that carry continuity.",
		"replaceable actors":     "Keeps execution bound to narrow, swappable workers instead of one special actor.",
		"stigmergic traces":      "Uses durable traces such as packets, citations, logs, and handoffs for coordination.",
		"selection gates":        "Applies explicit checks that decide what is allowed to survive.",
		"evolutionary promotion": "Promotes validated patterns back into reusable defaults.",
		"governance":             "Shapes boundaries, ownership, and allowed moves for the operator loop.",
	}
}

func knowledgePrimitivesForTopic(topic knowledgeTopicDetail) []string {
	corpus := strings.ToLower(strings.Join(append(append(append([]string{topic.Title, topic.Summary}, topic.Aliases...), topic.KeyDecisions...), topic.RepeatedPatterns...), " "))
	keywords := map[string][]string{
		"fitness gradient":       {"goal", "fitness", "validation", "acceptance", "test", "review", "gate"},
		"stateful environment":   {"context", "memory", "environment", "packet", "state", "control plane", "knowledge"},
		"replaceable actors":     {"actor", "agent", "worker", "handoff", "owner", "ownership", "swarm"},
		"stigmergic traces":      {"trace", "provenance", "citation", "handoff", "log", "queue", "artifact"},
		"selection gates":        {"gate", "validation", "check", "policy", "proof", "review", "pre-mortem"},
		"evolutionary promotion": {"promotion", "promote", "retro", "learning", "flywheel", "reuse", "playbook"},
		"governance":             {"governance", "scope", "boundary", "operator", "policy", "constraint"},
	}
	type scored struct {
		primitive string
		score     int
	}
	var scoredPrimitives []scored
	for primitive, hints := range keywords {
		score := 0
		for _, hint := range hints {
			if strings.Contains(corpus, hint) {
				score++
			}
		}
		if score > 0 {
			scoredPrimitives = append(scoredPrimitives, scored{primitive: primitive, score: score})
		}
	}
	sort.Slice(scoredPrimitives, func(i, j int) bool {
		if scoredPrimitives[i].score != scoredPrimitives[j].score {
			return scoredPrimitives[i].score > scoredPrimitives[j].score
		}
		return scoredPrimitives[i].primitive < scoredPrimitives[j].primitive
	})
	selected := make([]string, 0, 3)
	selected = append(selected, "stateful environment")
	for _, item := range scoredPrimitives {
		if stringSliceContainsFold(selected, item.primitive) {
			continue
		}
		selected = append(selected, item.primitive)
		if len(selected) == 3 {
			break
		}
	}
	if len(selected) < 2 {
		selected = append(selected, "selection gates")
	}
	return dedupeKnowledgeStrings(selected)
}

func renderKnowledgePlaybook(topic knowledgeTopicDetail, agentsRoot string) string {
	var b strings.Builder
	chunksPath := filepath.Join(agentsRoot, "packets", "chunks", topic.ID+".md")
	promotedPath := filepath.Join(agentsRoot, "packets", "promoted", topic.ID+".md")
	rules := make([]string, 0, len(topic.KeyDecisions)+len(topic.RepeatedPatterns))
	rules = append(rules, topic.KeyDecisions...)
	rules = append(rules, topic.RepeatedPatterns...)
	rules = dedupeKnowledgeStrings(rules)
	if len(rules) > 6 {
		rules = rules[:6]
	}
	primitives := knowledgePrimitivesForTopic(topic)
	primitiveDescriptions := knowledgePrimitiveDescriptions()

	b.WriteString(fmt.Sprintf("# Playbook Candidate: %s\n\n", topic.Title))
	b.WriteString("## When To Use\n\n")
	b.WriteString(knowledgeWhenToUse(topic))
	b.WriteString("\n\n## Summary\n\n")
	if topic.Summary != "" {
		b.WriteString(topic.Summary)
	} else {
		b.WriteString(fmt.Sprintf("%s has an eligible topic packet but no summary text yet.", topic.Title))
	}
	b.WriteString("\n\n## Primitives Covered\n\n")
	for _, primitive := range primitives {
		b.WriteString(fmt.Sprintf("- `%s`: %s\n", primitive, primitiveDescriptions[primitive]))
	}
	b.WriteString("\n## Operator Policy\n\n")
	if len(rules) == 0 {
		b.WriteString("- No durable operator policy surfaced yet.\n")
	} else {
		for _, rule := range rules {
			b.WriteString(fmt.Sprintf("- %s\n", rule))
		}
	}
	b.WriteString("\n## Stigmergic Traces\n\n")
	b.WriteString(fmt.Sprintf("- Topic packet: `%s`\n", topic.Path))
	b.WriteString(fmt.Sprintf("- Chunk bundle: `%s`\n", chunksPath))
	b.WriteString(fmt.Sprintf("- Promoted packet: `%s`\n", promotedPath))
	b.WriteString("\n## Selection Gates\n\n")
	b.WriteString("- Validate the chosen path against current repo or workspace reality before promotion.\n")
	b.WriteString("- Keep thin topics discovery-only; only healthy topics and promoted packets may define default behavior.\n")
	b.WriteString("- Prefer planning rules, pre-mortem checks, and direct artifact lineage over conversational confidence.\n")
	b.WriteString("\n## Trust Status\n\n")
	b.WriteString(fmt.Sprintf("- Topic health: `%s`\n", topic.Health))
	b.WriteString(fmt.Sprintf("- Promoted packet present: `%s`\n", yesNo(knowledgePathExists(promotedPath))))
	if topic.Health != "healthy" {
		b.WriteString("- Treat this playbook as non-canonical until the topic health improves.\n")
	}
	b.WriteString("\n## Source Surfaces\n\n")
	b.WriteString(fmt.Sprintf("- `%s`\n", topic.Path))
	if knowledgePathExists(chunksPath) {
		b.WriteString(fmt.Sprintf("- `%s`\n", chunksPath))
	}
	if knowledgePathExists(promotedPath) {
		b.WriteString(fmt.Sprintf("- `%s`\n", promotedPath))
	}
	return b.String()
}

func renderKnowledgePlaybooksIndex(rows []knowledgePlaybookRow) string {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Health != rows[j].Health {
			return knowledgeHealthRank(rows[i].Health) > knowledgeHealthRank(rows[j].Health)
		}
		return rows[i].Topic < rows[j].Topic
	})

	var b strings.Builder
	b.WriteString("# Playbook Candidates\n\n")
	b.WriteString("Candidate operator playbooks promoted from topic packets.\n\n")
	b.WriteString("| Topic | Health | Canonical |\n")
	b.WriteString("|---|---|---|\n")
	for _, row := range rows {
		canonical := "no"
		if row.Canonical {
			canonical = "yes"
		}
		b.WriteString(fmt.Sprintf("| [%s](%s) | `%s` | `%s` |\n", row.Topic, row.Path, row.Health, canonical))
	}
	return b.String()
}

func renderKnowledgeBriefing(goal string, topics []knowledgeTopicDetail, beliefs []string, evidence []knowledgeBriefEvidence, warnings, sourceSurfaces []string, playbookPath, operatorModelPath string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Briefing: %s\n\n", goal))
	b.WriteString(fmt.Sprintf("**Date:** %s\n\n", time.Now().Format("2006-01-02")))
	b.WriteString("## Objective + Fitness Gradient\n\n")
	b.WriteString(fmt.Sprintf("- Objective: %s\n", goal))
	b.WriteString("- Fitness gradient: prefer the smallest change that improves the environment for this objective and survives explicit proof surfaces.\n")
	b.WriteString("\n## Environment State\n\n")
	if knowledgePathExists(operatorModelPath) {
		b.WriteString(fmt.Sprintf("- Canonical doctrine: `%s`\n", operatorModelPath))
	}
	for _, belief := range beliefs {
		b.WriteString(fmt.Sprintf("- Core belief: %s\n", belief))
	}
	if len(topics) == 0 {
		b.WriteString("- Healthy topics in scope: none selected from the current taxonomy.\n")
	} else {
		for _, topic := range topics {
			b.WriteString(fmt.Sprintf("- Healthy topic in scope: `%s`\n", topic.ID))
		}
	}
	b.WriteString("- Retrieval mode: file-backed operator surfaces remain authoritative; lexical truth beats semantic drift when freshness is degraded.\n")
	b.WriteString("\n## Actors / Ownership\n\n")
	b.WriteString("- Primary actor: the current Codex session operating on one objective at a time.\n")
	b.WriteString("- Replaceable actors: use the briefing for local state and healthy playbooks for behavior policy instead of conversational memory.\n")
	if strings.TrimSpace(playbookPath) != "" {
		b.WriteString(fmt.Sprintf("- Reusable behavior policy: `%s`\n", playbookPath))
	} else {
		b.WriteString("- Reusable behavior policy: no healthy playbook matched this objective yet.\n")
	}
	b.WriteString("- Surface ownership: briefing=local state, playbook=policy, planning rules=selection pressure, packets/provenance=evidence.\n")
	b.WriteString("\n## Stigmergic Traces + Source Surfaces\n\n")
	for _, item := range evidence {
		b.WriteString(fmt.Sprintf("- Trace `%s` `%s`: %s\n", item.TopicID, item.ChunkID, item.Claim))
	}
	if len(evidence) == 0 {
		b.WriteString("- Trace: no matching chunk claims were found in healthy topics.\n")
	}
	for _, source := range sourceSurfaces {
		b.WriteString(fmt.Sprintf("- Source surface: `%s`\n", source))
	}
	b.WriteString("\n## Selection Gates / Proof Surfaces\n\n")
	b.WriteString("- Gate: planning rules and pre-mortem checks outrank preference when they disagree with the current plan.\n")
	b.WriteString("- Gate: promote only changes backed by direct artifacts, promoted packets, or runnable validation.\n")
	if strings.TrimSpace(playbookPath) != "" {
		b.WriteString(fmt.Sprintf("- Proof surface: healthy playbook at `%s`\n", playbookPath))
	}
	if len(evidence) > 0 {
		b.WriteString("- Proof surface: use the chunk claims above as candidate evidence, then confirm against the linked topic or promoted packet.\n")
	}
	b.WriteString("\n## Promotion / Reuse Warnings\n\n")
	if len(warnings) == 0 {
		b.WriteString("- No promotion warnings surfaced for the selected healthy topics.\n")
	} else {
		for _, warning := range warnings {
			b.WriteString(fmt.Sprintf("- %s\n", warning))
		}
	}
	b.WriteString("\n## Refresh Command\n\n")
	b.WriteString(fmt.Sprintf("`ao knowledge brief --goal %q`\n", goal))
	return b.String()
}

func knowledgeWhenToUse(topic knowledgeTopicDetail) string {
	phrases := make([]string, 0, 2)
	for _, item := range topic.Aliases {
		phrases = append(phrases, strings.ToLower(strings.TrimSpace(item)))
		if len(phrases) == 2 {
			break
		}
	}
	if len(phrases) == 0 {
		phrases = append(phrases, strings.ToLower(topic.Title))
	}
	return fmt.Sprintf("Use this playbook when the task is primarily about %s and you need a bounded operator loop instead of freeform exploration.", strings.Join(phrases, ", "))
}

func pruneKnowledgeMarkdown(dir string, preserve ...string) error {
	if GetDryRun() {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	preserveSet := make(map[string]bool, len(preserve))
	for _, name := range preserve {
		preserveSet[name] = true
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" || preserveSet[entry.Name()] {
			continue
		}
		if err := os.Remove(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func writeKnowledgeOutput(path, content string) error {
	if GetDryRun() {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimRight(content, "\n")+"\n"), 0o644)
}

func knowledgeSectionText(text, heading string) string {
	lines := strings.Split(text, "\n")
	var section []string
	inSection := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == heading {
			inSection = true
			continue
		}
		if !inSection {
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			break
		}
		section = append(section, strings.TrimSpace(line))
	}
	return strings.Join(strings.Fields(strings.Join(section, " ")), " ")
}

func knowledgeFrontmatterStringList(frontmatter map[string]any, key string) []string {
	if frontmatter == nil {
		return nil
	}
	raw, ok := frontmatter[key]
	if !ok {
		return nil
	}

	switch typed := raw.(type) {
	case []string:
		return dedupeKnowledgeStrings(typed)
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(fmt.Sprint(item))
			if text != "" && text != "<nil>" {
				values = append(values, text)
			}
		}
		return dedupeKnowledgeStrings(values)
	default:
		text := strings.TrimSpace(fmt.Sprint(typed))
		if text == "" || text == "<nil>" {
			return nil
		}
		return []string{text}
	}
}

func knowledgeFrontmatterNestedInt(frontmatter map[string]any, parent, key string) int {
	if frontmatter == nil {
		return 0
	}
	raw, ok := frontmatter[parent]
	if !ok {
		return 0
	}
	nested, ok := raw.(map[string]any)
	if !ok {
		return 0
	}
	value, ok := nested[key]
	if !ok {
		return 0
	}

	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		var parsed int
		_, _ = fmt.Sscanf(fmt.Sprint(typed), "%d", &parsed)
		return parsed
	}
}

func knowledgeFieldValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	value := strings.TrimSpace(parts[1])
	return strings.Trim(value, "`")
}

func knowledgeBuilderGoal(args []string) string {
	for idx := 0; idx < len(args); idx++ {
		if args[idx] != "--goal" {
			continue
		}
		if idx+1 < len(args) {
			return strings.TrimSpace(args[idx+1])
		}
	}
	return ""
}

func slicesContain(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func knowledgeTokens(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) >= 3 {
			tokens = append(tokens, field)
		}
	}
	return dedupeKnowledgeStrings(tokens)
}

func containsKnowledgeTopic(topics []knowledgeTopicDetail, topicID string) bool {
	for _, topic := range topics {
		if topic.ID == topicID {
			return true
		}
	}
	return false
}

func knowledgeHealthRank(health string) int {
	switch strings.ToLower(strings.TrimSpace(health)) {
	case "healthy":
		return 2
	case "thin":
		return 1
	default:
		return 0
	}
}

func knowledgeChunkRank(chunkType string) int {
	switch strings.ToLower(strings.TrimSpace(chunkType)) {
	case "decision":
		return 3
	case "pattern":
		return 2
	case "overview":
		return 1
	default:
		return 0
	}
}

func appendKnowledgeCandidate(items []string, candidate string) []string {
	candidate = strings.Join(strings.Fields(strings.TrimSpace(candidate)), " ")
	if candidate == "" {
		return items
	}
	for _, item := range items {
		if item == candidate {
			return items
		}
	}
	return append(items, candidate)
}

func knowledgeExistingPaths(paths ...string) []string {
	existing := make([]string, 0, len(paths))
	for _, path := range paths {
		if knowledgePathExists(path) {
			existing = append(existing, path)
		}
	}
	return dedupeKnowledgeStrings(existing)
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
