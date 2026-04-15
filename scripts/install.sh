#!/bin/bash
set -euo pipefail

echo "Building track-..."
cd "$(dirname "$0")/.."
go build -o track- .

echo "Installing binary to /usr/local/bin/track-..."
sudo cp track- /usr/local/bin/track-
rm track-

echo "Installing launchd agent (daily update at 8 AM)..."
cp com.track-.update.plist ~/Library/LaunchAgents/com.track-.update.plist

# Unload first if already loaded (ignore errors)
launchctl unload ~/Library/LaunchAgents/com.track-.update.plist 2>/dev/null || true
launchctl load ~/Library/LaunchAgents/com.track-.update.plist

echo ""
echo "Done! You can now:"
echo "  1. Run 'track- setup' to configure your USPS API credentials"
echo "  2. Run 'track-' to start tracking packages"
echo ""
echo "Packages will auto-update daily at 8:00 AM."
