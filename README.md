# Welcome to the Cluster API Provider for Intel

## Overview

The Cluster API Provider for Intel enables using the [Cluster API](https://cluster-api.sigs.k8s.io/)
framework to create Kubernetes clusters on hosts managed by the Intel® Open™ Edge Platform.
Cluster API is a Kubernetes sub-project focused on providing declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.

The Cluster API Provider for Intel consists of two components: the Infrastructure Provider
controllers and the Southbound Handler.

## Get Started

The recommended way to try out the Cluster API Provider for Intel is by using the Intel® Open™ Edge
Platform. Refer to the [Documentation](https://literate-adventure-7vjeyem.pages.github.io/edge_orchestrator/user_guide_main/content/user_guide/get_started_guide/gsg_content.html)
to get started with Intel® Open™ Edge Platform.

## Develop

If you are interested in contributing to the development of CAPI Provider for Intel, follow these
steps to get started:

```
make devenv
```

This command creates a KinD cluster, deploys cert-manager and Cluster API operator, and
builds and deploys Cluster API Provider for Intel from your local repository.

After making changes, rebuild and deploy the updated code.

```
make redeploy
```

## Contribute

We welcome contributions from the community! To contribute, please open a pull request to have your changes reviewed and merged into the main. We encourage you to add appropriate unit tests and e2e tests if your contribution introduces a new feature.

Additionally, ensure the following commands are successful:

```
make test
make lint
make license
```

## Community and Support

To learn more about the project, its community, and governance, visit the Edge Orchestrator Community. 
For support, start with Troubleshooting or contact us. 

## License

Cluster API Provider Intel is licensed under [Apache 2.0 License](LICENSES/Apache-2.0.txt)
