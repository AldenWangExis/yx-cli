package safety

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	Summary string
	DryRun  bool
	Yes     bool
}

type Environment struct {
	ConfirmWrites bool
	IsTerminal    bool
	Input         io.Reader
	Output        io.Writer
}

type Decision struct {
	Allowed bool
	DryRun  bool
}

func Decide(req Request, env Environment) (Decision, error) {
	if req.DryRun {
		return Decision{DryRun: true}, nil
	}
	if !env.ConfirmWrites || req.Yes {
		return Decision{Allowed: true}, nil
	}
	if !env.IsTerminal {
		return Decision{}, errors.New("confirmation required; pass --yes to run non-interactively")
	}
	if env.Input == nil {
		return Decision{}, errors.New("confirmation input is unavailable")
	}
	if env.Output != nil {
		if req.Summary != "" {
			fmt.Fprintf(env.Output, "%s\n", req.Summary)
		}
		fmt.Fprint(env.Output, "Continue? [y/N] ")
	}

	line, err := bufio.NewReader(env.Input).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return Decision{}, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	if answer != "y" && answer != "yes" {
		return Decision{}, errors.New("operation cancelled")
	}
	return Decision{Allowed: true}, nil
}
