# Ecommerce setup

Follow [README.md](README.md#run-locally) for the current SQLite and PostgreSQL setup, migration, admin login, server, and UI test commands.

The default configuration uses SQLite. PostgreSQL requires an existing database and the `FORGE_DATABASE_*` connection settings. The example's admin login comes from `admin.username` and `admin.password` in `config/config.yaml` or their environment overrides. `forge createsuperuser` creates a user for Forge's separate identity system and is not needed for this example's admin UI.
