package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	ipfsFiles "github.com/ipfs/go-ipfs-files"
	httpapi "github.com/ipfs/go-ipfs-http-client"
	caopts "github.com/ipfs/interface-go-ipfs-core/options"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const infuraAPI = "https://ipfs.infura.io:5001"

func main() {
	// read configuration file
	viper.SetDefault("PIN", "true")
	viper.SetDefault("URL", "https://ipfs.infura.io:5001")
	viper.SetConfigName(".infura-ipfs-upload-client")
	viper.SetConfigType("env")
	viper.AddConfigPath("$HOME")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error parsing config file: %w", err))
		}
	}

	flag.String("id", "", "your Infura ProjectID")
	flag.String("secret", "", "your Infura ProjectSecret")
	flag.String("url", infuraAPI, "the API URL")
	flag.Bool("pin", true, "whether or not to pin the data")

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	projectId := viper.GetString("id")
	projectSecret := viper.GetString("secret")
	api := viper.GetString("url")
	pin := viper.GetBool("pin")

	if projectId == "" {
		_, _ = fmt.Fprintln(os.Stderr, "parameter --id is required")
		os.Exit(-1)
	}
	if projectSecret == "" {
		_, _ = fmt.Fprintln(os.Stderr, "parameter --secret is required")
		os.Exit(-1)
	}

	httpClient := &http.Client{}
	client, err := httpapi.NewURLApiWithClient(api, httpClient)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	client.Headers.Add("Authorization", "Basic "+basicAuth(projectId, projectSecret))

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

	resolved, err := client.Unixfs().Add(ctx, file, caopts.Unixfs.Pin(pin), caopts.Unixfs.Progress(true))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	fmt.Println(resolved)
}

func basicAuth(projectId, projectSecret string) string {
	auth := projectId + ":" + projectSecret
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
