# tox4go [![build](https://github.com/alexbakker/tox4go/actions/workflows/build.yml/badge.svg)](https://github.com/alexbakker/tox4go/actions/workflows/build.yml)

__tox4go__ is a collection of utilities for [Tox](https://tox.chat) in Go. It
implements:
- A small portion of the Tox protocol. Enough for
  [nodes.tox.chat](https://nodes.tox.chat) to do its job.
- (De)serializers for the Tox state format (used by Tox clients to save the user
  profile).
- Client to fetch the nodes list from [nodes.tox.chat](https://nodes.tox.chat).

This project does not seek to become a full implementation of the Tox protocol.
