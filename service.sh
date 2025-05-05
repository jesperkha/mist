#!/bin/bash

# Setup script for minimal daemon config file

set -e

defaultName=$(basename "$PWD")
read -p "Service name [$defaultName]: " name
name="${name:-$defaultName}"

read -p "Description: " description

serviceFile=$name.service
defaultWorkdir="$PWD"

read -p "Service dir [$defaultWorkdir]: " workdir
workdir="${workdir:-$defaultWorkdir}"

cat > "$serviceFile" <<EOF
[Unit]
Description=$description
After=network.target

[Service]
ExecStart=$workdir/build/$name
WorkingDirectory=$workdir
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

echo "Created $serviceFile"

ask_yes_no() {
    while true; do
        read -p "$1 (y/n): " yn
        case $yn in
            [Yy]* ) return 0;;
            [Nn]* ) return 1;;
            * ) echo "Please answer yes or no.";;
        esac
    done
}

if ask_yes_no "Create service now?"; then
    mkdir -p build

    echo "Building executable"
    go build -o build/$name cmd/main.go

    echo "Copying service file"
    sudo cp $serviceFile /etc/systemd/system/
    sudo chmod 644 /etc/systemd/system/$serviceFile
else
    exit 0
fi

echo "Done"

if ask_yes_no "Start service?"; then
    sudo systemctl start $name
    echo "Service running 🚀"
fi
