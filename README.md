# Pi-HoleAutoGrouping
Move Pi-Hole clients into certain groups based on a given prefix in the client comment


# TODO
- ~~Allow matching groups against a prefix rather than just the group ID~~
- Properly package script for easy installation of dependencies (notably the requests library)
- Create bash script to pull files and schedule script to run periodically
- Add logging to script
- Allow a person to configure program so that matched clients are added to different groups depending on their prefix


# CHANGE LOG

## 2026-04-24:
- Completed feature allowing the script to use a list of prefixes on the groups to add clients to allowing the app to be used without knowing group IDs
- [FIX] API session is now terminated when the script completes
- [FIX] JSON values for TARGET_CLIENT_PREFIXES and TARGET_GROUP_PREFIXES now actually set as an array instead of a string

## Before 2026-04-24:
- You'll have to read the commits I only just started this log