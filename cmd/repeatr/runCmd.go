package main

import (
	"context"
	"io"
	"sync"

	. "github.com/warpfork/go-errcat"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/repeatr"
	"go.polydawn.net/repeatr/executor/impl/memo"
	"go.polydawn.net/rio/fs"
)

func RunCmd(
	ctx context.Context,
	executorName string,
	formulaPath string,
	stdout, stderr io.Writer,
	memoDir *fs.AbsolutePath,
) (err error) {
	defer RequireErrorHasCategory(&err, repeatr.ErrorCategory(""))

	// Load formula.
	formula, formulaCtx, err := loadFormula(formulaPath)
	if err != nil {
		return err
	}

	// Run!
	printer := &ansi{stdout: stdout, stderr: stderr} // todo: switch
	_, err = Run(ctx, executorName, *formula, *formulaCtx, printer, memoDir)
	return err
}

// Run with all the I/O wiring to the terminal.
// (Not particularly reusable, except in the Batch mode, which is also
// somewhat placeholder and should later use an exec boundary and API.)
func Run(
	ctx context.Context,
	executorName string,
	formula api.Formula,
	formulaCtx api.FormulaContext,
	printer printer,
	memoDir *fs.AbsolutePath,
) (rr *api.RunRecord, err error) {
	// Demux and initialize executor.
	executor, err := demuxExecutor(executorName)
	if err != nil {
		return nil, err
	}
	// If memodir was given, decorate the executor with memoization.
	if memoDir != nil {
		executor, err = memo.NewExecutor(*memoDir, executor)
		if err != nil {
			return nil, err
		}
	}

	// Prepare monitor and IO forwarding.
	evtChan := make(chan repeatr.Event)
	monitor := repeatr.Monitor{evtChan}
	monitorWg := sync.WaitGroup{}
	monitorWg.Add(1)
	go func() {
		defer monitorWg.Done()
		for {
			select {
			case evt, ok := <-evtChan:
				if !ok {
					return
				}
				switch {
				case evt.Log != nil:
					printer.printLog(*evt.Log)
				case evt.Output != nil:
					printer.printOutput(*evt.Output)
				case evt.Result != nil:
					// pass
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	inputControl := repeatr.InputControl{}

	// Run!  (And wait for output forwarding worker to finish.)
	rr, err = executor(
		ctx,
		formula,
		formulaCtx,
		inputControl,
		monitor,
	)
	close(monitor.Chan)
	monitorWg.Wait()

	// If a runrecord was returned always try to print it, even if we have
	//  an error and thus it may be incomplete.
	printer.printResult(repeatr.Event_Result{*rr, nil}.SetError(err))

	return rr, err
}
