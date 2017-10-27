# FedTidy

FedTidy is a tool used to clean up the resources created by Kubernetes Cluster Federation tests.

# Building from the source

Install glide by running

```shell
curl https://glide.sh/get | sh
```

See (Glide README)[https://github.com/Masterminds/glide] for other
methods.


Clone FedTidy's git repository and initialize the dependencies by
running the following command from the source root:

```shell
glide install
```

This command must be run the first time you want to build the source,
in order to initialize the dependencies. You must also run this every
time you want to update the dependencies.

After that, to build the source run make from the source root every
time:

```shell
make
```

This places the binary `fedtidy` in `_output/` directory.

Clean the build artifacts by running:

```shell
make clean
```

# Usage

`fedtidy` takes a single command line flag as input and it is
required.

`-c/--config` is the path to a JSON formatted config file that lists
the names of the projects and the GCP DNS zones that must be cleaned
up. 

Example:
```json
[
  {
    "project": "my-project",
    "dnsZone": "my-dns-zone-name" // Note that this is the name of the DNS Zone resource in GCP, not the domain name.
  },
  {
    "project": "my-project-release-1-7",
    "dnsZone": "release1-7-my-project-my-dns-zone-name"
  }
]
```
