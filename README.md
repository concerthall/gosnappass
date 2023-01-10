# Snappass, Gopher Edition ("gosnappass")

![gosnappass gopher](assets/gosnappass.png)

**GoSnappass** is a reimplementation of
[pinterest/snappass](https://github.com/pinterest/snappass) in Golang.

The project goal is to start with a drop-in compatible binary. You should be to
run this in a nearly identical environment to the original implementation and
have things working as expected.

We think this will change in the future, as we intend to simplify some of the
environment calls and provide some help output to the user. As we work towards
this goal, we intend to provide a configuration adaptor taking the original
snappass environment variable inputs and translating them to those used by
gosnappass.

For now, however, we'll keep with the original intent of being drop-in
compatible.

## Why did you build it?

For fun and practice, mostly! We've always liked the simplicty and effectiveness
of the original snappass. Porting this has allowed us to work with some new
libraries such as fernet-go, file embedding, the upcoming `slog` golang standard
library, the redis-go library, and Gorilla Mux.

## How do I check it out?

You can launch a quick redis server using `make redis`. This requires a
container engine like `podman` (default) or `docker`. Tear it down with `make
redis-teardown`. See the [Makefile](./Makefile) for information on how to
configure the target.

Then just `make run` to get the application running.

### With a Reverse Proxy

To test with a reverse proxy, try out the included [Caddyfile](./Caddyfile). It
binds **gosnappass** to `localhost:8080/sharepassword/`. Use
`URL_PREFIX=/sharepassword/ make run` after starting redis and caddy.