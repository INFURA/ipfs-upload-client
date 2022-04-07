package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	ipfsFiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	coreiface "github.com/ipfs/interface-go-ipfs-core"
	caopts "github.com/ipfs/interface-go-ipfs-core/options"
	ipfsPath "github.com/ipfs/interface-go-ipfs-core/path"
	flag "github.com/spf13/pflag"
)

const infuraAPI = "https://ipfs.infura.io:5001"

func main() {
	projectId := flag.String("id", "", "your Infura ProjectID")
	projectSecret := flag.String("secret", "", "your Infura ProjectSecret")
	api := flag.String("url", infuraAPI, "the API URL")
	pin := flag.Bool("pin", true, "whether or not to pin the data")

	flag.Parse()

	if *projectId == "" {
		_, _ = fmt.Fprintln(os.Stderr, "parameter --id is required")
		os.Exit(-1)
	}
	if *projectSecret == "" {
		_, _ = fmt.Fprintln(os.Stderr, "parameter --secret is required")
		os.Exit(-1)
	}

	httpClient := &http.Client{}
	client, err := httpapi.NewURLApiWithClient(*api, httpClient)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	client.Headers.Add("Authorization", "Basic "+basicAuth(*projectId, *projectSecret))

	args := flag.Args()
	if len(args) != 1 {
		_, _ = fmt.Fprintln(os.Stderr, "file or directory path required as an argument")
		os.Exit(-1)
	}
	path := args[0]

	stat, err := os.Lstat(path)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// also support directory
	file, err := ipfsFiles.NewSerialFile(path, false, stat)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	var res ipfsPath.Resolved
	errCh := make(chan error, 1)
	events := make(chan interface{}, 8)
	start := time.Now()

	go func() {
		var err error
		defer close(events)
		res, err = client.Unixfs().Add(ctx, file, caopts.Unixfs.Pin(*pin), caopts.Unixfs.Progress(true), caopts.Unixfs.Events(events))
		errCh <- err
	}()

	for event := range events {
		output, ok := event.(*coreiface.AddEvent)
		if !ok {
			panic("unknown event type")
		}

		if output.Path != nil && output.Name != "" {
			_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("Added %v", output.Name))
		}
	}

	if err := <-errCh; err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		exit(start, -1)
	}

	_, _ = fmt.Fprintln(os.Stdout, res.Cid().String())
	exit(start, 0)
}

func exit(start time.Time, exitCode int) {
	duration := time.Since(start)
	_, _ = fmt.Fprintln(os.Stderr, duration)
	os.Exit(exitCode)
}

func basicAuth(projectId, projectSecret string) string {
	auth := projectId + ":" + projectSecret
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
