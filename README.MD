# Backyard

Lightweight content management system.

# Development

```
go install github.com/mitranim/gow@latest
ENV=dev gow -e=go,mod,html -i=target run .
```

# Running on production

You probably want to disable sign-up to avoid spammers.

```
JWT_SECRET=random_password ENABLE_SIGNUP=false WHITELIST_HOST=example.com  go run .
```
