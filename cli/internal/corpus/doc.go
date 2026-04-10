// Package corpus computes corpus-quality fitness vectors for the Dream
// nightly compounder's MEASURE stage.
//
// This package deliberately lives OUTSIDE the goals subsystem. Directives
// in cli/internal/goals are human strategic records in GOALS.md; fitness
// metrics here are computed probes. Overloading the Directive type
// would create a schema split — see the plan's pm-002 finding for the
// full rationale.
package corpus
