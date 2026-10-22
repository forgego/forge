---
sidebar_position: 4
description: Database settings.
image: /social-card.png
---

# Database Settings

- Driver
- Host, Port
- User, Password, Name
- SSLMode
- MaxOpenConns, MaxIdleConns
- ConnMaxLifetime, ConnMaxIdleTime

These settings configure Go's `database/sql` pool through `db.NewDB` or
`db.NewDBFromConfig`. The pool options are exercised against an in-process SQLite
connection. The PostgreSQL connection path uses the same pool setup, but a live
PostgreSQL pool test is not part of the in-process test suite.
