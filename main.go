package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"gopkg.in/yaml.v2"
)

func fatal(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	os.Exit(1)
}

func e(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, f+"\n", args...)
}

type config struct {
	Oauth struct {
		ConsumerKey    string `yaml:"consumer_api_key"`
		ConsumerSecret string `yaml:"consumer_api_secret_key"`
		AccessToken    string `yaml:"access_token"`
		AccessSecret   string `yaml:"access_token_secret"`
	} `yaml:"oauth"`
	List struct {
		Owner string `yaml:"owner"`
		Name  string `yaml:"name"`
	} `yaml:"list"`
	Delay     int `yaml:"delay"`
	Interval  int `yaml:"interval"`
	Threshold int `yaml:"threshold"`
}

type stats struct {
	tweet      int64
	popularity int
}

func main() {
	configPath := flag.String("config", "creme-rt.yml", "config file path")
	flag.Parse()

	f, err := os.Open(*configPath)
	if err != nil {
		fatal("error: open config file: %s", err.Error())
		return
	}

	var config config
	err = yaml.NewDecoder(f).Decode(&config)
	if err != nil {
		fatal("error: decode config file: %s", err.Error())
		return
	}

	oauthConfig := oauth1.NewConfig(config.Oauth.ConsumerKey, config.Oauth.ConsumerSecret)
	token := oauth1.NewToken(config.Oauth.AccessToken, config.Oauth.AccessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	list, _, err := client.Lists.Show(&twitter.ListsShowParams{
		OwnerScreenName: config.List.Owner,
		Slug:            config.List.Name,
	})
	if err != nil {
		fatal("error: get list: %s", err.Error())
		return
	}
	listId := list.ID

	var id int64 = 0
outer:
	for {
		now := time.Now()
		if id == 0 {
			for {
				tweets, _, err := client.Lists.Statuses(&twitter.ListsStatusesParams{
					ListID: listId,
					Count:  1000000,
					MaxID:  id,
				})
				if err != nil {
					e("error: get list statuses: %s", err.Error())
					time.Sleep(1 * time.Minute)
					id = 0
					continue outer
				}
				if len(tweets) == 0 {
					time.Sleep(time.Duration(config.Interval) * time.Minute)
					id = 0
					continue outer
				}
				ti, err := tweets[len(tweets)-1].CreatedAtTime()
				if err != nil {
					e("error: get list status time: %s", err.Error())
					time.Sleep(time.Duration(config.Interval) * time.Minute)
					id = 0
					continue outer
				}
				if !ti.Add(time.Duration(config.Delay+config.Interval) * time.Minute).Before(now) {
					id = tweets[len(tweets)-1].ID
					continue
				}
				for i := len(tweets) - 1; i >= 0; i-- {
					t := tweets[i]
					id = t.ID
					ti, err := t.CreatedAtTime()
					if err != nil {
						e("error: get list status time: %s", err.Error())
						time.Sleep(time.Duration(config.Interval) * time.Minute)
						id = 0
						continue outer
					}
					if ti.Add(time.Duration(config.Delay+config.Interval) * time.Minute).After(now) {
						break
					}
				}
				break
			}
		}
		best := make(map[int64]stats)
	inner:
		for {
			tweets, _, err := client.Lists.Statuses(&twitter.ListsStatusesParams{
				ListID:  listId,
				Count:   1000000,
				SinceID: id - 1,
			})
			if err != nil {
				e("error: get list statuses: %s", err.Error())
				time.Sleep(1 * time.Minute)
				id = 0
				continue outer
			}
			if len(tweets) == 0 {
				time.Sleep(time.Duration(config.Interval) * time.Minute)
				id = 0
				continue outer
			}
			for i := len(tweets) - 1; i >= 0; i-- {
				t := tweets[i]
				ti, err := t.CreatedAtTime()
				if err != nil {
					e("error: get list status time: %s", err.Error())
					time.Sleep(time.Duration(config.Interval) * time.Minute)
					id = 0
					continue outer
				}
				if ti.Add(time.Duration(config.Delay) * time.Minute).After(now) {
					id = t.ID
					break inner
				}
				popularity := t.RetweetCount + t.FavoriteCount
				if popularity < config.Threshold {
					continue
				}
				if b, ok := best[t.User.ID]; (ok && b.popularity <= popularity) || !ok {
					best[t.User.ID] = stats{
						tweet:      t.ID,
						popularity: popularity,
					}
				}
			}
			id = tweets[0].ID + 1
			if len(tweets) == 1 {
				break inner
			}
		}
		if len(best) > 0 {
			b := make([]int64, len(best))
			p := rand.Perm(len(best))
			i := 0
			for _, s := range best {
				b[p[i]] = s.tweet
				i++
			}
			go func() {
				for _, t := range b {
					if _, _, err := client.Statuses.Retweet(t, nil); err != nil {
						e("error: retweet status: %s", err.Error())
					}
					time.Sleep(time.Duration(config.Interval) * time.Minute / time.Duration(len(b)))
				}
			}()
		}
		time.Sleep(time.Duration(config.Interval) * time.Minute)
	}
}
