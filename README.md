# Pi-HoleAutoGrouping
Move Pi-Hole clients into certain groups based on a given prefix in the client comment


# TODO
- Allow matching groups against a prefix rather than just the group ID
  - Changes made to config file to support this but need to actual implementation
- Properly package script for easy installation of dependencies (notably the requests library)
- Create bash script to pull files and schedule script to run periodically
- Add logging to script