#!/usr/bin/env sh
set -euo
#  This file is just to create the user with the same UID and GID as the env vars provided in the file.
USER_NAME="seedstore"
GROUP_NAME="seedstore"
PUID=${PUID:-1001}
PGID=${PGID:-1001}

# Create group if it doesn't exist
if ! getent group "$GROUP_NAME" >/dev/null 2>&1; then
  echo "Creating group '$GROUP_NAME' with GID $PGID..."
  addgroup -g "$PGID" "$GROUP_NAME"
else
  echo "Group '$GROUP_NAME' already exists."
fi

# Create user if it doesn't exist
if ! id -u "$USER_NAME" >/dev/null 2>&1; then
  echo "Creating user '$USER_NAME' with UID $PUID and GID $PGID..."
  adduser -u "$PUID" -G "$GROUP_NAME" -h /home/"$USER_NAME" -s /bin/sh -D "$USER_NAME"
else
  echo "User '$USER_NAME' already exists."
fi

# Ensure home directory exists
if [ ! -d /home/"$USER_NAME" ]; then
  echo "Creating home directory for user '$USER_NAME'..."
  mkdir -p /home/"$USER_NAME"
  chown "$USER_NAME":"$GROUP_NAME" /home/"$USER_NAME"
else
  echo "Home directory for user '$USER_NAME' already exists."
fi

echo "User and group setup complete."
