package generate

import (
	"context"

	"github.com/antonmedv/expr/vm"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/expr"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

// listTags lists GitHub Tags by GitHub API and filter them with `version_filter`.
func (ctrl *Controller) listTags(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) []*github.RepositoryTag {
	// List GitHub Tags by GitHub API
	// Filter tags with version_filter
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 100, //nolint:gomnd
	}
	var versionFilter *vm.Program
	if pkgInfo.VersionFilter != nil {
		var err error
		versionFilter, err = expr.CompileVersionFilter(*pkgInfo.VersionFilter)
		if err != nil {
			return nil
		}
	}
	var arr []*github.RepositoryTag
	for i := 0; i < 10; i++ {
		tags, _, err := ctrl.github.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list releases")
			return arr
		}
		for _, tag := range tags {
			if versionFilter != nil {
				f, err := expr.EvaluateVersionFilter(versionFilter, tag.GetName())
				if err != nil || !f {
					continue
				}
			}
			arr = append(arr, tag)
		}
		if len(tags) != opt.PerPage {
			return arr
		}
		opt.Page++
	}
	return arr
}

func (ctrl *Controller) listAndGetTagNameFromTag(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
	// List GitHub Tags by GitHub API
	// Filter tags with version_filter
	// Get a tag
	repoOwner := pkgInfo.RepoOwner
	repoName := pkgInfo.RepoName
	opt := &github.ListOptions{
		PerPage: 30, //nolint:gomnd
	}
	var versionFilter *vm.Program
	if pkgInfo.VersionFilter != nil {
		vf, err := expr.CompileVersionFilter(*pkgInfo.VersionFilter)
		if err != nil {
			return ""
		}
		versionFilter = vf
	}
	for {
		tags, _, err := ctrl.github.ListTags(ctx, repoOwner, repoName, opt)
		if err != nil {
			logerr.WithError(logE, err).WithFields(logrus.Fields{
				"repo_owner": repoOwner,
				"repo_name":  repoName,
			}).Warn("list tags")
			return ""
		}
		for _, tag := range tags {
			tagName := tag.GetName()
			if versionFilter == nil {
				return tagName
			}
			f, err := expr.EvaluateVersionFilter(versionFilter, tagName)
			if err != nil || !f {
				continue
			}
			return tagName
		}
		if len(tags) != opt.PerPage {
			return ""
		}
		opt.Page++
	}
}

func (ctrl *Controller) selectVersionFromGitHubTag(ctx context.Context, logE *logrus.Entry, pkgInfo *registry.PackageInfo) string {
	tags := ctrl.listTags(ctx, logE, pkgInfo)
	versions := make([]*Version, len(tags))
	for i, tag := range tags {
		versions[i] = &Version{
			Name:    tag.GetName(),
			Version: tag.GetName(),
		}
	}
	idx, err := ctrl.versionSelector.Find(versions)
	if err != nil {
		return ""
	}
	return versions[idx].Version
}

func (ctrl *Controller) getVersionFromGitHubTag(ctx context.Context, logE *logrus.Entry, param *config.Param, pkgInfo *registry.PackageInfo) string {
	if param.SelectVersion {
		return ctrl.selectVersionFromGitHubTag(ctx, logE, pkgInfo)
	}
	return ctrl.listAndGetTagNameFromTag(ctx, logE, pkgInfo)
}
