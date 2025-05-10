# mist

Reverse proxy and daemon service manager.

## Build and run

Run `service.sh` to build and start mist as a daemon.

By default, mist handles these endpoints:

- `/dashboard`: Dashboard for viewing and editing services.
- `/service`: API for service management.

Any proxied services will be found at `/<name>`.

