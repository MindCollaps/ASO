#!/bin/bash

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        --go)
            GO_PATH="$2"
            shift # past argument
            shift # past value
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Use custom Go path if provided
if [[ -n "$GO_PATH" ]]; then
    GO_COMMAND="$GO_PATH"
else
    GO_COMMAND="go"
fi

# Check if Go is installed
if [[ -z $(command -v $GO_COMMAND) ]]; then
    echo "Error: Go is not installed."
    exit 1
fi

# Build server binary
echo "Building server binary..."
$GO_COMMAND build -o ASOServer

# Define paths
SERVER_BINARY="ASOServer"
SYSTEM_BINARY_DIR="/usr/local/sbin/"
SYSTEM_ETC_DIR="/etc/aso"
SYSTEM_LOG_DIR="/var/log/aso"
SYSTEMD_UNIT_FILE="/etc/systemd/system/aso-server.service"

# Check if server binary exists in the current directory
if [ ! -f "$SERVER_BINARY" ]; then
    echo "Error: Server binary '$SERVER_BINARY' not found in the current directory"
    exit 1
fi

if systemctl is-active --quiet aso-server; then
    read -pr "The ASO Server service is currently running. Do you want to stop it to update? (y/n): " choice
    case "$choice" in
      y|Y )
          systemctl stop aso-server
          ;;
      n|N )
          echo "Skipping service stop. The service will not be updated."
          exit 1
          ;;
      * )
          echo "Invalid choice. The service will not be updated."
          exit 1
          ;;
    esac
fi

# Create System directory's if it doesn't exist
mkdir -p "$SYSTEM_ETC_DIR"
mkdir -p "$SYSTEM_LOG_DIR"

# Copy the server binary to the system binary directory
cp -f "$SERVER_BINARY" "$SYSTEM_BINARY_DIR"

# Create a systemd unit file for the server service
cat <<EOF > "$SYSTEMD_UNIT_FILE"
[Unit]
Description=ASO Server
After=network.target

[Service]
Type=simple
ExecStart=$SYSTEM_BINARY_DIR$SERVER_BINARY --unix
WorkingDirectory=$SYSTEM_BINARY_DIR
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd to apply the changes
systemctl daemon-reload

# Start and enable the server service
systemctl start aso-server
systemctl enable aso-server

echo "ASO Server has been installed and configured as a service."
