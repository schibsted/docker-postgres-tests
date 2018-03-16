package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"bitbucket.org/zombiezen/cardcpx/natsort"
	docopt "github.com/docopt/docopt-go"
	"github.com/heroku/docker-registry-client/registry"
)

const (
	readMeDisclaimer = `
<!--
Do not edit README.md, it is automatically generated and will be overridden on next build.

Instead, edit README-content.md that will be prepended by the list of available tags
-->

:bangbang: Those images are designed for test purposes and should *NOT* be used for production load :bangbang:

`
	alpineSuffix = "-alpine"
	usage        = `USAGE: ci [options]

Options:
	--user=<user>          The username for docker registry login
	--password=<password>  The password for docker registry login
	--push                 Push all branches to github
	--tag=<tag>            Add a tag pattern to whitelist. Only tags matching this pattern will be processed
`
)

func command(c string, args ...string) error {
	cmd := exec.Command(c, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func commit(f func(string, ...interface{}), workdir, message string, files ...string) {
	if err := command("git", append([]string{"-C", workdir, "diff", "--quiet", "HEAD", "--"}, files...)...); err != nil {
		ifErrorf(f, command("git", append([]string{"-C", workdir, "add", "--"}, files...)...), "failed to add %s to index for tag", strings.Join(files, " "))
		ifErrorf(f, command("git", append([]string{"-C", workdir, "commit", "-m", message, "--"}, files...)...), "failed to commit %s to index", strings.Join(files, " "))
	} else {
		fmt.Printf("files %s are up to date, nothing to commit", strings.Join(files, " "))
	}
}

func ifErrorf(f func(string, ...interface{}), err error, format string, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		format += " :%s"
		f(format, args...)
	}
}

func noopf(format string, args ...interface{}) {}

func releaseVersion(version string) string {
	suffix := ""
	if strings.HasSuffix(version, alpineSuffix) {
		suffix = alpineSuffix
		version = strings.TrimRight(version, alpineSuffix)
	}
	for _, pattern := range []string{"-rc", "-beta", "-alpha"} {
		re := regexp.MustCompile(pattern + ".*$")
		version = re.ReplaceAllString(version, ".0")
	}
	if strings.HasPrefix(version, "9") || strings.HasPrefix(version, "8") {
		version = regexp.MustCompile("([^.]+.[^.]+).*").ReplaceAllString(version, "$1")
	} else {
		version = strings.Split(version, ".")[0]
	}
	return version + suffix
}

func push(branches ...string) {
	branchesToPush := []string{}
	for _, branch := range branches {
		b := &bytes.Buffer{}
		cmd := exec.Command("git", "diff", "--shortstat", branch, "origin/"+branch)
		cmd.Stdout = b
		cmd.Stderr = b
		err := cmd.Run()
		if b.Len() > 0 || err != nil {
			branchesToPush = append(branchesToPush, branch)
		}
	}

	if len(branchesToPush) > 0 {
		cmd := []string{"push", "-f", "origin"}
		for _, tag := range branchesToPush {
			cmd = append(cmd, fmt.Sprintf("%s:refs/heads/%s", tag, tag))
		}
		ifErrorf(log.Fatalf, command("git", cmd...), "Failed to push new references")
	}
}

func main() {
	args, err := docopt.Parse(usage, os.Args[1:], true, "0.0", false)
	user := ""
	password := ""
	if args["--user"] != nil {
		user = args["--user"].(string)
	}
	if args["--password"] != nil {
		password = args["--password"].(string)
	}
	url := "https://registry-1.docker.io/"
	hub, err := registry.New(url, user, password)

	ifErrorf(log.Fatalf, err, "failed to create docker registry client for %s", url)
	dockerfile, err := ioutil.ReadFile("Dockerfile")
	ifErrorf(log.Fatalf, err, "failed to read Dockerfile")
	header := "FROM postgres\n"
	if string(dockerfile[:len(header)]) != header {
		log.Fatalf("dockerfile should start with '%v', got: '%v'", header, string(dockerfile[:len(header)]))
	}
	dockerfile = dockerfile[len(header):]
	tags, err := hub.Tags("library/postgres")
	ifErrorf(log.Fatalf, err, "failed to get tags for library/postgres on docker hub")

	dockerfileSSL, err := ioutil.ReadFile("Dockerfile.ssl")
	ifErrorf(log.Fatalf, err, "failed to read Dockerfile.ssl")

	workdir := "/tmp/postgres-build"

	readme := []byte(readMeDisclaimer)
	readme = append(readme, []byte("\n# Supported tags and respective `Dockerfile` links\n")...)
	natsort.Strings(tags)
	l := len(tags)
	for i := 0; i < l/2; i++ {
		tags[i], tags[l-1-i] = tags[l-1-i], tags[i]
	}
	matrixTags := []string{}
	versions := map[string]interface{}{}
	for _, tag := range tags {
		classifier := releaseVersion(tag)
		if _, ok := versions[classifier]; !ok {
			readme = append(readme, fmt.Sprintf(
				"\n-	[`%s` (*%s:Dockerfile*)](https://github.com/schibsted/docker-postgres-tests/blob/%s/Dockerfile)", tag, tag, tag)...)
			versions[classifier] = nil
			matrixTags = append(matrixTags, tag)
		}
	}
	readmeContent, err := ioutil.ReadFile("README-content.md")
	ifErrorf(log.Fatalf, err, "failed to read README content")
	readme = append(readme, []byte("\n\n")...)
	readme = append(readme, readmeContent...)
	ifErrorf(log.Fatalf, ioutil.WriteFile("./README.md", readme, 0666), "failed to write README")

	travis, err := ioutil.ReadFile(".travis.yml")
	ifErrorf(log.Fatalf, err, "failed to read travis yml content")

	travis = travis[:bytes.Index(travis, []byte("\nmatrix:\n"))]
	travis = append(travis, []byte("\nmatrix:\n  include:\n")...)
	for _, tag := range matrixTags {
		travis = append(travis, []byte(fmt.Sprintf("  - env: TEST_TAG=%s\n", tag))...)
	}
	ifErrorf(log.Fatalf, ioutil.WriteFile(".travis.yml", travis, 0666), "failed to write travis yml")
	commit(log.Fatalf, ".", "[auto] update Readme and travis yml", "README.md", ".travis.yml")

	defer func() {
		ifErrorf(noopf, command("rm", "-rf", workdir), "")
		ifErrorf(noopf, command("git", "worktree", "prune"), "")
		ifErrorf(noopf, command("git", "branch", "rm", "-f", "pg-update"), "")
	}()
	ifErrorf(log.Fatalf, command("git", "branch", "-f", "pg-update"), "failed to create update branch")
	ifErrorf(log.Fatalf, command("git", "worktree", "add", workdir, "pg-update"), "failed to create git worktree")

	keep := func(string) bool { return true }

	if args["--tag"] != nil {
		keep = func(name string) bool {
			matched, err := filepath.Match(args["--tag"].(string), name)
			ifErrorf(log.Fatalf, err, "failed to match %s using pattern %s", name, args["--tag"].(string))
			return matched
		}
	}

	tagsToPush := []string{}
	for _, tag := range tags {
		if keep(tag) {
			ifErrorf(noopf, command("git", "-C", workdir, "branch", "-f", tag, "master"), "")
			ifErrorf(log.Fatalf, command("git", "-C", workdir, "checkout", tag), "failed to checkout to %s in %s", tag, workdir)
			d := append([]byte("FROM postgres:"+tag+"\n"), dockerfile...)
			ifErrorf(log.Fatalf, ioutil.WriteFile(workdir+"/Dockerfile", d, 0666), "failed to write Dockerfile for tag %s", tag)
			commit(log.Fatalf, workdir, "[auto] use postgres tag "+tag, "Dockerfile")
			tagsToPush = append(tagsToPush, tag)

			tag += "-ssl"
			ifErrorf(noopf, command("git", "-C", workdir, "branch", "-f", tag, "master"), "")
			ifErrorf(log.Fatalf, command("git", "-C", workdir, "checkout", tag), "failed to checkout to %s in %s", tag, workdir)
			ifErrorf(log.Fatalf, ioutil.WriteFile(workdir+"/Dockerfile", append(d, dockerfileSSL...), 0666), "failed to write Dockerfile for tag %s", tag)
			commit(log.Fatalf, workdir, "[auto] update SSL Readme and use postgres tag "+tag, "Dockerfile")
			tagsToPush = append(tagsToPush, tag)
		}
	}

	if args["--push"].(bool) {
		push(append([]string{"master"}, tagsToPush...)...)
	}
}
