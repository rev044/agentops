package rpi

// ExecutionPacketFile is the canonical filename for execution packets.
const ExecutionPacketFile = "execution-packet.json"

// ExecutionPacketProgram describes an autodev program embedded in an execution packet.
type ExecutionPacketProgram struct {
	Path               string   `json:"path"`
	MutableScope       []string `json:"mutable_scope,omitempty"`
	ImmutableScope     []string `json:"immutable_scope,omitempty"`
	ExperimentUnit     string   `json:"experiment_unit,omitempty"`
	ValidationCommands []string `json:"validation_commands,omitempty"`
	DecisionPolicy     []string `json:"decision_policy,omitempty"`
	StopConditions     []string `json:"stop_conditions,omitempty"`
}
