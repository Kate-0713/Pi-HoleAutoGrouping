#!/bin/bash
echo "Starting Pi-Hole Auto Group script"
source Pi-HoleAutoGrouping-venv/bin/activate
python3 /opt/Pi-HoleAutoGrouping/main.py
deactivate
echo "Pi-Hole Auto Group script complete"
