# Backyard

Lightweight content management system. Run your very own social network in your backyard.

# Status of the project

Backyard is currently alpha quality, and in MPV (minimum viable product) phase.

For the MVP (minimum viable product) phase, the development of Backyard is exploratory, and aims to get
as many features as possible without spending too much time on automatic testing, security, performance,
or architecture.

If the software has no features, or nobody uses this software, then there is no point spending time in
great testing coverage, architecture, or security.

When Backyard reaches version 1.0, the MPV is complete and emphasis in quality and other areas will be given.

# Running on development machine

```
ENV=dev go run .
```

For automatic reload when code changes:

```
go install github.com/mitranim/gow@latest
ENV=dev gow -e=go,mod,html -i=target run .
```

# Running on production

You probably want to disable sign-up to avoid spammers.

```
JWT_SECRET=random_password ENABLE_SIGNUP=false WHITELIST_HOST=example.com  go run .
```
