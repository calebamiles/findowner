package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Owners struct {
	Approvers []string `yaml:"approvers"`
	Reviewers []string `yaml:"reviewers"`
}

var gitRepo string

func init() {
	flag.StringVar(&gitRepo, "gitrepo", "", "")
	flag.Parse()
}

func patch(path string, rs []string) {
	p := filepath.Join(gitRepo, path, "OWNERS")

	existingOwnersBytes, err := ioutil.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	existingOwners := Owners{}
	err = yaml.Unmarshal(existingOwnersBytes, &existingOwners)
	if err != nil {
		panic(err)
	}

	existingOwnersSet := make(map[string]bool)
	for _, existingOwner := range existingOwners.Reviewers {
		existingOwnersSet[existingOwner] = true
	}

	for _, newOwner := range rs {
		if !existingOwnersSet[newOwner] {
			existingOwners.Reviewers = append(existingOwners.Reviewers, newOwner)
		}
	}

	newOwnersBytes, err := yaml.Marshal(existingOwners)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(p, newOwnersBytes, os.ModePerm)
	if err != nil {
		panic(err)
	}

}

// How to run it:
//   cat $OWNER_FILE | ./genpatch --gitrepo="$GOPATH/src/k8s.io/kubernetes"
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		segs := strings.Split(line, ",")
		if len(segs) != 2 {
			panic("unexpected")
		}
		pathSegs := strings.Split(segs[0], ":")
		if len(pathSegs) != 2 {
			panic("unexpected")
		}
		path := strings.TrimSpace(pathSegs[1])

		reviwerSegs := strings.Split(segs[1], ":")
		if len(reviwerSegs) != 2 {
			panic("unexpected")
		}
		reviewerListStr := strings.TrimSpace(reviwerSegs[1])
		reviewers := strings.Split(reviewerListStr[1:len(reviewerListStr)-1], " ")

		patch(path, reviewers)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
