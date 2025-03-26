package git

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/NishantBansal2003/LND-Fuzz/config"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CloneRepositories clones the project and storage repositories based on the
// provided configuration. It clones the project repository into
// config.ProjectDir and the storage repository into config.CorpusDir, logging
// progress and returning an error if any cloning step fails.
func CloneRepositories(ctx context.Context, logger *slog.Logger,
	cfg *config.Config) error {

	logger.Info("Cloning project repository", "url", cfg.ProjectSrcPath)
	if _, err := git.PlainCloneContext(ctx, config.ProjectDir, false,
		&git.CloneOptions{
			URL: cfg.ProjectSrcPath,
		}); err != nil {
		return fmt.Errorf("project clone failed: %w", err)
	}

	logger.Info("Cloning storage repository", "url", cfg.GitStorageRepo)
	if _, err := git.PlainCloneContext(ctx, config.CorpusDir, false,
		&git.CloneOptions{
			URL: cfg.GitStorageRepo,
		}); err != nil {
		return fmt.Errorf("storage repo clone failed: %w", err)
	}

	return nil
}

// CommitAndPushResults commits any changes in the corpus repository and pushes
// the commit to the remote repository. It opens the corpus repository from
// config.CorpusDir, checks for uncommitted changes, stages changes, creates a
// commit using the provided commit message and author information, and then
// pushes the commit. If there are no changes to commit, it logs that
// information.
func CommitAndPushResults(logger *slog.Logger) error {
	repo, err := git.PlainOpen(config.CorpusDir)
	if err != nil {
		return fmt.Errorf("failed to open git repo: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if status, err := worktree.Status(); err != nil || status.IsClean() {
		logger.Info("No corpus changes to commit")
		return err
	}

	if _, err := worktree.Add("."); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	commitOpts := &git.CommitOptions{
		Author: &object.Signature{
			Name:  config.GitUserName,
			Email: config.GitUserEmail,
			When:  time.Now(),
		},
	}

	if _, err := worktree.Commit(config.CommitMessage, commitOpts); err !=
		nil {

		return fmt.Errorf("commit failed: %w", err)
	}

	if err := repo.Push(&git.PushOptions{}); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	logger.Info("Successfully updated corpus repository")
	return nil
}
