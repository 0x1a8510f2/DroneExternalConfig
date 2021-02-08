# DroneExternalConfig
[![Build Status](https://he.an-atom-in.space/api/badges/trslimey/DroneExternalConfig/status.svg)](https://he.an-atom-in.space/trslimey/DroneExternalConfig)

An extremely simple Drone CI (drone.io) configuration extension to allow fetching of build configs from various locations depending on the repository being built.

## Usage and configuration
DroneExternalConfig is primarily used as a [Docker container](https://hub.docker.com/r/trslimey/drone-external-config).

The container can be started like so:
```
docker run \
  -d \
  -p 8080:80 \
  --restart always \
  -v drone-external-config:/conf \
  --name drone-external-config \
  trslimey/drone-external-config:latest
```
This will fail initially as a config needs to be created. You can do this by placing a `config.ini` file in the `drone-external-config` volume or using and modifying the `example-config.ini` file which is already there.

The config file follows the `ini` format, and has two sections: `[server]` and `[config-map]`:

- `[server]` - Configuration for the HTTP server itself:
  - `listen-addr` - the address to listen for connections on
  - `listen-port` - the port to listen on
  - `tls-cert` - the location of the TLS certificate to use*
  - `tls-key` - the location of the TLS key to use*

\* Only if both of these options are enabled will the server use HTTPS. The port will remain as set (or default to 8080) so you may want to change the port to `443` or `8443`.

- `[config-map]` - A mapping of repositories to configuration files, represented as `<repo_name>`=`<config_url/uri>` where:
  - `repo_name` - the name of the repository including the author; for example, `TR-SLimey/DroneExternalConfig`
  - `config_url/uri` - the location of the config file (`http://`, `https://` or `file://`); for example, `https://example.com/configs/repo/project/drone.yml`

Example: `TR-SLimey/DroneExternalConfig=https://example.com/build-configs/TR-SLimey/DroneExternalConfig/drone.yml`
Note: The config files are fetched by DroneExternalConfig not Drone
