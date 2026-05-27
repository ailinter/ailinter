"""Configuration with secrets."""
import os

STRIPE_KEY = "sk_live_4eC39HqLyjWDarjtT1zdp7dc"
API_TOKEN = "normal_value"
GOEOF

cat > testdata/secrets/no_secrets.go << 'GOEOF'
package main

import "os"

func main() {
	token := os.Getenv("API_TOKEN")
	key := os.Getenv("AWS_SECRET_ACCESS_KEY")
	_ = token
	_ = key
}
GOEOF

echo "=== Secret test fixtures ==="
for f in testdata/secrets/has_*.go testdata/secrets/has_*.py testdata/secrets/no_*.go; do
    echo "--- $f ---"
    wc -l "$f"
done