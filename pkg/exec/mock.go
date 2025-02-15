package exec

import (
	"context"
)

type Mock struct {
	ExitCode int
	Err      error
}

func (exe *Mock) Exec(ctx context.Context, exePath string, args []string) (int, error) {
	return exe.ExitCode, exe.Err
}

func (exe *Mock) ExecWithEnvs(ctx context.Context, exePath string, args, envs []string) (int, error) {
	return exe.ExitCode, exe.Err
}

func (exe *Mock) ExecXSys(exePath string, args []string) error {
	return exe.Err
}

func (exe *Mock) GoBuild(ctx context.Context, exePath, src, exeDir string) (int, error) {
	return exe.ExitCode, exe.Err
}

func (exe *Mock) GoInstall(ctx context.Context, path, gobin string) (int, error) {
	return exe.ExitCode, exe.Err
}
