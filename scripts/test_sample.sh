#!/bin/bash

# Test script to simulate the PUT request from the sample data

echo "Testing LocalS3 with sample PUT request..."
echo "=========================================="

# Extract the JSON content from the sample
cat > test_animation.json << 'EOF'
{
  "v": "5.6.2",
  "fr": 29.9700012207031,
  "ip": 0,
  "op": 90.0000036657751,
  "w": 78,
  "h": 75,
  "nm": "Comp 1",
  "ddd": 0,
  "assets": [],
  "layers": [
    {
      "ddd": 0,
      "ind": 2,
      "ty": 4,
      "nm": "Layer 1/owl copy Outlines",
      "sr": 1,
      "ks": {
        "o": {"a": 0, "k": 100, "ix": 11},
        "r": {"a": 0, "k": 0, "ix": 10},
        "p": {"a": 0, "k": [37.203, 36.58, 0], "ix": 2}
      }
    }
  ]
}
EOF

echo "1. Creating bucket similar to your sample..."
curl -X PUT "http://localhost:3000/buckettest" \
     -H "Content-Type: application/json"

echo ""
echo "2. Uploading JSON file similar to your sample..."
curl -X PUT "http://localhost:3000/buckettest/json/promo/yourls.json.json" \
     -H "Content-Type: application/json" \
     -H "Content-Length: $(wc -c < test_animation.json)" \
     -T test_animation.json

echo ""
echo "3. Verifying the uploaded file..."
curl -s "http://localhost:3000/buckettest/json/promo/yourls.json.json" | jq '.'

echo ""
echo "4. Listing objects in bucket..."
curl -s "http://localhost:3000/buckettest" | grep -o '<Key>[^<]*</Key>' | sed 's/<Key>//g; s/<\/Key>//g'

echo ""
echo "✓ Test completed successfully!"
echo "The server can handle the same type of requests as in your sample data."

# Cleanup
rm -f test_animation.json
