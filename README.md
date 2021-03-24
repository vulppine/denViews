denViews
========

View counter for websites.

Note: this currently **only** works with GCP.
Implementations for SQL and other cloud
providers will come eventually.

Usage
-----

Server:
 - Upload the server as a container to Cloud Run.
 - Run the container with environmental variable
   `HOSTURL` set to your host URL
   - For example: `subdomain.domain.com`.
   
**NOTE:** Plans on making denViews run on other platforms, and
a self hosting solution will eventually be implemented.
   
JS:
 - Upload the JavaScript file somewhere accessable by the end user.
 - Insert a script containing these HTML attributes:
   - `data-denviews-host="[ YOUR DENVIEW SERVER URL HERE ]"`
   - `id="denviews"`

Tool:

Setting up:
 - Download a **service account key** for your
   Google Cloud project.
 - Set the path to it as environmental variable
   `GOOGLE_APPLICATION_CREDENTIALS`.

Tool verbs:
 - `add`
   - Direct path: adds the path to the database.
   - `qurange`: Adds a given path, but with a key/value
     pair that increments from zero to the maximum
     amount given.
     - Example: `add qurange path/to/page key 200`
     - This outputs a set of pages with query `key=0` to `k=200`.
   - `qustring`: Adds a given path, but with a key/value
     pair that auto-adds all the strings following
     the query key.
     - Example: `add qustring path/to/page key foo bar baz`
     - This outputs a set of pages with queries `key=foo`, `key=bar`, and `key=baz`.
 - `get`
   - Direct path: gets the page stats for a specific page.
   
Building
--------

Simply run go build in the respective directories for the
server/tool. You may want to rename the resulting binary
for both if needed.

For example:

 - `go build -o denviews-tool` in the `tool` directory
 - `go build -o denviews-server` in the `server` directory.

Privacy
-------

Hopefully, this *should* be **GDPR** compliant in its current state:

A unique identifier for every client is created whenever the client
connects, but is hashed using SHA-512 with a salt that changes every 
different day. This effectively anonymizes the client, making it
impossible to obtain any distinct information from a user.

Notes
-----

denViews will automatically listen on port **80**. As this is currently
setup for running within a container, this means that all that needs to
occur is connecting to the container (via URL or IP) via any HTTP client.

denViews converts paths from slashes to underscores automatically
(whether this is an implementation detail with GCP or otherwise
will be decided)

So, a URL that looks like this:

`https://foo.bar/foo/bar/baz`

will be queried as this:

`foo.bar/pageviews/_foo_bar_baz`

in Google Cloud Firestore.

Equally, paths that end in slashes or index.html are automatically
truncinated down to previous path element, to avoid confusion between
three different representations of the same page:

If `foo/bar/baz` is a directory with an index.html file:

```
https://foo.bar/foo/bar/baz
https://foo.bar/foo/bar/baz/
https://foo.bar/foo/bar/baz/index.html
```

will usually all lead to the exact same page when a web
browser points to it. This is entirely dependent on
your web server configuration.

`_` is a reserved path for tracking views at the root of your webpage.
To add it, just add `_` using the tool.

License
-------

Flipp Syder, MIT License, 2021
