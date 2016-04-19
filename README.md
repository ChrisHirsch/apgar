# Apgar Design Goals

We wanted a quick, simple and standardized way of doing health checks for the various services in our environment. Apgar walks a directory tree (by default `/etc/apgar/healthchecks`), runs the healthCheck scripts it finds there and aggregates the results into a directory (`/var/lib/apgar` by default). That directory is then served by a simple standalone web server so the results can be used as health checks by Amazon Load Balancers and Auto Scaling Groups.

## Details

Apgar consists of two parts, `apgar-server` which serves the health information, and `apgar-probe` which collects & aggregates the individual server health checks.

# Writing Checks

An apgar check should be executable, with a shebang line. It should either write "OK" to console and return 0, or write "NOT OK" to console and return anything other than 0.

The check should never print anything else unless --verbose is passed on the command line. If called with --quiet, it may print nothing at all, but still return zero or non-zero.

These checks should be fast and non-destructive - running them any number of times must be fine.

# FAQ

## Why Apgar?

Why not Apgar? Virginia Apgar invented the Apgar score as a method to quickly summarize the health of newborn children. Seemed appropriate for a quick health check system.

## Why not just piggyback on a system's existing web server?

You are of course free to use your own webserver instead of `apgar-server`, but we opted for a stand-alone server for the following reasons:

* I don't want to have to maintain configuration files for every webserver out there
* Not every system has a webserver installed
* Using a standalone server helps minimize Apgar's impact on existing services. Both apgar-server and apgar-probe are written in golang so that using Apgar doesn't pull in any dependencies that might conflict with those needed by the services you actually care about on a system.
* Using a standalone server keeps you from having to alter your existing webserver's configuration to use Apgar.

## Why Go?

I wanted Apgar to have as little impact on the host system as possible. Go gives us static binaries, and it was a good excuse to start learning Go.

## How can I package this so I don't have to build it on every server?

If you have bundler installed, `rake deb` will build a deb, and `rake rpm` will build a rpm.
