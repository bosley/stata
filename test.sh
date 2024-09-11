#!/bin/bash

# Build the stata binary
go build -o stata main.go

# Generate random data for test.html
random_data=$(openssl rand -base64 1536 | tr -d '\n' | cut -c1-2048)
echo "<html><body><div>$random_data</div></body></html>" > test.html

# Start the server in the background
./stata -port 8083 -secure -directory . &
server_pid=$!

# Wait for the server to start
sleep 2

# Curl the server and save the output
curl -k https://127.0.0.1:8083/test.html > output.html

# Compare the content
if grep -q "$random_data" output.html; then
    echo "Test passed: The correct data was retrieved from the server."
else
    echo "Test failed: The retrieved data does not match the expected content."
    echo "Expected: $random_data"
    echo "Got: $(cat output.html)"
fi

# Clean up
kill $server_pid
rm stata test.html output.html
