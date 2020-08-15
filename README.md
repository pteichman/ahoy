ahoy is a lightweight, low scale ActivityPub server for you and your
pals. It's inspired by Darius Kazemi's [Run your own social](https://runyourown.social).

Goals
=====

It supports a small number of users (<50).

It requires minimal attention to operate.

It has good controls for abuse, but at first these will assume a
benevolent local community. They'll focus on blocking remote abusers
and instances.

It federates with the greater ActivityPub network.

It supports local only communication.

Notes
=====

An ActivityPub server is a collection of actors, each with an inbox and
an outbox. These boxes contain [Activity Streams](https://www.w3.org/TR/activitystreams-core/),
which are "application/activity+json" documents containing activity events.

The actors on a server are discoverable (given a known username) via 
[webfinger](https://tools.ietf.org/html/rfc7033).
e.g. `curl 'https://mastodon.social/.well-known/webfinger?resource=acct:pteichman@mastodon.social'`

Server
======

The server is a single binary that assumes it will be run behind a
reverse proxy. It does not terminate its own TLS. Aside from that, it
is safe enough to run exposed to the internet.

CLI
===

The CLI is roughly based on [toot](https://pypi.org/project/toot/).

* ahoy login
* ahoy post
* ahoy follow
* ahoy timeline

Architecture
============

The inbox and outbox collections are treated as queues: we write to
them synchronously for durability, then enqueue the activities for
processing.

We do not allow an actor to read another actor's inbox, even items
that would otherwise be public.

Writing takes into account the actor's current blocklist. Maybe we'll
write blocked items (spam also) to a quarantine inbox.

ActivityPub sharedInbox delivery won't be necessary to support, since
this supports a small number of users. All writes will fan out.

Inbox write
-----------

Check signature; discard if bad.
Rate limit per inbox w/backpressure.

Inbox handle
------------

Check messages against actor's blocklist. Write to quarantine.

Outbox write
------------

Check signature; discard if bad.
Rate limit per outbox w/backpressure.

Outbox handle
-------------

Deliver to all recipients, both local and remote. Make a
per-remote-server queue?
