package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wgoodall01/smssh/util"

	"github.com/chzyer/readline"
)

// MaxMessageLength is the maximum length of a single message
var MaxMessageLength = 1000

func main() {
	// Get the query from argv
	query := strings.Join(os.Args[1:], " ")

	// Get the article text extract
	article, err := GetWikiArticle(query)
	util.Fatal("couldn't get article", err)

	rl, err := readline.New("")
	defer rl.Close()
	util.Fatal("could not initialize readline", err)

	// Show the title
	Send(article.Title)

	// Chunk up the article text
	chunks := SplitChunks(article.Extract)

Paginate:
	for _, chunk := range chunks {
		Send(chunk)

	Input:
		for {
			Send("[ next:n quit:q ]")
			line, err := rl.Readline()
			util.Fatal("error calling Readline", err)

			// normalize line
			line = strings.ToLower(line)
			line = strings.TrimSpace(line)

			// Quit
			if line == "q" {
				break Paginate
			} else if line == "n" {
				continue Paginate
			} else {
				continue Input
			}
		}
	}

	Send("[ Done ]")
}

var printTicker = time.Tick(2 * time.Second)

// Send sends the text as an SMS (through stdout), waiting 1s for each message.
func Send(text string) {
	<-printTicker
	fmt.Println(text)
}

func SplitChunks(text string) []string {
	lines := strings.Split(text, "\n")
	// TODO: break up lines longer than 1000 chars.
	return lines
}
