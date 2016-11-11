package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	var (
		root, host, org, repo string
		update                = flag.Bool("u", false, "update existing repo if exists")
		verbose               = flag.Bool("v", false, "verbose output")
	)
	flag.Parse()

	out := ioutil.Discard
	if *verbose {
		out = os.Stderr
	}

	if len(flag.Args()) < 1 {
		log.Fatalf("usage of %s : [ git url | github browse url]\n")
	}
	cloneURL := flag.Arg(0)
	gopath, ok := os.LookupEnv("GOPATH")
	if !ok {
		log.Fatal("need GOPATH env variable")
	}
	root = path.Join(gopath, "src")

	if strings.HasPrefix(cloneURL, "git@") {
		cloneURL = cloneURL[4:]
		ss := strings.SplitN(cloneURL, ":", 2)
		host = ss[0]
		ss = strings.SplitN(ss[1], "/", 3)
		org = ss[0]
		repo = strings.TrimRight(ss[1], filepath.Ext(ss[1]))

	} else if u, err := url.Parse(cloneURL); err == nil {
		ss := strings.SplitN(strings.TrimLeft(u.Path, "/"), "/", 3)
		host = u.Host
		org = ss[0]
		repo = ss[1]
	}
	// fmt.Println("host:", host, "org:", org, "repo:", repo)
	gitURL := fmt.Sprintf("git@%s:%s/%s.git", host, org, repo)
	targetPath := path.Join(root, host, org, repo)

	if fi, err := os.Stat(targetPath); err == nil && fi.IsDir() {
		//targetPath exists
		if fi, err := os.Stat(path.Join(targetPath, ".git")); err == nil && fi.IsDir() {
			//targetPath is a git repo
			if !*update {
				log.Fatalf("Target %q not updated", targetPath)
			}
			//git pull
			fmt.Fprintln(out, "updating", targetPath)
			cmd := exec.Command("git", "pull")
			cmd.Dir = targetPath
			cmd.Stdout = out
			cmd.Stderr = out
			err := cmd.Run()
			if err != nil {
				log.Fatalf("Error executing cmd %v: %v", []string{"git", "pull"}, err)
			}
			return
		}
		log.Fatalf("Target dir %v is not a git repo.", targetPath)
	}

	//git clone
	cmd := exec.Command("git", "clone", gitURL, targetPath)
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error executing cmd %v: %v", []string{"git", "clone", gitURL, targetPath}, err)
	}
	fmt.Fprintln(out, "checked out", targetPath)
}
