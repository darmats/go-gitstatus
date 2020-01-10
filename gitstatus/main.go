package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	// run() // not handling
}

func run() error {
	if err := os.Setenv("GIT_PAGER_IN_USE", "true"); err != nil {
		return err
	}
	bytes, err := exec.Command("git", "status", "--porcelain", "--branch").Output()
	if err != nil {
		return err
	}

	status := strings.Split(string(bytes), "\n")

	var untracked, staged, changed, conflicts uint
	var ahead, behind = "0", "0"
	var branch string

	for _, st := range status {
		if st == "" {
			break
		}
		if st[0] == '#' && st[1] == '#' {
			switch {
			case strings.Contains(st, "Initial commit on"):
				fallthrough
			case strings.Contains(st, "No commits yet on"):
				sp := strings.Split(st[2:], " ")
				branch = sp[len(sp)-1]
			case strings.Contains(st, "no branch"):
				branch, err = getTagNameOrHash()
				if err != nil {
					return err
				}
			default:
				sp := strings.Split(strings.TrimSpace(st[2:]), "...")
				if len(sp) == 1 {
					// remoteがない
					branch = strings.TrimSpace(st[2:])
				} else {
					var rest string
					branch, rest = sp[0], sp[1]
					rsp := strings.SplitN(rest, " ", 2)
					if len(rsp) == 1 {
						// ahead or behindがない
					} else {
						// ## branch_name...origin/branch_name [ahead 1, behind 2]
						for _, div := range strings.Split(rsp[1][1:len(rsp[1])-1], ", ") {
							if strings.Contains(div, "ahead ") {
								ahead = div[6:]
							} else if strings.Contains(div, "behind ") {
								behind = div[7:]
							}
						}
					}
				}
			}
		} else if st[0] == '?' && st[1] == '?' {
			untracked++
		} else {
			if st[1] == 'M' || st[1] == 'D' {
				changed++
			}
			if st[0] == 'U' {
				conflicts++
			} else if st[0] != ' ' {
				staged++
			}
		}
	}

	fmt.Print(branch, " ", ahead, " ", behind, " ", staged, " ", conflicts, " ", changed, " ", untracked)
	return nil
}

var (
	tagExp = regexp.MustCompile(` \(tag: `)
)

func getTagNameOrHash() (string, error) {
	bytes, err := exec.Command("git", "log", "-1", "--format=%h%d", "--decorate-refs=tags").Output()
	if err != nil {
		return "", nil
	}
	status := string(bytes) // "1234567" or "1234567 (tag: TAG)"

	tags := tagExp.Split(status, -1)
	if len(tags) <= 1 {
		// hash
		return ":" + strings.TrimSpace(tags[0]), nil
	}
	tag := strings.TrimSpace(tags[1])
	return "tag/" + tag[0:len(tag)-1], nil
}
