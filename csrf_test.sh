#!/bin/bash
# 1. Register
curl -s -c /tmp/c.txt https://url.taverns.red/register > /tmp/p.html
TOK=$(grep -o 'gorilla.csrf.Token" content="[^"]*"' /tmp/p.html | sed 's/.*content="//;s/"//')
EM="t${RANDOM}@taverns.red"
curl -s -c /tmp/c.txt -b /tmp/c.txt -X POST "https://url.taverns.red/api/v1/auth/register" -H "X-CSRF-Token: $TOK" -H "Origin: https://url.taverns.red" -H "Referer: https://url.taverns.red/register" -d "name=Test&email=$EM&password=Password1!" > /dev/null

# 2. Create link
curl -s -i -b /tmp/c.txt -X POST "https://url.taverns.red/api/v1/links" -H "X-CSRF-Token: $TOK" -H "Origin: https://url.taverns.red" -H "Referer: https://url.taverns.red/dashboard" -d "url=https://www.google.com"
