package main

import (
	"fmt"
	"os"
  "regexp"
	rss "github.com/jteeuwen/go-pkg-rss"
  "strings"
	"time"
)

type flag int

const (
	Unset flag = iota
	True
	False
)


var newState chan flag

func main() {
  oldState := Unset // TODO read old state from disk
  //newState := make(chan flag)

	// This sets up a new feed and polls it for new channels/items.
	// Invoke it with 'go PollFeed(...)' to have the polling performed in a
	// separate goroutine, so you can continue with the rest of your program.
	go PollFeed("http://feeds.gothamistllc.com/gothamist05", 5)

  for {
    fmt.Printf("wait state\n")
    state := <-newState // block until we get a drudge siren
    if oldState != state {
      // do update, write state to disk
      fmt.Printf("%d state\n", state)
    }
    oldState := state
    fmt.Printf("%d ostate\n", oldState)
  }
}

func PollFeed(uri string, timeout int) {
	feed := rss.New(timeout, true, chanHandler, itemHandler)

	for {
		if err := feed.Fetch(uri, nil); err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s: %s", uri, err)
			return
		}

		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
	}
}

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	fmt.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)

  if  len(newitems) < 5 { 
    return // Let's not point fingers here
  }

  for _,item := range newitems {
    //fmt.Printf("item %s\n",item.Title)
    if m, _ := regexp.MatchString("hipster", strings.ToLower(item.Title)); m == true {
      fmt.Printf("HIPSTER!!!! %s\n",item.Title)
      newState<-True
    }
  }
  newState<-False
}
