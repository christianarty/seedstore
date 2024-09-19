#! /usr/bin/env bash
# Catch errors, but you don't have to print it to me
set -euo &>/dev/null

# Read the server configuration file or env variable for the host path
SEEDSTORE_HOST=$(jq -r '.client.serverInfo.host' </config/config.json)

# Check to see if seedstore_host does not equal null
if [ "$SEEDSTORE_HOST" == "null" ]; then
  echo "Seedstore host is not set. Checking env vars..."
  # Check to see if envVar SEEDSTORE_CLIENT_SERVERINFO_HOST is set
  if [ -z "$SEEDSTORE_CLIENT_SERVERINFO_HOST" ]; then
    echo "Seedstore host is not set in env vars. Exiting..."
    exit 1
  else
    echo "Seedstore host is set in env vars. Continuing..."
    SEEDSTORE_HOST=$SEEDSTORE_CLIENT_SERVERINFO_HOST
  fi
fi
# Ensure the /root/.ssh directory exists
mkdir -p /root/.ssh
# Running ssh-keyscan to add the host to known_hosts so we avoid the "Host key verification failed" and make the output silent
ssh-keyscan "$SEEDSTORE_HOST" >>/root/.ssh/known_hosts 2>/dev/null

# Run the seedstore client
/usr/local/bin/seedstore "$@"
