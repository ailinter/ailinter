STRIPE_SECRET = "sk_test_4eC39HqLyjWDarjtT1zdp7dc"  # gitleaks:allow
GOEOF

cat > /Users/ivanb/Projects/ai-guard/testdata/secrets/github_token.js << 'JSEOF'
const GITHUB_TOKEN = "ghp_1A2b3C4d5E6f7G8h9I0jK1lM2nO3pQ4r5S" // gitleaks:allow
JSEOF

cat > /Users/ivanb/Projects/ai-guard/testdata/secrets/clean_file.go << 'GOEOF'
package main

import "os"

func getToken() string {
	return os.Getenv("API_TOKEN")
}

func getKey() string {
	return os.Getenv("AWS_SECRET_ACCESS_KEY")
}
GOEOF

echo "Secret fixtures created."
echo "=== All fixtures ===" && find /Users/ivanb/Projects/ai-guard/testdata -type f | sort