// Package main is responsible for command-line interface of dnsproxy.
package main

//import (
//	"fmt"
//	"github.com/AdguardTeam/golibs/log"
//	"github.com/AdguardTeam/golibs/logutil/slogutil"
//	"github.com/AdguardTeam/golibs/osutil"
//	"github.com/fridaylabx/dnsproxy/sv"
//	"golang.org/x/net/context"
//	"gopkg.in/natefinch/lumberjack.v2"
//	"os"
//)
//
//// main is the entry point.
//func main() {
//	opts, exitCode, err := sv.ParseOptions()
//	if err != nil {
//		_, _ = fmt.Fprintln(os.Stderr, err)
//	}
//
//	if opts == nil {
//		os.Exit(exitCode)
//	}
//
//	logOutput := os.Stdout
//	if opts.LogOutput != "" {
//		// #nosec G302 -- Trust the file path that is given in the
//		// configuration.
//		logOutput, err = os.OpenFile(opts.LogOutput, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
//		if err != nil {
//			_, _ = fmt.Fprintln(os.Stderr, fmt.Errorf("cannot create a log file: %s", err))
//
//			os.Exit(osutil.ExitCodeArgumentError)
//		}
//
//		defer func() { _ = logOutput.Close() }()
//	}
//
//	l := slogutil.New(&slogutil.Config{
//		Output: logOutput,
//		Format: slogutil.FormatDefault,
//		// TODO(d.kolyshev): Consider making configurable.
//		AddTimestamp: true,
//		Verbose:      opts.Verbose,
//	})
//
//	if opts.LogOutput != "" {
//		log.SetOutput(&lumberjack.Logger{
//			Filename:   opts.LogOutput,
//			MaxSize:    30,
//			MaxBackups: 5,
//		})
//	}
//
//	ctx := context.Background()
//
//	if opts.Pprof {
//		sv.RunPprof(l)
//	}
//
//	err = sv.RunProxy(ctx, l, opts)
//	if err != nil {
//		l.ErrorContext(ctx, "running dnsproxy", slogutil.KeyError, err)
//
//		// As defers are skipped in case of os.Exit, close logOutput manually.
//		//
//		// TODO(a.garipov): Consider making logger.Close method.
//		if logOutput != os.Stdout {
//			_ = logOutput.Close()
//		}
//
//		os.Exit(osutil.ExitCodeFailure)
//	}
//}
