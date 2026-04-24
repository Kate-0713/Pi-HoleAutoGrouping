#!/bin/bash

MAIN_URL="https://raw.githubusercontent.com/Kate-0713/Pi-HoleAutoGrouping/refs/heads/main/src/main.py"
CONFIG_URL="https://raw.githubusercontent.com/Kate-0713/Pi-HoleAutoGrouping/refs/heads/main/src/config.json"
RUNSH_URL="https://raw.githubusercontent.com/Kate-0713/Pi-HoleAutoGrouping/refs/heads/main/src/run.sh"
DEST="/opt/Pi-HoleAutoGrouping"

echo "Creating script directory"
mkdir -p $DEST
cd $DEST
echo "Directory created"
echo "Creating Python virtual enviroment and installing dependancies"
python3 -m venv Pi-HoleAutoGrouping-venv
source Pi-HoleAutoGrouping-venv/bin/activate
pip install requests 
deactivate
echo "Virtual enviroment created"
echo "Downloading script files"
curl -L -O --output-dir "$DEST" "$MAIN_URL"
curl -L -O --output-dir "$DEST" "$CONFIG_URL"
curl -L -O --output-dir "$DEST" "$RUNSH_URL"
echo "Adding script to crontab"
(crontab -l; echo "0 20 * * * /opt/Pi-HoleAutoGrouping/run_pihole_auto_group.sh")|awk '!x[$0]++'|crontab -
echo "Crontab added"
echo "Script completed"

