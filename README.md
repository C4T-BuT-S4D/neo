[![Go](https://img.shields.io/badge/--00ADD8?logo=go&logoColor=ffffff)](https://golang.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/c4t-but-s4d/neo/v2)](https://goreportcard.com/report/github.com/c4t-but-s4d/neo/v2)
[![tests](https://github.com/c4t-but-s4d/neo/actions/workflows/tests.yml/badge.svg)](https://github.com/c4t-but-s4d/neo/actions/workflows/tests.yml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/c4t-but-s4d/neo)
[![Github all releases](https://img.shields.io/github/downloads/c4t-but-s4d/neo/total.svg)](https://GitHub.com/c4t-but-s4d/neo/releases/)


# neo

Client + server for exploit distribution during Attack & Defence CTF competitions.

## Why

Usually during large CTFs, where the regular laptop can't run the exploits on all participants, teams rent cloud
servers to run the exploits. However, uploading, managing and monitoring the exploits on a remote machine
can be a pain and wastes time. **Neo** helps in two primary ways:

1. Every player can start an instance of **Neo client** and attack a proportional part of the whole team pool
   automatically.

2. Exploit writers don't upload the newly-created exploits to the exploit server, neither do they manage the
   distribution manually, but rather submit them to the **Neo server** using the same client,
   and the server does all the work, distributing the exploit among the available clients.

## Usage

To start Neo, you'll need to start one server instance, DestructiveFarm and run one or more client instances.

### Farm

Neo uses the exploit farm to acquire the team list and submit the flags. The protocol is the following
(compatible with the [DestructiveFarm](https://github.com/DestructiveVoice/DestructiveFarm)):

- `GET /api/get_config` must return the configuration with keys `FLAG_FORMAT` and `TEAMS`. The first is the regex of the
  flag, and the second is the mapping `map(string -> string)`, where the key is the team name, and the value is the ip.

- `POST /api/post_flags` will receive an array of mappings with keys `flag`, `sploit` and `team`. `sploit` is the
  exploit name for statistics, and `team` is the team name.

Farm password will be passed in `Authorization` and `X-Token` headers, so the protocol is compatible with
**DestructiveFarm**.

### Server

Server coordinates the clients and distributes targets among them. It must have access to the farm, and all
clients must have access both to the server and the farm, so you might want to start it somewhere with the public IP
address.

To start the server:

1. Download the latest server release (`neo_server_...`) from
   the [Releases](https://github.com/pomo-mondreganto/neo/releases)
   page for your platform (64-bit amd linux and macOS are supported).

2. Edit the configuration in `configs/server/config.yml` file. Edit the `grpc_auth_key` (as it's the password required
   to
   connect to the server), `farm.url` and `farm.password`. You can also add some environment variables for all exploits
   in the `env` section

3. Start the server by simply running `./neo_server`

### Client

Client has a full-featured CLI and the single binary performs all operations required during the CTF. Client is
distributed as a docker image with a lot of useful python packages preinstalled, see the full list in `requirements.txt`
file (located at [client_env/requirements.txt](./client_env/requirements.txt) in the repository).

Download the latest client release (named `neo_client_env_{version}.zip`) from the
[Releases](https://github.com/pomo-mondreganto/neo/releases) page. The `start.sh` file starts the docker container with
the environment if one has not already been run and passes all arguments inside. For example, to get a shell inside the
container, one can run

```shell
./start.sh bash
```

The environment also contains neo binary:

```shell
./start.sh neo --help

Neo client

Usage:
  client [command]

Available Commands:
  add         Add an exploit
  broadcast   Run a command on all connected clients
  disable     Disable an exploit by id
  enable      Enable a disabled exploit by id
  help        Help about any command
  info        Print current state
  run         Start Neo client
  single      Run an exploit once on all teams immediately

Flags:
  -c, --config string   config file (default "client_config.yml")
  -h, --help            help for client
      --host string     server host (default "127.0.0.1")
  -v, --verbose         enable debug logging (default true)

Use "client [command] --help" for more information about a command.
```

As you can see, the binary has a nice help message. Each subcommand has a help message too, for example `add`:

```shell
./start.sh neo add --help

Add an exploit

Usage:
  client add [flags]

Flags:
  -d, --dir                 add exploit as a directory
  -e, --endless             mark script as endless
  -h, --help                help for add
      --id string           exploit name
  -i, --interval duration   run interval (default 15s)
  -t, --timeout duration    timeout for a single run (default 15s)

Global Flags:
  -c, --config string   config file (default "client_config.yml")
      --host string     server host (default "127.0.0.1")
  -v, --verbose         enable debug logging (default true)
```

Each exploit is identified by its file name, and if you try to add the same file again, Neo can replace the exploit with
its newer version.

Neo client only has access to the directory where the `start.sh` file is located, so to add a new exploit, you'll need
to put it somewhere next to `start.sh` (`exploits` directory might be a good place).

There are also `start_light.sh` and `start_sage.sh` scripts, which start the shallow alpine image
(useful for exploit management without running) and the largest image with Sage installed respectively.

## Development notice

Neo is very green and was only tested on a few CTFs by our team. Feel free to open issues and contribute in any way.
