package slsa

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/aquaproj/aqua/pkg/template"
	"github.com/sirupsen/logrus"
	"github.com/slsa-framework/slsa-verifier/v2/cli/slsa-verifier/verify"
	"github.com/spf13/afero"
)

type VerifierImpl struct {
	downloader download.ClientAPI
	fs         afero.Fs
	exe        Executor
}

func New(downloader download.ClientAPI, fs afero.Fs, exe Executor) *VerifierImpl {
	return &VerifierImpl{
		downloader: downloader,
		fs:         fs,
		exe:        exe,
	}
}

type Verifier interface {
	Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *ParamVerify) error
}

type MockVerifier struct {
	err error
}

func (mock *MockVerifier) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *ParamVerify) error {
	return mock.err
}

type ParamVerify struct {
	// e.g. github.com/suzuki-shunsuke/test-cosign-keyless-aqua
	SourceURI string
	// e.g. v0.1.0-7
	SourceTag    string
	ArtifactPath string
}

type ExecutorImpl struct{}

func NewExecutor() *ExecutorImpl {
	return &ExecutorImpl{}
}

type Executor interface {
	Verify(ctx context.Context, param *ParamVerify, provenancePath string) error
}

type MockExecutor struct {
	Err error
}

func (mock *MockExecutor) Verify(ctx context.Context, param *ParamVerify, provenancePath string) error {
	return mock.Err
}

func (exe *ExecutorImpl) Verify(ctx context.Context, param *ParamVerify, provenancePath string) error {
	v := verify.VerifyArtifactCommand{
		ProvenancePath: provenancePath,
		SourceURI:      param.SourceURI,
		SourceTag:      &param.SourceTag,
	}
	if _, err := v.Exec(ctx, []string{param.ArtifactPath}); err != nil {
		return fmt.Errorf("run slsa-verifier's verify-artifact command: %w", err)
	}
	return nil
}

func (verifier *VerifierImpl) Verify(ctx context.Context, logE *logrus.Entry, rt *runtime.Runtime, sp *registry.SLSAProvenance, art *template.Artifact, file *download.File, param *ParamVerify) error {
	f, err := download.ConvertDownloadedFileToFile(sp.ToDownloadedFile(), file, rt, art)
	if err != nil {
		return err //nolint:wrapcheck
	}
	rc, _, err := verifier.downloader.GetReadCloser(ctx, logE, f)
	if err != nil {
		return fmt.Errorf("download a SLSA Provenance: %w", err)
	}
	defer rc.Close()

	provenanceFile, err := afero.TempFile(verifier.fs, "", "")
	if err != nil {
		return fmt.Errorf("create a temporal file: %w", err)
	}
	defer provenanceFile.Close()
	defer verifier.fs.Remove(provenanceFile.Name()) //nolint:errcheck
	if _, err := io.Copy(provenanceFile, rc); err != nil {
		return fmt.Errorf("copy a provenance to a temporal file: %w", err)
	}

	return verifier.exe.Verify(ctx, param, provenanceFile.Name()) //nolint:wrapcheck
}
