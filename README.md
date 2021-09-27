# !!! Work In Progress !!!

![main pipeline](https://github.com/iamsmartad/secretsyncer/actions/workflows/go.yml/badge.svg) [![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/iamsmartad/secretsyncer.svg)](https://github.com/iamsmartad/secretsyncer) [![GoReportCard](https://goreportcard.com/badge/github.com/iamsmartad/secretsyncer)](https://goreportcard.com/report/github.com/iamsmartad/secretsyncer) [![GitHub license](https://img.shields.io/github/license/iamsmartad/secretsyncer)](https://github.com/iamsmartad/secretsyncer/blob/master/LICENSE)

<!-- [![Docker](https://badgen.net/badge/icon/docker?icon=docker&label)](https://https://docker.com/) -->

# secretsyncer

> sync kubernetes secrets between namespaces and across clusters

## Getting started

Add helm repo:

```shell
helm repo add iampublic https://iamstudent.dev/chartrepo/iampublic
helm repo update
```

Create truth namespace:

```shell
kubectl create namespace truth
```

Install local-cluster-only secretsyncer:

```shell
helm -n truth upgrade --install secretsyncer iampublic/iamsyncer
```

### Initial Configuration

secretsyncer can run either

1. standalone on a cluster
   `manage syncronisation / cloning of secrets between namespaces`
1. as agent
   `manage syncronisation / cloning of secrets between namespaces AND across clusters`

## Developing

### Building

<!-- Run go build: -->

<!-- ```shell
go build -o myapp *.go
``` -->

<!-- ### Deploying / Publishing

In case there's some step you have to take that publishes this project to a
server, this is the right time to state it.

```shell
packagemanager deploy awesome-project -s server.com -u username -p password
```

And again you'd need to tell what the previous code actually does. -->

## Features

<!-- - What's the main functionality
- You can also do another thing
- If you get really randy, you can even do this -->

## Configuration

<!-- Here you should write what are all of the configurations a user can enter when
using the project. -->

## Contributing

"If you'd like to contribute, please fork the repository and use a feature
branch. Pull requests are warmly welcome."

## Links

<!-- - Project homepage: https://your.github.com/awesome-project/ -->

- Repository: https://github.com/iamsmartad/secretsyncer
- Issue tracker: https://github.com/iamsmartad/secretsyncer/issues
- Related projects:

## Licensing

"The code in this project is licensed under MIT license."
