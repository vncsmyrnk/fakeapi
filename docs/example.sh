#!/usr/bin/env bash

# Spins up server docker container
docker run --rm -it -d \
  --name fakeapi-demo \
  -v ./config.json:/data/config.json \
  -p 8080:8080 \
  vncsmyrnk/fakeapi \
  config.json # Also uses a config file with some endpoints

echo "Waiting for some time for the server to properly spin up..."
sleep 1

# Simulates an expected request
echo "Sending first request: POST /items"
sleep 2
r=$(curl -s -X POST http://localhost:8080/items \
  -H 'header1:value1' \
  -d '{"id": 1}' \
  -w '%{http_code}')
test "$r" -eq "204"

# Simulates another expected request
echo "Sending second request: GET /items/1"
sleep 2
r=$(curl -s -X GET http://localhost:8080/items/1 \
  -H 'header2:value2')
test "1" -eq "1"

# Asserts request calls

echo -e "\nNow lets assert those requests!"
echo "Asserting the first request comparing method, URI, headers and payload body..."
sleep 2
fakeassert POST /items -H 'header1:value1' -b 'id=1' # You can match headers and body payload

echo -e "\nAsserting the first request comparing method and URI..."
sleep 2
fakeassert GET /items/1 # You can also omit headers and body and assert multiple requests at once

echo -e "\nAsserting there are no request left..."
sleep 2
fakeassert -c 0 # You can assert for no pending assertions left

sleep 1
echo -e "\nStopping docker container..."
docker stop fakeapi-demo
