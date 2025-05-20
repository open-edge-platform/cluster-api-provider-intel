# Welcome to the Cluster API Provider for Intel

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/open-edge-platform/cluster-api-provider-intel/badge)](https://scorecard.dev/viewer/?uri=github.com/open-edge-platform/cluster-api-provider-intel)

## Overview

The Cluster API Provider for Intel enables using the [Cluster API](https://cluster-api.sigs.k8s.io/)
framework to create Kubernetes clusters on hosts managed by the [Edge Orchestrator](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/index.html).
Cluster API is a Kubernetes sub-project focused on providing declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.

The Cluster API Provider for Intel consists of two components: the Infrastructure Provider
controllers and the Southbound Handler.

## Get Started

The recommended way to try out the Cluster API Provider for Intel is by using the Edge Orchestrator.
Refer to the [Getting Started Guide](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/user_guide/get_started_guide/index.html) to get started with the Edge Orchestrator.

## Develop

If you are interested in contributing to the development of CAPI Provider for Intel, you will need
an environment where you can use it to create and delete clusters.  

The [cluster-tests](https://github.com/open-edge-platform/cluster-tests) repo provides a
lightweight environment for integration testing of the CAPI Provider for Intel as well as other
Edge Orchestrator components related to cluster management.  Clone that repo, change into the
cluster-tests directory, and run:

```
make test
```

This command creates a KinD cluster and deploys cert-manager, Cluster API operator, CAPI Provider for Intel,
Cluster Manager, and Cluster Connect Gateway.  It then creates and deletes a cluster inside a Kubernetes
pod.  Consult the cluster-tests [README](https://github.com/open-edge-platform/cluster-tests/blob/main/README.md)
for details on how to test your code in this environment.

## Contribute

We welcome contributions from the community! To contribute, please open a pull request to have your changes reviewed and merged into the main. To learn how to contribute to the project, see the [Contributor's Guide](https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/contributor_guide/index.html). We encourage you to add appropriate unit tests and e2e tests if your contribution introduces a new feature.

Additionally, ensure the following commands are successful:

```
make test
make lint
make license
```

## Community and Support

To learn more about the project, its community, and governance, visit the [Edge Orchestrator Community].
For support, start with [Troubleshooting] or [contact us].

## License

Cluster API Provider Intel is licensed under [Apache 2.0 License](LICENSES/Apache-2.0.txt)

[Edge Orchestrator Community]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/index.html
[Contact us]: https://github.com/open-edge-platform
[Troubleshooting]: https://docs.openedgeplatform.intel.com/edge-manage-docs/main/developer_guide/troubleshooting/index.html
