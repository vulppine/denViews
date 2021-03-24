package main

// This is meant for use with
// Google Cloud Platform.
// Please note: to avoid mass querying of invalid
// addresses, you **must** initialize the database
// manually, and pass in any valid query strings
// as well.

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"log"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

var conn *firestore.Client

// Retrieve retrieves a record from the Firestorm storage
// relative to the project ID. If the HOSTURL env var
// is defined (and it should be, if this is being
// run off of GCP), it will join the host name with
// the path retrieved in order to obtain it from the
// collection under the host name.
// index.html is removed in order to prevent duplicate entries
// for when a directory is intended to be accessed as a webpage
// (e.g., https://foo.bar/directory/index.html should be
// accessed the exact same as https://foo.bar/directory)

func Retrieve(ctx context.Context, p *url.URL) (*Record, error) {
	r := new(Record)

	// one length path implies website root
	if len(p.Path) > 1 {
		if a := strings.Split(p.Path, "/") ; a[len(a)-1] == "index.html" || a[len(a)-1] == "" {
			a = a[:len(a)-1]
			p.Path = strings.Join(a, "/")
		}
	}

	r.rawPath = path.Join(host.Host, "pageviews", strings.ReplaceAll(p.String(), "/", "_"))

	d, err := conn.Collection(r.rawPath).Doc("stats").Get(ctx)
	if err != nil { return nil, err }

	var a map[string]interface{}

	if a = d.Data() ; a == nil { return nil, errors.New("could not get record") }
	r.Views = int(a["views"].(int64))
	r.Hits = int(a["hits"].(int64))

	return r, nil
}

// UniqueVisitor uses the 'does this exist in a map' optional
// var in order to check if an ip has already visited the
// given page associated with the record.
//
// How fast this is, is dependent on Firestorm's DataAt().

func (r *Record) UniqueVisitor(ctx context.Context, i string) (bool, error) {
	d, err := conn.Collection(r.rawPath).Doc("visitors").Get(ctx)
	if err != nil { return false, err }
	v, err := d.DataAt("addrs")
	if err != nil { return false, err }
	_, b := v.(map[string]interface{})[i]

	return !b, nil
}

func (r *Record) AddVisitor(ctx context.Context, i string) error {
	rec := conn.Collection(r.rawPath).Doc("visitors")
	err := conn.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		d, err := tx.Get(rec)
		if err != nil { return err }
		p, err := d.DataAt("addrs")
		if err != nil { return err }

		p.(map[string]interface{})[i] = true

		return tx.Set(rec, map[string]interface{}{
			"addrs": p,
		}, firestore.MergeAll)
	})
	if err != nil {
		return err
	}

	return nil
}

// Increment runs a single transaction, grabbing the given counter
// and incrementing it by one.

func (r *Record) Increment(ctx context.Context, c Counter) error {
	rec := conn.Collection(r.rawPath).Doc("stats")
	err := conn.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// d, err := tx.Get(rec)
		// if err != nil { return err }
		// i, err := d.DataAt(string(c))
		// if err != nil { return err }

		return tx.Update(rec, []firestore.Update{
			{ Path: string(c), Value: firestore.Increment(1) },
		})
	})
	if err != nil {
		return err
	}

	return nil
}

func GetSalt(ctx context.Context) (*Salt, error) {
	s := new(Salt)
	d, err := conn.Collection(host.Host).Doc("salt").Get(ctx)

	if !d.Exists() {
		_, err = conn.Collection(host.Host).Doc("salt").Create(ctx, map[string]interface{}{})
		if err != nil { return nil, err }

		d, err = conn.Collection(host.Host).Doc("salt").Get(ctx)
		if err != nil { return nil, err }

		err = SetSalt(ctx, NewSalt(ctx))
		if err != nil { return nil, err }
	}

	if err != nil { return nil, err }

	a := d.Data()
	if err != nil { return nil, err }
	s.salt = a["salt"].(string)
	s.updated = a["updated"].(time.Time)

	return s, nil
}

func SetSalt(ctx context.Context, s *Salt) error {
	log.Println("making a new salt")
	rec := conn.Collection(host.Host).Doc("salt")
	err := conn.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return tx.Update(rec, []firestore.Update{
			{ Path: "salt", Value: s.salt },
			{ Path: "updated", Value: s.updated },
		})
	})
	if err != nil {
		return err
	}

	return nil
}

func getProjectID() string {
	c := new(http.Client)
	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	if err != nil { panic(err) }
	req.Header.Add("Metadata-Flavor", "Google")

	r, err := c.Do(req)
	if err != nil { panic(err) }

	i, err := io.ReadAll(r.Body)
	if err != nil { panic(err) }

	return string(i)
}

func init() {
	var err error

	conn, err = firestore.NewClient(context.Background(), getProjectID())
	if err != nil {
		panic(err)
	}
}
