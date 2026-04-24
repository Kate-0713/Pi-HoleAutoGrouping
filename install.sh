#!/bin/bash

if [ "$(id -u)" -ne 0 ]; then
  echo "This script must be run as root" >&2
  exit 1
fi

if [ "$(dpkg -l | awk '/python3-venv/ {print }' |wc -l)" -ge 1 ]; then
  echo "python3-venv is installed"
else
  apt-get update
  apt-get install python3-venv -y
fi


MAIN_URL="https://raw.githubusercontent.com/Kate-0713/Pi-HoleAutoGrouping/refs/heads/main/src/main.py"
CONFIG_URL="https://raw.githubusercontent.com/Kate-0713/Pi-HoleAutoGrouping/refs/heads/main/src/config.json"
RUNSH_URL="https://raw.githubusercontent.com/Kate-0713/Pi-HoleAutoGrouping/refs/heads/main/src/run.sh"
DEST="/opt/Pi-HoleAutoGrouping"

echo "Creating script directory"
mkdir -p $DEST
cd $DEST
echo "Directory created"
echo "Creating Python virtual environment and installing dependencies"
python3 -m venv Pi-HoleAutoGrouping-venv
source Pi-HoleAutoGrouping-venv/bin/activate
pip install requests 
deactivate
echo "Virtual environment created"
echo "Downloading script files"
curl -L -O --output-dir "$DEST" "$MAIN_URL"
curl -L -O --output-dir "$DEST" "$CONFIG_URL"
curl -L -O --output-dir "$DEST" "$RUNSH_URL"
echo "Adding script to crontab"
(crontab -l; echo "0 20 * * * /opt/Pi-HoleAutoGrouping/run_pihole_auto_group.sh")|awk '!x[$0]++'|crontab -
echo "Crontab added"
echo "Script completed - ensure you update details in config.json at $DEST"

