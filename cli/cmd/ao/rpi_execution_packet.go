package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/boshu2/agentops/cli/internal/autodev"
)

type executionPacket struct {
	SchemaVersion      int                     `json:"schema_version"`
	Objective          string                  `json:"objective"`
	EpicID             string                  `json:"epic_id,omitempty"`
	PlanPath           string                  `json:"plan_path,omitempty"`
	ContractSurfaces   []string                `json:"contract_surfaces"`
	ValidationCommands []string                `json:"validation_commands,omitempty"`
	TrackerMode        string                  `json:"tracker_mode"`
	TrackerHealth      *trackerHealth          `json:"tracker_health,omitempty"`
	DoneCriteria       []string                `json:"done_criteria,omitempty"`
	Complexity         string                  `json:"complexity,omitempty"`
	AutodevProgram     *executionPacketProgram `json:"autodev_program,omitempty"`
}

type executionPacketProgram struct {
	Path               string   `json:"path"`
	MutableScope       []string `json:"mutable_scope,omitempty"`
	ImmutableScope     []string `json:"immutable_scope,omitempty"`
	ExperimentUnit     string   `json:"experiment_unit,omitempty"`
	ValidationCommands []string `json:"validation_commands,omitempty"`
	DecisionPolicy     []string `json:"decision_policy,omitempty"`
	StopConditions     []string `json:"stop_conditions,omitempty"`
}

func writeExecutionPacketSeed(cwd string, state *phasedState) error {
	tracker := detectTrackerHealth(state.Opts.BDCommand)
	packet := executionPacket{
		SchemaVersion:    1,
		Objective:        state.Goal,
		EpicID:           state.EpicID,
		ContractSurfaces: []string{},
		TrackerMode:      tracker.Mode,
		TrackerHealth:    &tracker,
		Complexity:       string(state.Complexity),
	}
	if isPlanFileEpic(state.EpicID) {
		packet.PlanPath = planFileFromEpic(state.EpicID)
	} else if planPath, err := discoverPlanFile(cwd); err == nil {
		packet.PlanPath = planPath
	}

	if _, err := os.Stat(filepath.Join(cwd, "docs", "contracts", "repo-execution-profile.md")); err == nil {
		packet.ContractSurfaces = append(packet.ContractSurfaces, "docs/contracts/repo-execution-profile.md")
	}

	if state.ProgramPath != "" {
		packet.ContractSurfaces = append(packet.ContractSurfaces, state.ProgramPath)
		prog, _, err := autodev.LoadProgram(filepath.Join(cwd, state.ProgramPath))
		if err != nil {
			return fmt.Errorf("load %s for execution packet: %w", state.ProgramPath, err)
		}
		packet.ValidationCommands = append(packet.ValidationCommands, prog.ValidationCommands...)
		packet.DoneCriteria = append(packet.DoneCriteria, prog.StopConditions...)
		packet.AutodevProgram = &executionPacketProgram{
			Path:               state.ProgramPath,
			MutableScope:       prog.MutableScope,
			ImmutableScope:     prog.ImmutableScope,
			ExperimentUnit:     prog.ExperimentUnit,
			ValidationCommands: prog.ValidationCommands,
			DecisionPolicy:     prog.DecisionPolicy,
			StopConditions:     prog.StopConditions,
		}
	}

	stateDir := filepath.Join(cwd, ".agents", "rpi")
	if err := os.MkdirAll(stateDir, 0o750); err != nil {
		return fmt.Errorf("create execution packet directory: %w", err)
	}
	data, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal execution packet: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(stateDir, "execution-packet.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write execution packet: %w", err)
	}
	return nil
}
