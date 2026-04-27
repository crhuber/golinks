
# Golinks

Golinks is a lightweight, self-hosted URL shortener that turns short keywords into full URLs. Type `go/keyword` in your browser and get redirected instantly — no external services required.

![alt text](screenshot.jpg “Screenshot”)

## Quickstart

Get up and running in under a minute:



Go to the [releases](https://github.com/crhuber/golinks/releases) page and download the latest binary for your platform.

Or install with [kelp](https://github.com/crhuber/kelp):


```bash
# 1. Download the latest binary from the releases page (or use kelp)
kelp add crhuber/golinks --install

# 2. Start the server
golinks serve

# 3. Open in your browser
open http://localhost:8998
```

That’s it. Golinks uses SQLite by default and stores everything at `~/.golinks/golinks.db` (the directory is created automatically).

For the best experience, install the [Chrome extension](#browser-extension-recommended) so you can type `go/keyword` directly in your address bar.

## About

Inspired by Google’s internal golink system — where Benjamin Staffin introduced the “go/” domain so Googlers could use shortlinks like “go/link” (described as “AOL keywords for the corporate network”).

I built this to scratch my own itch and to learn Go. It runs locally on localhost with a SQLite database, designed to be lightweight and simple. Inspired by @thesephist’s [tools](https://thesephist.com/posts/tools/) and the concept of [building software for yourself](https://changelog.com/podcast/455). The backend is written in Go, the frontend in Vue.js as a single page app.

## Installation

### Binary

Go to the [releases](https://github.com/crhuber/golinks/releases) page and download the latest binary for your platform.

Or install with [kelp](https://github.com/crhuber/kelp):

```bash
kelp add crhuber/golinks --install
```

### Docker

```bash
docker build . -t crhuber/golinks:latest
docker run -p 8998:8998 crhuber/golinks
```

## Configuration

All configuration can be set via CLI flags, environment variables, or a `GOLINKS.toml`/`GOLINKS.yaml` config file in the working directory.

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `-d, --db` | `GOLINKS_DB` | `~/.golinks/golinks.db` | Database DSN or SQLite file path |
| `-t, --dbtype` | `GOLINKS_DBTYPE` | `sqlite` | Database type: `sqlite`, `postgres`, or `mysql` |
| `-p, --port` | `GOLINKS_PORT` | `8998` | Port to run the server on |

### Using a different database

By default, golinks uses SQLite with no setup required. To use PostgreSQL or MySQL instead, set the database type and DSN:

```bash
# MySQL
export GOLINKS_DBTYPE=”mysql”
export GOLINKS_DB=”user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local”

# PostgreSQL
export GOLINKS_DBTYPE=”postgres”
export GOLINKS_DB=”host=localhost user=golinks password=secret dbname=golinks port=5432 sslmode=disable”
```

### Run at startup (macOS)

To start golinks automatically on boot, use the included launchd plist:

```bash
# Edit the plist to set the correct path to the golinks binary
vi io.intra.golinks.plist
cp io.intra.golinks.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/io.intra.golinks.plist
launchctl start io.intra.golinks
```

See [launchd.info](https://www.launchd.info/) for more details. Logs are written to `/tmp/golinks.log`.

## Browser Extension (Recommended)

The easiest way to use golinks is with the included Chrome extension. It lets you type `go/keyword` directly in your address bar — no DNS or port forwarding hacks needed.

### Install

1. Open `chrome://extensions` in Chrome
2. Enable **Developer mode** (toggle in the top right)
3. Click **Load unpacked** and select the `extension/` directory from this repo

### Usage

- **Direct navigation** — Type `go/keyword` in your address bar and press Enter. The extension intercepts this and redirects through your golinks server.
- **Omnibox search** — Type `go` then press <kbd>Tab</kbd> (or <kbd>Space</kbd>) to activate the omnibox. Start typing to see live search suggestions.
- **Popup search** — Click the extension icon to open a quick search popup with live results.

### Extension settings

Right-click the extension icon and select **Options** to configure:

- **Server URL** — Address of your golinks server (default: `http://localhost:8998`)
- **Intercept Prefix** — The keyword that triggers URL interception (default: `go`). Change this to `g`, `link`, `jump`, etc. if you prefer a different shortcut. The omnibox keyword is always `go` regardless of this setting.

## Alternative Setup (Without Extension)

If you prefer not to use the browser extension, you can set up DNS and port forwarding so that `go/keyword` resolves natively in your browser.

<details>
<summary>DNS setup</summary>

**Automatic:** Use [dev-proxy](https://github.com/crhuber/dev-proxy).

**Manual:** Add a host record pointing to your golinks server, then add `.internal` to your search domains:

```bash
# /etc/hosts
127.0.0.2       go.internal
```

System Preferences > Network > Advanced > DNS > Search Domains: add `.internal`

</details>

<details>
<summary>Port redirection (so you don’t need :8998 in the URL)</summary>

**Automatic:** Use [dev-proxy](https://github.com/crhuber/dev-proxy).

**Manual:** Create a loopback alias and a port forwarding rule:

```bash
# Create loopback alias
sudo ifconfig lo0 alias 127.0.0.2

# Persist after reboot
sudo cp io.intra.ifconfig.plist /Library/LaunchDaemons/

# Forward port 80 → 8998
echo “rdr pass inet proto tcp from any to 127.0.0.2 port 80 -> 127.0.0.1 port 8998” | sudo pfctl -ef -
```

To remove: `sudo pfctl -F all -f /etc/pf.conf`

</details>

## Usage

### Creating links

Open `http://localhost:8998` in your browser to view, create, and manage links through the web UI.

### Programmatic links

Create short links that inject variables by using `{*}`. For example, a link with keyword `gh` and URL `https://github.com/{*}` lets you type `go/gh/torvalds` to go directly to `https://github.com/torvalds`.

## Troubleshooting

- If you change the server port, make sure the frontend is also configured to connect to the same port.

## Developing

I use [air](https://github.com/cosmtrek/air) for live reloading during development:

```bash
air
```

## Contributing

If you find bugs, please open an issue first. If you have feature requests, I probably will not honor it because this project is being built mostly to suit my personal workflow and preferences.
