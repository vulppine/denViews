package main

import (
	"context"
	"errors"
	"net/url"
	"log"
	"os"
	"path"
	"strings"
	"strconv"

	"cloud.google.com/go/firestore"
)

var conn *firestore.Client
var ctx context.Context

func AddPage(p string) error {
	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}

	v("attempting to create page collection representation")
	h := strings.Replace(path.Join(p), "/", "_", -1)
	rec := conn.Collection(path.Join(host, "pageviews", h))
	v(rec)

	v("creating stats doc")
	_, err := rec.Doc("stats").Create(ctx, map[string]interface{}{
		"views": 0,
		"hits": 0,
	})
	if err != nil {
		log.Println(err)
	}

	v("creating visitors doc")
	_, err = rec.Doc("visitors").Create(ctx, map[string]interface{}{
		"addrs": map[string]interface{}{},
	})
	if err != nil {
		log.Println(err)
	}

	return nil
}

func AddPageQueryRange(p string, key string, max int) {
	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}

	for i := 0; i < max; i++ {
		u := new(url.URL)
		v := make(url.Values)
		u.Path = p

		v.Add(key, strconv.Itoa(i))
		u.RawQuery = v.Encode()

		AddPage(u.String())
	}
}

func AddPageQueryStrings(p string, key string, strs ...string) {
	if p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}

	for _, s := range strs {
		u := new(url.URL)
		v := make(url.Values)
		u.Path = p

		v.Add(key, s)
		u.RawQuery = v.Encode()

		AddPage(u.String())
	}
}

type Retrievee func(*Record, error)

func GetPageStats(p string) (*Record, error) {
	r := new(Record)
	u, err := url.Parse(p)
	if err != nil { return nil, err }

	// one length path implies website root
	if len(u.Path) > 1 {
		if a := strings.Split(u.Path, "/") ; a[len(a)-1] == "index.html" || a[len(a)-1] == "" {
			a = a[:len(a)-1]
			u.Path = strings.Join(a, "/")
		}
	}

	r.rawPath = path.Join(host, "pageviews", strings.ReplaceAll(u.String(), "/", "_"))

	d, err := conn.Collection(r.rawPath).Doc("stats").Get(ctx)

	log.Println(d.ReadTime)
	if err != nil { return nil, err }

	var a map[string]interface{}
	if a = d.Data() ; a == nil { return nil, errors.New("could not get record") }
	r.Views = int(a["views"].(int64))
	r.Hits = int(a["hits"].(int64))

	return r, nil
}

func init() {
	var err error

	ctx = context.Background()
	conn, err = firestore.NewClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		panic(err)
	}
}
