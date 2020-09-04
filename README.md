# RedCraft Server Management (rcsm)

This Golang project was made to manage Minecraft servers on Linux, with efficiency and scalability in mind.

## Requirements to run

rcsm requires tmux to be installed and in the PATH to work.

## How to install

- Download the latest release of rcsm from the [GitHub releases](https://github.com/redcraft-org/redcraft_server_management/releases) as `rcsm`
- Set it as runnable with `chmod +x rcsm`
- Download [.env.example](https://raw.githubusercontent.com/redcraft-org/redcraft_server_management/master/.env.example) as `.env` and edit the settings
- Edit `.env` to make sure it's set as you want
- Start rcsm with `./rcsm`

## Development

### Disclaimer

We're not golang experts and I'm pretty sure we're doing stuff wrong in this project because it will set the `GOPATH` to `src`, and we use some scripts to install deps.

Feel free to contribute to fix this if that's not how we're supposed to do it.

### Start the project

rcsm requires tmux, golang 1.13.8 or later to develop for, and that's pretty much it!

You can start the dev server by running `scripts/test.sh` and compile the project using `scripts/build.sh`.

## Features

rcsm was designed to automatically handle server updates from S3, alerts on Discord, and a way to communicate between services with little to no configuration using Redis.

Once set up, you shouldn't have anything to change.

### Config

rcsm uses env files that are .ini compliant.
By default, rcsm will look for a .env file, it's recommended to copy the `.env.example` to `.env` and then edit it to suit your needs.

#### Environment variables

You can also override some parameters using env variables.
For example, you can share one config file with all your servers and override `INSTANCE_NAME` and `MINECRAFT_SERVERS_TO_CREATE` by running rcsm like this:
`INSTANCE_NAME=survival MINECRAFT_SERVERS_TO_CREATE=survival rcsm`
Environment variables can also be used with systemd scripts.

#### Custom env file location

If you install rcsm in your path such as in `/usr/bin/rcsm`, you'll want to store your configuration somewhere else like in `/etc/rcsm`.

You can in fact do this using the `ENV_FILE` variable, like `ENV_FILE=/etc/rcsm /usr/bin/rcsm` (or using systemd files of course)

### Handling of servers

rcsm was primarily built to handle Minecraft servers, and it has multiple features and configuration for this.

Servers are started using tmux to be able to re-attach the console and manually take direct control of the server without rcsm.

tmux session prefixes can be changed with `MINECRAFT_TMUX_SESSION_PREFIX` (default is `rcsm_`)

#### S3 templates

rcsm was built primarily because other solutions didn't have any -good- solution for plugin updates.

At RedCraft.org, we use [server templates](https://github.com/redcraft-org/redcraft_server_config) that generate a usable .tar archives that can be restored on a server.

This is very useful if you use cron tasks with a tool to update a plugin repository such as [our plugin updater](https://github.com/redcraft-org/redcraft_plugins_updater).

If `S3_ENABLED` is set to true, then when a server starts (including automated restart), S3 will be used to check for templates in the specified S3 bucket.

Please notice that you can also use a 3rd party S3 compatible provider, such as Scaleway Object Storage (in fact that's what we use) or even [host it yourself](https://min.io/) by changing `S3_ENDPOINT`.

#### Server config

When rcsm starts, it will do a discovery of the folder specified by `MINECRAFT_SERVERS_DIRECTORY`. For every server, it will try to read a `rcsm_config.json` file that contains the following configuration:

- `start_args` to specify Java flags such as memory usage. By default, it's set to use 6 GB of memory and uses [these flags](https://aikar.co/2018/07/02/tuning-the-jvm-g1gc-garbage-collector-flags-for-minecraft/)
- `jar_name` for the server executable name, by default it's `server.jar`
- `stop_command` which is the command to gracefully stop the server, by default it's `stop` but for BungeeCord you'll have to set it to `end` for example.

#### Auto start/stop and "health checks"

rcsm was also built for high availability in mind.

Therefore, a few settings are dedicated to auto restart/stop/etc.

##### Auto start/stop

By default, rcsm will start every server specified in `MINECRAFT_SERVERS_DIRECTORY` when it starts, if not already started. This behavior can be disabled by setting `AUTO_START_ON_BOOT` to false.

That's right, rcsm isn't technically a wrapper because servers will continue to run even if rcsm is closed by default, but you can specify `AUTO_STOP_ON_CLOSE` to true to stop them on SIGINT (regular kill or ctrl + c)

##### Health checks

By default, rcsm checks if servers stops and restart them if they stop on their own or even if someone uses `/stop` from the console or in-game. If you want to stop a server, you'll have to send a redis pub command (cf down bellow for Redis help).

This behavior can be disabled by setting `AUTO_RESTART_CRASH_ENABLED` to false.

Also, if a server fails to reboot 3 times in under 2 minutes, the server will be marked as crashed and rcsm won't attempt to restart it automatically.
These values can be changed with `AUTO_RESTART_CRASH_MAX_TRIES` and `AUTO_RESTART_CRASH_TIMEOUT_SEC`

### Webhooks

rcsm has support for webhooks, more specifically for Discord webhooks.

You can enable this feature by setting `WEBHOOKS_ENABLED` to true.

Make sure to add a webhook integration on the desired Discord channel and set the endpoint on rcsm using `WEBHOOKS_ENDPOINT`.

Here's how the formatting looks:

<img width="393" alt="Discord webhooks" src="https://user-images.githubusercontent.com/2182934/92288579-55071f00-eedb-11ea-9ed2-0650a29593d7.png">

### Redis

Redis is a cache database, but a very interesting feature added years ago is the pub/sub feature that rcsm supports.

You can enable this feature by setting `REDIS_ENABLED` to true.

Basically, on any log, rcsm will publish a message on a channel (set with `REDIS_PUB_SUB_CHANNEL`) and other parts of your infrastructure can subscribe to this channel to get messages.

rcsm will also listen for commands on the channel.

#### Command format for rcsm

rcsm will listen on the pub/sub channel for JSON formats using the following fields:

- target (can be a server name or `*` for all servers)
- action (can be `start`/`stop`/`restart` or `command`)
- content (used only for `command` for now, it's the command to run in the console)

Please notice that:

- rcsm works well with UTF-8 characters, you can even send unicode characters in commands
- rcsm won't send you back the response for a command, but will acknowledge via an event

##### Examples

Restarting the `test1` server:

```json
{
    "target": "test1",
    "action": "restart"
}
```

Running `/op lululombard` on the `test2` server:

```json
{
    "target": "test2",
    "action": "command",
    "content": "op lululombard"
}
```

#### Format of the logs sent by rcsm

rcsm will publish logs with the following format:

- level (can be debug/info/warn/severe/fatal)
- instance (it's the instance name, by default `server` and can be changed with `INSTANCE_NAME`)
- service (it's the server name or any of the components like `redis`, `healthcheck`, `setup`, `updater` or `rcsm`)
- message (it's the log message)

Example:

```json
{
    "level": "SEVERE",
    "instance": "server",
    "service": "test1",
    "message": "Could not start: Server crashed on start, check server logs"
}
```

### Auto update of rcsm

By default, rcsm will check for updates and auto update itself.

You can disable auto updates when setting `AUTO_UPDATE_ENABLED` to false.

Here's how the process works:

- rcsm checks for update on the GitHub repo at boot and every hour
- if an update is available, it will download it and replace itself with the new binary
- if `EXIT_ON_AUTO_UPDATE` is set, rcsm will quit, otherwise the update will be applied at the next restart of rcsm

:warning: `EXIT_ON_AUTO_UPDATE` will NOT restart rcsm, it will only exit it. That's useful if you're using it as a system daemon and you automatically restart it on quit.
