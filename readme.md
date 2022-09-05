
# Golinks

Golinks is an internal URL shortener that organizes your company links into easily rememberable keywords. If you’re on the company network, you can type in <code>go/keyword</code> in your browser, and that will redirect you to the expanded url.


## History of Golinks

Benjamin Staffin at Google developed a golink system that introduced the "go/" domain and allowed Googlers to simply use the shortlink “go/link” in their browser. Benjamin described golinks as "AOL keywords for the corporate network."

## Why

I developed this to scratch my own itch mostly and to learn Go. It was built intending to be run locally on localhost using a sqllite database. It is meant to be lightweight and simple. I was inspired by
@thesephist's [tools](https://thesephist.com/posts/tools/) and the concept of [building software for yourself](https://changelog.com/podcast/455).
The backend API is written in Go and the frontend in Vue.js as a single page app.


## Setup

### Install

Go to the [releases](https://github.com/crhuber/golinks/releases) page and download the latest release.
Or, use my own tool: [kelp](https://github.com/crhuber/kelp)

```bash
kelp add crhuber/golinks
kelp install golinks
```

### Database

Setup a path where you want your golinks sqllite database to live and set the environment variable

```bash
mkdir ~/.golinks
export GOLINKS_DB="/Users/username/.golinks/golinks.db"
```

You can also use postgres or mysql database using a valid DSN like:

```bash
export GOLINKS_DB_TYPE="mysql"
export GOLINKS_DB="user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
```

### Run

Copy the `static` and `static/index.html` folder/files to a location in the same directory as the golinks binary


Run

```bash
golinks serve
```

### Run At Startup
To run as an Agent on boot for mac edit and copy the `io.intra.golinks.plist` file to `~/Library/LaunchAgents`  directory.
See [launchd.info](https://www.launchd.info/)

```
launchctl load ~/Library/LaunchAgents/io.intra.golinks.plist
launchctl start io.intra.golinks
tail -f /tmp/golinks.log
tail -f /private/var/log/com.apple.xpc.launchd/launchd.log
```

### Docker

Build image and run
```
docker build . -t crhuber/golinks:latest
docker run -p 8080:8080 crhuber/golinks
```

### DNS Setup

Add a host record to point to your golinks server.
If running locally,  edit your local hostfile:`sudo nano /etc/hosts`

```bash
127.0.0.1       go.internal
```

Add the host suffix to your search domains.
System Preferences>Network>Advanced>DNS>Search Domains:

```
Search Domains:
.internal
```

### Port Redirection Setup

If you have a local instance of golinks running on your machine, you will need to append the port everytime you want to use golinks in the browser
ie: `go:8080/foo` which is not ideal. To get around this we can run a few hacks.

Create an alias for 127.0.0.2 to point to loopback:

```
sudo ifconfig lo0 alias 127.0.0.2
```

Create a port forwarding rule to forward traffic destined for `127.0.0.2:80` to be redirected to local golinks on port 8080

```
echo "rdr pass inet proto tcp from any to 127.0.0.2 port 80 -> 127.0.0.1 port 8080" | sudo pfctl -ef -
```

Edit hosts file to modify go.internal to point to 127.0.0.2


```bash
127.0.0.2       go.internal
```

Display current port forwarding

```
sudo pfctl -s nat
```

Remove port forwarding

```
sudo pfctl -F all -f /etc/pf.conf
```

## FAQ

* How can I see all the links available

    http://go:8080/


* How do programmatic links work?

    Create short links that inject variables by using `{*}`. For example: `gh/{*}` to link to `https://github.com/{*}`.
    So when a user types `gh/torvalds` the `{*}` will be replaced and the browser will be redirected to `https://github.com/torvalds`

## Troubleshooting

- If you change the port of the API. Be sure that you change the frontend index.html to connect to the same port

## Developing

I use [air](https://github.com/cosmtrek/air) for live reloading Go apps.
Just run

```
> air

watching .
building...
running...
INFO[0000] Starting server on port :8080
```

## Roadmap

- Add CLI interface to adding/removing/searching links from command line
- Browser extension (maybe)
- UI Refactor

## Contributing

If you find bugs, please open an issue first. If you have feature requests, I probably will not honor it because this project is being built mostly to suit my personal workflow and preferences.
