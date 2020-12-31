package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/josa42/go-gitutils"
)

type RunOptions struct {
	Push    bool
	Verbose bool
}

func Run(o RunOptions) {
	s := &session{
		o:             o,
		defaultBranch: gitutils.DefaultBranch(),
	}
	s.syncGlobal()

	if gitutils.IsCurrentBranch(s.defaultBranch) {
		s.syncMainBranch()
	} else {
		s.syncFeatureBranch()
	}

	func(v interface{}) {
		s, _ := json.MarshalIndent(v, "", "  ")
		fmt.Printf("%s\n", s)
	}(gitutils.DefaultBranch())
}

type session struct {
	o             RunOptions
	defaultBranch string
}

var (
	grayBG = color.New(color.FgBlack, color.BgWhite).SprintfFunc()
	redBG  = color.New(color.FgBlack, color.BgRed).SprintfFunc()
	blueBG = color.New(color.FgBlack, color.BgBlue).SprintfFunc()
	blue   = color.New(color.FgBlue).SprintfFunc()
	purple = color.New(color.FgMagenta).SprintfFunc()
)

func (s *session) syncGlobal() {
	upstream := s.upstreamRemote()

	fmt.Printf("main: %s\nupstream: %s\n", s.defaultBranch, upstream)

	s.waitForLock()

	fmt.Printf("> git fetch %s --prune --prune-tags\n", upstream)
	// cmd git fetch $upstream --prune --prune-tags

	if !gitutils.IsCurrentBranch(s.defaultBranch) {
		s.info("%[2]s <= %[1]s/%[2]s", blue(upstream), purple(s.defaultBranch))
		fmt.Printf("> git fetch -u %[1]s %[2]s:%[2]s\n", upstream, s.defaultBranch)
		// cmd git fetch -u $upstream $main:$main
	}
}

func (s *session) syncMainBranch() {
	upstream := s.upstreamRemote()

	s.assertBranch(s.defaultBranch)

	s.resetBranchToUpstream()

	//   # Update fork origin
	if s.o.Push && upstream == "upstream" && gitutils.RemoteExists("origin") {
		s.pushOrigin()
	}

	s.cleanupMergedBranches()
}

func (s *session) syncFeatureBranch() {
	s.assertNotBranch(s.defaultBranch)

	s.pullUpstreamMaster()

	if s.o.Push {
		s.pushOrigin()
	}
}

func (s *session) cleanupMergedBranches() {
	s.info("Clean up")
	s.waitForLock()

	for _, branch := range gitutils.MergedBranches() {
		if branch == s.defaultBranch {
			continue
		}

		gitutils.DeleteBranch(branch)
	}
}

func (s *session) resetBranchToUpstream() {
	branch := gitutils.CurrentBranch()
	upstream := s.upstreamRemote()

	s.waitForLock()

	s.info("<= %s/%s", blue(upstream), purple(branch))
	fmt.Printf("> git reset --hard %s/%s\n", upstream, branch)
	// cmd git reset --hard "$upstream/$b"
}

func (s *session) pullUpstreamMaster() {
	branch := gitutils.CurrentBranch()
	upstream := s.upstreamRemote()

	s.waitForLock()

	s.info("<= %s/%s", blue(upstream), purple(branch))
	fmt.Printf("> git pull --rebase \"%s\" \"%s\"\n", upstream, branch)
	// cmd git pull --rebase "$upstream" "$b"
}

func (s *session) pushOrigin() {
	branch := gitutils.CurrentBranch()

	s.waitForLock()

	s.info("=> %s/%s", blue("origin"), purple(branch))
	fmt.Printf("git push --force origin \"%s\"\n", branch)
	// cmd git push --force origin "$b"
}

func (s *session) verbose(message string, args ...interface{}) {
	if s.o.Verbose {
		if len(args) > 0 {
			message = fmt.Sprintf(message, args...)
		}

		fmt.Printf("%s %s\n", grayBG(" verbose "), message)
	}
}

func (s *session) info(message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	fmt.Printf("%s %s\n", blueBG(" info    "), message)
}

func (s *session) error(message string, args ...interface{}) {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	fmt.Printf("%s %s\n", redBG(" error   "), message)
	os.Exit(1)
}

func (s *session) upstreamRemote() string {
	if gitutils.RemoteExists("upstream") {
		return "upstream"
	}
	return "origin"
}

func (s *session) assertBranch(branch string) {
	if !gitutils.IsCurrentBranch(branch) {
		s.error("not on %s branch", branch)
	}
}

func (s *session) assertNotBranch(branch string) {
	if gitutils.IsCurrentBranch(branch) {
		s.error("on %s branch", branch)
	}
}

func (s *session) waitForLock() {
	repo := ".git"

	for true {
		// TODO fix infinity look if not in a repo
		if _, err := os.Stat(repo); err == nil {
			break
		}
		repo = filepath.Join("..", repo)
	}

	lock, _ := filepath.Abs(filepath.Join(repo, "index.lock"))

	for i := 0; ; i++ {
		if _, err := os.Stat(lock); os.IsNotExist(err) {
			break
		}
		fmt.Printf("\r> Waiting for lock [%d]\r", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Print("\r                         \r")
}

// cmd() {
//   if [[ $VERBOSE -eq 1 ]]; then
//     verbose "$ $@"
//     eval "$@ 2>&1" | sed 's/^/          /g' || error "Command failed: $@"
//   else
//     eval "$@ 2> /dev/null > /dev/null" || error "Command failed: $@"
//   fi
// }
//
// run $@
//
