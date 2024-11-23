package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/karalabe/go-bluesky"
	"mvdan.cc/xurls/v2"
)

type FacetIndex struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}

type FacetFeature struct {
	Type string `json:"$type"`
	URI  string `json:"uri"`
}

type Facet struct {
	Index    FacetIndex     `json:"index"`
	Features []FacetFeature `json:"features"`
}

type Record struct {
	Type      string    `json:"$type"`
	Text      string    `json:"text"`
	Facets    []Facet   `json:"facets,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

func facets(msg string) []Facet {
	regex := xurls.Strict()
	urls := regex.FindAllString(msg, -1)
	if len(urls) == 0 {
		return nil
	}
	facets := make([]Facet, len(urls))
	for idx, url := range urls {
		start := bytes.Index([]byte(msg), []byte(url))
		end := start + len([]byte(url))
		facets[idx] = Facet{
			Index: FacetIndex{
				ByteStart: start,
				ByteEnd:   end,
			},
			Features: []FacetFeature{{
				Type: "app.bsky.richtext.facet#link",
				URI:  url,
			}},
		}
	}
	return facets
}

func record(msg string) (*util.LexiconTypeDecoder, error) {
	r := Record{
		Type:      "app.bsky.feed.post",
		Text:      msg,
		Facets:    facets(msg),
		CreatedAt: time.Now(),
	}
	rJSON, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	log.Println(string(rJSON))
	var result util.LexiconTypeDecoder
	err = result.UnmarshalJSON(rJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}
	return &result, nil
}

func post(as string, msg string) func(api *xrpc.Client) error {
	return func(api *xrpc.Client) error {
		record, err := record(msg)
		if err != nil {
			return fmt.Errorf("failed to create record: %w", err)
		}

		cri := &atproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.post",
			Repo:       as,
			Record:     record,
		}
		cro, err := atproto.RepoCreateRecord(context.Background(), api, cri)
		log.Printf("%+v\n", cro)
		return err
	}
}

func msg() (string, error) {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func blueskyMessage(bskyUser, bskyAppKey, msgText string) error {
	ctx := context.Background()

	client, err := bluesky.Dial(ctx, bluesky.ServerBskySocial)
	if err != nil {
		return fmt.Errorf("failed to connect to bluesky: %w", err)
	}
	defer client.Close()

	err = client.Login(ctx, bskyUser, bskyAppKey)
	switch {
	case errors.Is(err, bluesky.ErrMasterCredentials):
		return fmt.Errorf("appkey required: %w", err)
	case errors.Is(err, bluesky.ErrLoginUnauthorized):
		return fmt.Errorf("not authorized: %w", err)
	case err != nil:
		return fmt.Errorf("failed to log in: %w", err)
	}

	profile, err := client.FetchProfile(ctx, bskyUser)
	if err != nil {
		return fmt.Errorf("failed to fetch profile: %w", err)
	}

	if err := client.CustomCall(post(profile.DID, msgText)); err != nil {
		return fmt.Errorf("failed to post: %w", err)
	}
	return nil
}

func main() {
	bskyUser := os.Getenv("BSKY_USERNAME")
	bskyAppKey := os.Getenv("BSKY_APPKEY")

	msg, err := msg()
	if err != nil {
		log.Fatalf("failed to read message from stdin: %v", err)
	}

	err = blueskyMessage(bskyUser, bskyAppKey, msg)
	if err != nil {
		log.Fatal(err)
	}
}
