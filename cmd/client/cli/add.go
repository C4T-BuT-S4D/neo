package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"neo/internal/client"
	"neo/pkg/archive"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	neopb "neo/lib/genproto/neo"
)

type addCLI struct {
	*baseCLI
	path      string
	isArchive bool
	exploitID string
}

func NewAdd(cmd *cobra.Command, args []string, cfg *client.Config) *addCLI {
	eid, err := cmd.Flags().GetString("id")
	if err != nil {
		logrus.Fatalf("Could not get exploit id")
	}
	isDir, err := cmd.Flags().GetBool("dir")
	if err != nil {
		logrus.Fatalf("Could not get dir param")
	}
	return &addCLI{
		baseCLI:   &baseCLI{cfg},
		path:      args[0],
		isArchive: isDir,
		exploitID: eid,
	}
}

func (ac *addCLI) Run(ctx context.Context) error {
	_, err := os.Stat(ac.path)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", ac.path, err)
	}
	// Replace path with abs path.
	if ac.path, err = filepath.Abs(ac.path); err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if errs := ac.validateEntry(ac.path); len(errs) > 0 {
		for _, v := range errs {
			logrus.Errorf("%v", v)
		}
		return errors.New("invalid exploit")
	}

	dir, file := path.Split(ac.path)
	if ac.exploitID == "" {
		ac.exploitID = file
	}
	logrus.Infof("Going to add exploit with id = %s", ac.exploitID)

	c, err := ac.client()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	state, err := c.Ping(ctx, neopb.PingRequest_CONFIG_REQUEST)
	if err != nil {
		return fmt.Errorf("failed to get config from server: %w", err)
	}
	exists := false
	for _, v := range state.GetExploits() {
		if v.GetExploitId() == ac.exploitID {
			exists = true
			break
		}
	}

	if exists {
		fmt.Println("The exploit with this id already exists. Do you want to override (add new version) y/N ?")
		var tmp string
		if _, err := fmt.Scanln(&tmp); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if !strings.Contains(strings.ToLower(tmp), "y") {
			logrus.Fatalf("Aborted.")
		}
	}

	var f *os.File
	if ac.isArchive {
		f, err = ioutil.TempFile("", "ARCHIVE")
		if err != nil {
			return fmt.Errorf("failed to create tmpfile: %w", err)
		}

		defer func() {
			_ = os.Remove(f.Name())
		}()
		if err := archive.Tar(dir, f); err != nil {
			return fmt.Errorf("failed to create TarGz archive: %w", err)
		}

		// Seek file to start to correctly use it for reading.
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek archive file: %w", err)
		}
	} else {
		f, err = os.Open(ac.path)
		if err != nil {
			return fmt.Errorf("failed to open exploit path: %w", err)
		}
	}
	defer func() {
		_ = f.Close()
	}()

	fileInfo, err := c.UploadFile(ctx, f)
	if err != nil {
		return fmt.Errorf("failed to upload exploit file: %w", err)
	}

	req := &neopb.UpdateExploitRequest{
		ExploitId: ac.exploitID,
		File:      fileInfo,
		Config: &neopb.ExploitConfiguration{
			Entrypoint: file,
			IsArchive:  ac.isArchive,
		},
	}
	if err := c.UpdateExploit(ctx, req); err != nil {
		return fmt.Errorf("failed to update exploit: %w", err)
	}
	return nil
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
