# Backyard

Lightweight content management system. Run your very own social network in your backyard.

# Business plan

To finance the development of Backyard, the following business model is in place.

- There is a hosting offering. The money goes towards covering the cost of hosting services and Backyard development.
- No lock-in. You can change hosting provider or use your own computer.
- That's it

Backyard aims to be a new content platform focused on content.
The goal for version 1.0 is to focus and functionalities and not federation.

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
