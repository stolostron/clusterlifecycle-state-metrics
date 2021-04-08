[comment]: # ( Copyright Contributors to the Open Cluster Management project )

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Contributing guidelines](#contributing-guidelines)
    - [Contributions](#contributions)
    - [Certificate of Origin](#certificate-of-origin)
    - [Contributing A Patch](#contributing-a-patch)
    - [Issue and Pull Request Management](#issue-and-pull-request-management)
    - [Pre-check before submitting a PR](#pre-check-before-submitting-a-pr)
    - [Build images](#build-images)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Contributing guidelines

## Terms

All contributions to the repository must be submitted under the terms of the [Apache Public License 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## Certificate of Origin

By contributing to this project, you agree to the Developer Certificate of Origin (DCO). This document was created by the Linux Kernel community and is a simple statement that you, as a contributor, have the legal right to make the contribution. See the [DCO](DCO) file for details.

## Contributing a patch

1. Submit an issue describing your proposed change to the repository in question. The repository owners will respond to your issue promptly.
2. Fork the desired repository, then develop and test your code changes.
3. Submit a pull request.

## Issue and pull request management

Anyone can comment on issues and submit reviews for pull requests. In order to be assigned an issue or pull request, you can leave a `/assign <your Github ID>` comment on the issue or pull request (PR).

## Pre-check before submitting a PR 
<!-- Customize this template for your repository -->

Before submitting a PR, please perform the following steps:

- List of steps to perform before submitting a PR.

After your PR is ready to commit, please run following commands to check your code.

```shell
make check
make test
make functional-test-full
```

## Build images

Make sure your code build passed.

```shell
export BUILD_LOCALLY=1
make
```

Now, you can follow the [getting started guide](./README.md#getting-started) to work with this repository.

## Add a new metric

1. Update the [pkg/options/collector.go](pkg/options/collector.go) `DefaultCollectors` varaible with your new metric name.
2. Update the [pkg/collectors/builder.go](pkg/collectors/builder.go) `availableCollectors` variable with your new metric name and create similar methods than `buildManagedClusterInfoCollector`
3. Clone the [pkg/collectors/managedclusterinfo.go](pkg/collectors/managedclusterinfo.go) to implement your metric and adapt it for your new metric.
4. Create unit tests.
5. Create functianal tests.

## Create a new project metrics

1. Clone this project
2. Update the [pkg/options/collector.go](pkg/options/collector.go) `DefaultCollectors` varaible with your new metric name.
3. Update the [pkg/collectors/builder.go](pkg/collectors/builder.go) `availableCollectors` variable with your new metric name and rename methods `buildManagedClusterInfoCollector*`. If you have multiple metrics then creates similar methods and update the `availableCollectors`.
4. Modify/rename the [pkg/collectors/managedclusterinfo.go](pkg/collectors/managedclusterinfo.go) to implement your metric and adapt it for your new metric. The method `getManagedClusterInfoMetricFamilies` and implement your own business logic.
5. Create unit tests.
6. Create functianal tests.