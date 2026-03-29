#!/bin/bash
curl -s -c /tmp/cookies.txt https://url.taverns.red/register > /tmp/page.html
TOKEN=$(grep -o 'gorilla.csrf.Token" content="[^"]*"' /tmp/page.html | sed 's/.*content="//;s/"//')
curl -s -w "\n%{http_code}" -b /tmp/cookies.txt -X POST "https://url.taverns.red/api/v1/auth/register" -H "Content-Type: application/x-www-form-urlencoded" -H "X-CSRF-Token: $TOKEN" -H "Origin: https://url.taverns.red" -H "Referer: https://url.taverns.red/register" -d "name=Test+User&email=test3@taverns.red&password=TestPassword123!"
