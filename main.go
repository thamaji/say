package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/hajimehoshi/oto"
	"golang.org/x/crypto/ssh/terminal"
)

var Dockerfile = `
FROM ubuntu:16.04

RUN set -x \
  && apt-get update \
  && apt-get install -y \
    open-jtalk \
    hts-voice-nitech-jp-atr503-m001 \
    open-jtalk-mecab-naist-jdic \
  && rm -rf /var/lib/apt/lists/*

RUN set -x \
  && apt-get update \
  && apt-get install -y curl unzip \
  && rm -rf /var/lib/apt/lists/*

RUN set -x \
  && curl -fsSLO "https://downloads.sourceforge.net/project/mmdagent/MMDAgent_Example/MMDAgent_Example-1.7/MMDAgent_Example-1.7.zip" \
  && unzip MMDAgent_Example-1.7.zip \
  && mv MMDAgent_Example-1.7/Voice/* /usr/share/hts-voice/ \
  && rm -rf MMDAgent_Example-1.7.zip MMDAgent_Example-1.7

ENV VOICE nitech-jp-atr503-m001/nitech_jp_atr503_m001

RUN echo 'exec open_jtalk -x /var/lib/mecab/dic/open-jtalk/naist-jdic -m /usr/share/hts-voice/${VOICE}.htsvoice -ow /dev/stdout' > /entrypoint.sh

ENTRYPOINT ["bash", "-eu", "/entrypoint.sh"]
`

func build(tag string) error {
	cmd := exec.Command("docker", "build", "-t", tag, "-")
	cmd.Stdin = strings.NewReader(Dockerfile)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
	return cmd.Run()
}

func run(tag string, voice string, message io.Reader) error {
	player, err := oto.NewPlayer(44100, 1, 2, 65536)
	if err != nil {
		return err
	}
	defer player.Close()

	cmd := exec.Command("docker", "run", "-i", "--rm", "-e", "VOICE="+voice, tag)
	cmd.Stdin = message
	cmd.Stdout = player
	cmd.Stderr = ioutil.Discard
	return cmd.Run()
}

func showHelp(output io.Writer) {
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Usage: "+os.Args[0]+" TEXT [TEXT...]")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Japanese text speech CLI")
	fmt.Fprintln(output)
	flag.CommandLine.PrintDefaults()
	fmt.Fprintln(output)
}

func showVersion(output io.Writer) {
	fmt.Fprintln(output, "v1.0.0")
}

func main() {
	var help, version bool

	flag.BoolVar(&help, "h", false, "show help")
	flag.BoolVar(&version, "v", false, "show version")
	flag.Parse()

	if help {
		showHelp(os.Stdout)
		return
	}

	if version {
		showVersion(os.Stdout)
		return
	}

	args := flag.Args()

	tag := "thamaji/say:latest"

	if err := build(tag); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	var reader io.Reader = strings.NewReader(strings.Join(args, " "))
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		reader = io.MultiReader(reader, os.Stdin)
	} else if len(args) <= 0 {
		showHelp(os.Stdout)
		return
	}

	if err := run(tag, "mei/mei_normal", reader); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
