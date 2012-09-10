package main

import (
	"fmt"
	"gh-pages-publish"
	rss "github.com/jteeuwen/go-pkg-rss"
	"html/template"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"
)

var newState chan []*rss.Item
var output *githubPagesPublish.Publisher

const templateName = "dieHipster.html"
const margin = 5 // margin of safety
const limit = 5  // number of items to show

func main() {
	oldState := make([]*rss.Item, 1) // TODO read old state from disk
	newState = make(chan []*rss.Item)
	var err error
	output, err = githubPagesPublish.New("git@github.com:WIZARDISHUNGRY/test-pages.git", "gh-pages")
	defer output.Close()
	if err != nil {
		panic(err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			fmt.Fprintf(os.Stderr, "Got signal %s\n", sig)
			output.Close()
			return
		}
	}()

	// This sets up a new feed and polls it for new channels/items.
	// Invoke it with 'go PollFeed(...)' to have the polling performed in a
	// separate goroutine, so you can continue with the rest of your program.
	go PollFeed("http://feeds.gothamistllc.com/gothamist05", 30)

	var state []*rss.Item = nil
	for {
		//fmt.Printf("wait state\n")
		state = <-newState // block until we get a drudge siren
		if len(oldState) != len(state) || func(state, oldState []*rss.Item) bool {
			for _, item := range state {
				fmt.Printf("1st order %s\n", item.Title)
			}
			return true
		}(oldState, state) {
			writeHtml(state)
		}
		fmt.Printf("%d hipsters dancing on the head of a pin!\n", len(state))
		oldState = state
	}
}

func PollFeed(uri string, timeout int) {
	feed := rss.New(timeout, false, chanHandler, itemHandler)

	for {
		if err := feed.Fetch(uri, nil); err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s: %s", uri, err)
			return
		}

		<-time.After(time.Duration(feed.SecondsTillUpdate()))
	}
}

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	fmt.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)

	if len(newitems) <= margin {
		return // Let's not point fingers here!
	}

	items := []*rss.Item{}
	for _, item := range newitems {
		//fmt.Printf("item %s\n",item.Title)
		if m, _ := regexp.MatchString("hipster", strings.ToLower(item.Title)); m == true {
			fmt.Printf("HIPSTER!!!! %s\n", item.Title)
			if len(items) < limit {
				items = append(items, item)
			}
		}
	}
	newState <- items
}

func writeHtml(items []*rss.Item) {
	t := template.New(templateName)
	t, err := t.ParseFiles(templateName)
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(output.Path+"/index.html", os.O_WRONLY|os.O_CREATE, 0644)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	err = t.Execute(file, items)
	file.Close()
	if err != nil {
		panic(err)
	}
	output.Push()
}
