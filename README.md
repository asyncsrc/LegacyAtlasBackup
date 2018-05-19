# LegacyAtlasBackup
Backup your legacy Hashicorp Atlas environment prior to May 31st

This script takes two parameters:

-c:  cookie value pulled from your browser *after* successfully authenticating against atlas.hashicorp.com
In Chrome, you can look at the dev tools, and look at the header for potentially any response from Atlas and pull the *_atlas_session_data* cookie value, stopping at the first ';'.

-p: path to parent folder where session states will be stored

-o: organization name

### How to run
```go run main.go -c "ZktQRWF2aUIyM1MrZ0lyUSs3bTIvNS90WUUvQU9rVWFn[...];" -p "/tmp" -o "OrgName" ```

This will iterate through all environments for your organization and grab the latest state and save to path specified by '-p'.  
