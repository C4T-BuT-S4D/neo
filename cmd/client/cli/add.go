package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	neopb "neo/lib/genproto/neo"
	"neo/pkg/archive"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"neo/internal/client"

	"github.com/sirupsen/logrus"
)

type addCLI struct {
	*baseCLI
	path      string
	isArchive bool
	exploitID string
}

func NewAdd(args []string, cfg *client.Config) *addCLI {
	flags := flag.NewFlagSet("add", flag.ExitOnError)
	isDir := flags.Bool("dir", false, "Dump exploit as an archive with entrypoint.")
	eid := flags.String("id", "", "Id of the exploit. Will determinate it by path by default.")
	if err := flags.Parse(args); err != nil {
		logrus.Fatalf("add: failed to parse cli flags: %v", err)
	}
	if flags.NArg() < 1 {
		logrus.Fatalf("Usage: %s add <path_to_exploit_folder_or_script>", os.Args[0])
	}
	return &addCLI{
		baseCLI:   &baseCLI{cfg},
		path:      flags.Arg(0),
		isArchive: *isDir,
		exploitID: *eid,
	}
}

func (ac *addCLI) Run(ctx context.Context) error {
	_, err := os.Stat(ac.path)
	if err != nil {
		logrus.Fatalf("add: failed to stat file(%s): %v", ac.path, err)
	}
	// Replace path with abs path.
	if ac.path, err = filepath.Abs(ac.path); err != nil {
		logrus.Fatalf("Failed to get absolute path: %v", err)
	}

	if errors := ac.validateEntry(ac.path); len(errors) > 0 {
		for _, v := range errors {
			fmt.Println(v)
		}
		logrus.Fatalf("Entrypoint is invalid.")
	}

	dir, file := path.Split(ac.path)
	if ac.exploitID == "" {
		ac.exploitID = file
	}
	fmt.Printf("Going to add exploit with id = %s\n", ac.exploitID)

	client, err := ac.client()
	if err != nil {
		logrus.Fatalf("add: failed to create client: %v", err)
	}
	state, err := client.Ping(ctx)
	if err != nil {
		logrus.Fatalf("add: failed to get config from server: %v", err)
	}
	exists := false
	for _, v := range state.GetExploits() {
		if v.GetExploitId() == ac.exploitID {
			exists = true
			break
		}
	}

	if exists {
		fmt.Println("The exploit with this id already exists. Do you wan't to override(add new version) y/n ?")
		var tmp string
		fmt.Scanln(&tmp)
		if !strings.Contains(strings.ToLower(tmp), "y") {
			logrus.Fatalf("Aborted.")
		}
	}

	var f *os.File
	if ac.isArchive {
		f, err = ioutil.TempFile("", "ARCHIVE")
		if err != nil {
			logrus.Fatalf("Failed to create tmpfile: %v", err)
		}

		defer os.Remove(f.Name())
		if err := archive.Tar(dir, f); err != nil {
			logrus.Fatalf("Failed to create TarGz archive: %v", err)
		}

		// Seek file to start to correctly use it for reading.
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			logrus.Fatalf("Failed to seek archive file: %v", err)
		}
	} else {
		f, err = os.Open(ac.path)
		if err != nil {
			logrus.Fatalf("Failed to open exploit path: %v", err)
		}
	}
	defer f.Close()

	fileInfo, err := client.UploadFile(ctx, f)
	if err != nil {
		logrus.Fatalf("Failed to upload exploit file: %v", err)
	}

	req := &neopb.UpdateExploitRequest{
		ExploitId: ac.exploitID,
		File:      fileInfo,
		Config: &neopb.ExploitConfiguration{
			Entrypoint: file,
			IsArchive:  ac.isArchive,
		},
	}
	return client.UpdateExploit(ctx, req)
}

func (ac *addCLI) validateEntry(f string) (errors []string) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		errors = append(errors, err.Error())
		return
	}
	if string(data[:2]) != "#!" {
		errors = append(errors,
			fmt.Sprintf("Please use shebang (e.g. %s) as the first line of your script",
				"#!/usr/bin/env python3"))
	}
	var re = regexp.MustCompile(`(?m)flush[(=]`)
	if !re.Match(data) {
		errors = append(errors, fmt.Sprintf("Please use print(..., flush=True)"))
	}
	return
}
