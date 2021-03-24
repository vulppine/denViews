package main

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var host *url.URL

/*
// ip sorting
type IPAddrs []*netaddr.IP

func (p IPAddrs) Len() int { return len(p) }
func (p IPAddrs) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p IPAddrs) Less(i, j int) bool {
	return p[i].Compare(*p[j]) < 0
}

// parsing string array to IPSet

func toIPSet(s []string) *netaddr.IPSet {
	p := new(netaddr.IPSetBuilder)

	for _, v := range s {
		i, err := netaddr.ParseIP(v)

		if err == nil {
			p.Add(i)
		}
	}

	return p.IPSet()
}
*/

type Record struct {
	rawPath      string
	Views        int `json:"views"`
	Hits         int `json:"hits"`
	validQueries []string
	// ipaddrs      *netaddr.IPSet
}

type Counter string

const (
	Views Counter = "views"
	Hits  Counter = "hits"
)

func queryViews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	d, err := Retrieve(r.Context(), r.URL)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	b, err := json.Marshal(d)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", host.String())
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

	// do the update after sending as this is the most
	// expensive part of the program, doing it before
	// sending the response would make everything take
	// too long
	d.Update(r)
}

func (r *Record) Update(q *http.Request) {
	var err error
	/*
	if !sort.StringsAreSorted(r.rawaddrs) { // just in case
		sort.Strings(r.rawaddrs)
	}
	*/


	var i string
	var f bool
	if v, b := q.Header["X-Forwarded-For"] ; b {
		i = v[0]
		f = true
	} else {
		i = q.RemoteAddr
	}

	if !f {
		if q.RemoteAddr[0] == '[' {
			i = strings.Split(i, "]")[0][1:] // HACKY - gets IPv6 address
		} else {
			i = strings.Split(i, ":")[0] // gets IPv4 address
		}
	}


	h := sha512.New()
	defer h.Reset()

	h.Write([]byte(SaltString(q.Context(), i + q.Header.Get("User-Agent"))))
	i = fmt.Sprintf("%x", h.Sum(nil))

	v, err := r.UniqueVisitor(q.Context(), i)
	if err != nil { log.Println(err) }

	if v && err == nil {
		err = r.Increment(q.Context(), Views)
		if err != nil { log.Println(err) }

		err = r.AddVisitor(q.Context(), i)
		if err != nil { log.Println(err) }
	}

	/*
	if p := sort.SearchStrings(r.rawaddrs, i); p == len(r.rawaddrs)-1 && r.rawaddrs[p] != r.rawaddrs[len(r.rawaddrs)-1] {
		r.rawaddrs = append(r.rawaddrs, i)
		sort.Strings(r.rawaddrs)
		r.Increment(Views)
	}
	*/

	err = r.Increment(q.Context(), Hits)
	if err != nil { log.Println(err) }
}

type Salt struct {
	salt    string
	updated time.Time
}

func SaltString(ctx context.Context, s string) string {
	t, err := GetSalt(ctx)
	if err != nil { log.Println(err) }

	if t.updated.Day() != time.Now().Day() { // a new salt every day, in case one hasn't already been made
		t = NewSalt(ctx)
		SetSalt(ctx, t)
	}

	return s + t.salt
}

func NewSalt(ctx context.Context) *Salt {
	i, err := rand.Int(rand.Reader, big.NewInt(512256128))
	if err != nil {
		log.Println(err)
	}

	s := &Salt{ salt: fmt.Sprintf("%x", md5.Sum([]byte(i.String()))), updated: time.Now() }

	return s
}

func init() {
	var err error

	host, err = url.Parse(os.Getenv("HOSTURL"))
	if err != nil {
		panic(err)
	}
}
