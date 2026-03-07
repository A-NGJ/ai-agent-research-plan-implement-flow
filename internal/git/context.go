package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Context struct {
	Branch         string       `json:"branch"`
	Commit         string       `json:"commit"`
	Status         StatusInfo   `json:"status"`
	RecentCommits  []CommitInfo `json:"recent_commits"`
	DiffSummary    DiffSummary  `json:"diff_summary"`
	SensitiveFiles []string     `json:"sensitive_files"`
}

type StatusInfo struct {
	Staged    []string `json:"staged"`
	Modified  []string `json:"modified"`
	Untracked []string `json:"untracked"`
}

type CommitInfo struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
}

type DiffSummary struct {
	FilesChanged int `json:"files_changed"`
	Insertions   int `json:"insertions"`
	Deletions    int `json:"deletions"`
}

type SensitiveMatch struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

var sensitiveFilePatterns = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	{regexp.MustCompile(`(?i)\.env$`), "matches .env pattern"},
	{regexp.MustCompile(`(?i)\.env\.`), "matches .env pattern"},
	{regexp.MustCompile(`(?i)credentials`), "matches credentials pattern"},
	{regexp.MustCompile(`(?i)secret`), "matches secret pattern"},
	{regexp.MustCompile(`(?i)api_key`), "matches api_key pattern"},
	{regexp.MustCompile(`(?i)private_key`), "matches private_key pattern"},
	{regexp.MustCompile(`(?i)\.pem$`), "matches .pem pattern"},
	{regexp.MustCompile(`(?i)\.key$`), "matches .key pattern"},
}

var sensitiveContentPatterns = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	{regexp.MustCompile(`password\s*=`), "contains password assignment"},
	{regexp.MustCompile(`API_KEY\s*=`), "contains API_KEY assignment"},
	{regexp.MustCompile(`BEGIN RSA PRIVATE KEY`), "contains RSA private key"},
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func GatherContext() (*Context, error) {
	ctx := &Context{}

	branch, err := runGit("branch", "--show-current")
	if err != nil {
		ctx.Branch = "HEAD"
	} else if branch == "" {
		ctx.Branch = "HEAD"
	} else {
		ctx.Branch = branch
	}

	commit, err := runGit("rev-parse", "--short", "HEAD")
	if err != nil {
		ctx.Commit = ""
	} else {
		ctx.Commit = commit
	}

	statusOut, err := runGit("status", "--porcelain")
	if err != nil {
		ctx.Status = StatusInfo{Staged: []string{}, Modified: []string{}, Untracked: []string{}}
	} else {
		ctx.Status = ParseStatus(statusOut)
	}

	logOut, err := runGit("log", "--oneline", "-10")
	if err != nil {
		ctx.RecentCommits = []CommitInfo{}
	} else {
		ctx.RecentCommits = ParseLog(logOut)
	}

	diffOut, err := runGit("diff", "--stat", "HEAD")
	if err != nil {
		ctx.DiffSummary = DiffSummary{}
	} else {
		ctx.DiffSummary = ParseDiffStat(diffOut)
	}

	sensitive, err := SensitiveCheck()
	if err != nil {
		ctx.SensitiveFiles = []string{}
	} else {
		files := make([]string, len(sensitive))
		for i, m := range sensitive {
			files[i] = m.File
		}
		ctx.SensitiveFiles = files
	}

	return ctx, nil
}

func ChangedFiles() ([]string, error) {
	out, err := runGit("diff", "--name-only", "main...HEAD")
	if err != nil || out == "" {
		out, err = runGit("diff", "--name-only", "HEAD~10")
		if err != nil {
			return []string{}, nil
		}
	}
	if out == "" {
		return []string{}, nil
	}
	return strings.Split(out, "\n"), nil
}

func SensitiveCheck() ([]SensitiveMatch, error) {
	out, err := runGit("diff", "--cached", "--name-only")
	if err != nil {
		return []SensitiveMatch{}, nil
	}
	if out == "" {
		return []SensitiveMatch{}, nil
	}

	files := strings.Split(out, "\n")
	var matches []SensitiveMatch

	matches = append(matches, SensitiveFilenames(files)...)

	for _, file := range files {
		content, err := runGit("show", ":0:"+file)
		if err != nil {
			continue
		}
		matches = append(matches, SensitiveContent(file, content)...)
	}

	return matches, nil
}

func ParseStatus(output string) StatusInfo {
	info := StatusInfo{
		Staged:    []string{},
		Modified:  []string{},
		Untracked: []string{},
	}
	if output == "" {
		return info
	}

	for _, line := range strings.Split(output, "\n") {
		if len(line) < 4 {
			continue
		}
		x := line[0]
		y := line[1]
		file := line[3:]

		if x == '?' {
			info.Untracked = append(info.Untracked, file)
			continue
		}
		if x != ' ' && x != '?' {
			info.Staged = append(info.Staged, file)
		}
		if y != ' ' && y != '?' {
			info.Modified = append(info.Modified, file)
		}
	}

	return info
}

func ParseDiffStat(output string) DiffSummary {
	ds := DiffSummary{}
	if output == "" {
		return ds
	}

	lines := strings.Split(output, "\n")
	summary := lines[len(lines)-1]

	filesRe := regexp.MustCompile(`(\d+) files? changed`)
	insRe := regexp.MustCompile(`(\d+) insertions?\(\+\)`)
	delRe := regexp.MustCompile(`(\d+) deletions?\(-\)`)

	if m := filesRe.FindStringSubmatch(summary); m != nil {
		ds.FilesChanged, _ = strconv.Atoi(m[1])
	}
	if m := insRe.FindStringSubmatch(summary); m != nil {
		ds.Insertions, _ = strconv.Atoi(m[1])
	}
	if m := delRe.FindStringSubmatch(summary); m != nil {
		ds.Deletions, _ = strconv.Atoi(m[1])
	}

	return ds
}

func ParseLog(output string) []CommitInfo {
	if output == "" {
		return []CommitInfo{}
	}

	var commits []CommitInfo
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		commits = append(commits, CommitInfo{Hash: parts[0], Message: parts[1]})
	}
	return commits
}

func SensitiveFilenames(files []string) []SensitiveMatch {
	var matches []SensitiveMatch
	for _, file := range files {
		for _, p := range sensitiveFilePatterns {
			if p.pattern.MatchString(file) {
				matches = append(matches, SensitiveMatch{File: file, Reason: p.reason})
				break
			}
		}
	}
	return matches
}

func SensitiveContent(file, content string) []SensitiveMatch {
	var matches []SensitiveMatch
	for _, p := range sensitiveContentPatterns {
		if p.pattern.MatchString(content) {
			matches = append(matches, SensitiveMatch{File: file, Reason: p.reason})
		}
	}
	return matches
}
