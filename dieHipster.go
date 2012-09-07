package main

import (
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	"html/template"
	"os"
	"regexp"
	"strings"
	"time"
)

var newState chan []*rss.Item

const templateName = "dieHipster.html"
const margin = 5 // margin of safety
const limit = 5  // number of items to show

func main() {
	oldState := make([]*rss.Item, 0) // TODO read old state from disk
	newState = make(chan []*rss.Item)

	// This sets up a new feed and polls it for new channels/items.
	// Invoke it with 'go PollFeed(...)' to have the polling performed in a
	// separate goroutine, so you can continue with the rest of your program.
	go PollFeed("http://feeds.gothamistllc.com/gothamist05", 30)

	var state []*rss.Item = nil
	for {
		//fmt.Printf("wait state\n")
		state = <-newState               // block until we get a drudge siren
		if len(oldState) != len(state) { // TODO add || uid!=uid
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
	err = t.Execute(os.Stdout, items)
	if err != nil {
		panic(err)
	}
}
