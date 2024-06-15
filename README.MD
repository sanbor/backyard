# Backyard

Lightweight content management system. Run your very own social network in your backyard.

# Running on development machine

```
go run . --env=dev
```

For automatic reload when code changes:

```
go install github.com/mitranim/gow@latest
gow -e=go,mod,html -i=target run . -env=dev
```

# Running on production

You probably want to disable sign-up to avoid spammers.

```
go run . -env=pro -enable-signup -address=example.com -port 8080 -jwt-secret=random_1024_string
```

# Status of the project

Backyard is currently alpha quality, and in MPV (minimum viable product) phase.

For the MVP (minimum viable product) phase, the development of Backyard is exploratory, and aims to get
as many features as possible without spending too much time on automatic testing, security, performance,
or architecture.

If the software has no features, or nobody uses this software, then there is no point spending time in
great testing coverage, architecture, or security.

When Backyard reaches version 1.0, the MPV is complete and emphasis in quality and other areas will be given.

# Design decisions

## URL strcture

Homepage: webbackyard.com
User profile page: webbackyard.com/{username}
User posts: webbackyard.com/{username}/posts/{id}
User images: webbackyard.com/{username}/images/{image_id}
Sitemap: webbackyard.com/sitemap.xml
Robots: webbackyard.com/robots.txt
[Well-known URI](https://en.wikipedia.org/wiki/Well-known_URI): webbackyard.com/.well-known/
Any other files: webbackyard.com/static/{file}

# Business plan

To finance the development of Backyard, the following business model is in place.

- There is a hosting offering. The money goes towards covering the cost of hosting services and Backyard development.
- No lock-in. You can change hosting provider or use your own computer.
- That's it

Backyard aims to be a new content platform focused on content.
The goal for version 1.0 is to focus and functionalities and not federation.
