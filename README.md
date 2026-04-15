# track-

A beautiful terminal UI for tracking USPS packages on macOS. Built with Go, [Bubbletea](https://github.com/charmbracelet/bubbletea), and SQLite.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![macOS](https://img.shields.io/badge/macOS-supported-000000?logo=apple&logoColor=white)
![License](https://img.shields.io/github/license/Subhrato20/track-)

```
╭──────────────────────────────────────────────────╮
│  USPS Package Tracker                            │
╰──────────────────────────────────────────────────╯

  STATUS       TRACKING #              NICKNAME        UPDATED
  ● Delivered  9400111899223456789     Mom's gift      Apr 12
  ◐ In Transit 9200190162397312345     New laptop      Apr 14
  ○ Pre-Transit 9405511206217654321    Return item     Apr 13

↑/↓ navigate  a add  d delete  enter details  r refresh  q quit
```

## Features

- Beautiful colored TUI with status icons
- Local SQLite database — your data stays on your machine
- Add, view, refresh, and delete tracked packages
- Scrollable tracking history with timestamps and locations
- Auto-updates all packages daily at 8 AM (via macOS launchd)
- Works from any terminal — just type `track-`

## Install

### Option 1: `go install` (recommended)

```bash
go install github.com/Subhrato20/track-@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `PATH`:

```bash
# Add to your ~/.zshrc or ~/.bashrc if not already there
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Option 2: Clone and build

```bash
git clone https://github.com/Subhrato20/track-.git
cd track-
go build -o track- .
sudo cp track- /usr/local/bin/
```

### Option 3: Full install with auto-updates

```bash
git clone https://github.com/Subhrato20/track-.git
cd track-
bash scripts/install.sh
```

This installs the binary to `/usr/local/bin/` **and** sets up a launchd agent to auto-refresh your packages daily at 8 AM.

## Setup

### 1. Get USPS API credentials (free)

1. Go to [developers.usps.com](https://developers.usps.com) and create an account
2. Click **"Apps"** → **"Create App"**
3. Select the **Package Tracking** API
4. After approval, copy your **Consumer Key** and **Consumer Secret**

### 2. Configure track-

```bash
track- setup
```

Enter your Consumer Key and Consumer Secret when prompted. That's it.

> Config is stored at `~/.config/track-/config.json`. Database at `~/.config/track-/track.db`.

## Usage

```bash
track-              # Launch the TUI
track- setup        # Configure USPS API credentials
track- update       # Manually refresh all packages (headless)
track- version      # Print version
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `a` | Add a new package |
| `enter` | View tracking history |
| `d` | Delete a package |
| `r` | Refresh all packages |
| `↑/↓` or `j/k` | Navigate |
| `tab` | Switch fields (in add form) |
| `esc` | Go back |
| `q` | Quit |

## How it works

- **Database**: All tracking data is stored locally in SQLite at `~/.config/track-/track.db`
- **API**: Uses the [USPS Package Tracking API v3](https://developers.usps.com) with OAuth 2.0
- **Auto-update**: The install script sets up a macOS launchd agent that runs `track- update` daily at 8 AM, refreshing all non-delivered packages
- **No CGO**: Uses pure-Go SQLite (`modernc.org/sqlite`), so `go install` works without a C compiler

## Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/track-

# Remove launchd agent
launchctl unload ~/Library/LaunchAgents/com.track-.update.plist
rm ~/Library/LaunchAgents/com.track-.update.plist

# Remove config and database
rm -rf ~/.config/track-
```

## License

MIT
