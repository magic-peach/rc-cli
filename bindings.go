package rc

import (
	"github.com/runtimeconditions/extensions/tooling/extension-bindings/orchestrator"
	"github.com/spf13/cobra"
)

const goProfilerModule = "github.com/runtimeconditions/go-rc-profiler"

func bindingsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bindings <command> [args...]",
		Short: "Generate and verify extension bindings",
		Long: `Runs the bindings tool built into rc and passes all arguments to it unchanged.

The tool comes from runtimeconditions/extensions (tooling/extension-bindings/orchestrator). The Go
profiler it verifies against is the one built into rc, unless --go-profiler or RC_GO_PROFILER_BIN
names an external one.`,
		DisableFlagParsing: true,
		RunE: func(c *cobra.Command, args []string) error {
			return orchestrator.RunWith(c.Context(), args, c.OutOrStdout(), c.ErrOrStderr(), orchestrator.Hooks{ProfilerIdentity: builtinProfilerIdentity})
		},
	}
}

func builtinProfilerIdentity(language string) (orchestrator.Tool, bool, error) {
	if language != "go" {
		return orchestrator.Tool{}, false, nil
	}
	identity, err := orchestrator.ModuleIdentity(goProfilerModule, "go-rc-profiler")
	return identity, err == nil, nil
}
