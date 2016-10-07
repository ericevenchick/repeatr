package runCmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/ugorji/go/codec"
	"go.polydawn.net/meep"

	"go.polydawn.net/repeatr/api/def"
	"go.polydawn.net/repeatr/api/hitch"
	"go.polydawn.net/repeatr/cmd/repeatr/bhv"
	legacy "go.polydawn.net/repeatr/core/cli"
	"go.polydawn.net/repeatr/core/executor/dispatch"
)

func Run(stdout, stderr io.Writer) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		// Parse args
		executor := executordispatch.Get(ctx.String("executor"))
		ignoreJobExit := ctx.Bool("ignore-job-exit")
		patchPaths := ctx.StringSlice("patch")
		envArgs := ctx.StringSlice("env")
		serialize := ctx.Bool("serialize")
		// One (and only one) formula should follow;
		//  we don't have a way to unambiguously output more than one result formula at the moment.
		var formulaPath string
		switch l := len(ctx.Args()); {
		case l < 1:
			panic(meep.Meep(&cmdbhv.ErrBadArgs{
				Message: "`repeatr run` requires a path to a formula as the last argument",
			}))
		case l > 1:
			panic(meep.Meep(&cmdbhv.ErrBadArgs{
				Message: "`repeatr run` requires exactly one formula as the last argument",
			}))
		case l == 1:
			formulaPath = ctx.Args()[0]
		}
		// Parse formula
		formula := hitch.LoadFormulaFromFile(formulaPath)
		// Parse patches into formulas as well.
		//  Apply each one as it's loaded.
		for _, patchPath := range patchPaths {
			formula.ApplyPatch(*hitch.LoadFormulaFromFile(patchPath))
		}
		// Any env var overrides stomp even on top of patches.
		for _, envArg := range envArgs {
			parts := strings.SplitN(envArg, "=", 2)
			if len(parts) < 2 {
				panic(meep.Meep(&cmdbhv.ErrBadArgs{
					Message: "env arguments must have an equal sign (like this: '-e KEY=val').",
				}))
			}
			formula.ApplyPatch(def.Formula{Action: def.Action{
				Env: map[string]string{parts[0]: parts[1]},
			}})
		}

		// Invoke!
		runRecord := legacy.RunFormula(executor, *formula, stdout, stderr, serialize)

		// Raise any errors that got in the way of execution.
		meep.TryPlan{
			// TODO this should filter out DataDNE, HashMismatch, etc.
			// examineCmd does a better job of this.
			// come back to this after more meep integration.
			{CatchAny: true,
				Handler: meep.TryHandlerMapto(&cmdbhv.ErrRunFailed{})},
		}.MustHandle(runRecord.Failure)

		// Output the results structure.
		//  This goes on stdout (everything is stderr) and so should be parsable.
		//  We strip some fields that aren't very useful to single-task manual runs.
		runRecord.HID = ""
		runRecord.FormulaHID = ""
		var err error
		if serialize {
			err = codec.NewEncoder(stdout, &codec.JsonHandle{}).Encode(def.Event{
				RunID:     runRecord.UID,
				RunRecord: runRecord,
			})
			stdout.Write([]byte{'\n'})
		} else {
			err = codec.NewEncoder(stdout, &codec.JsonHandle{Indent: -1}).Encode(runRecord)
			stdout.Write([]byte{'\n'})
		}
		if err != nil {
			panic(err)
		}
		// Exit nonzero with our own "your job did not report success" indicator code, if applicable.
		exitCode := runRecord.Results["$exitcode"].Hash
		if exitCode != "0" && !ignoreJobExit {
			panic(&cmdbhv.ErrExit{
				Message: fmt.Sprintf("job finished with non-zero exit status %s", exitCode),
				Code:    cmdbhv.EXIT_JOB,
			})
		}
		return nil
	}
}