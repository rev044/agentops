package bridge

import "fmt"

// FactoryRecommendedCommands returns the recommended next-step commands for the factory lane.
func FactoryRecommendedCommands(goal string) []string {
	if goal == "" {
		return []string{
			"Set a concrete goal, then run `ao factory start --goal \"your goal\"` for a briefing-first startup.",
			"Run `/rpi \"your goal\"` for the skill-first delivery lane, or `ao rpi phased \"your goal\"` for CLI-first phase isolation.",
			"Use `ao rpi status` to monitor long-running phased work.",
			"Run `ao codex stop` when the session ends so the flywheel closes explicitly.",
		}
	}

	quotedGoal := fmt.Sprintf("%q", goal)
	return []string{
		fmt.Sprintf("Run `/rpi %s` for the skill-first software-factory lane.", quotedGoal),
		fmt.Sprintf("Or run `ao rpi phased %s` for CLI-first phase isolation.", quotedGoal),
		"Use `ao rpi status` to monitor long-running phased work.",
		"Run `ao codex stop` when the session ends so the flywheel closes explicitly.",
	}
}
