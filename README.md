# k8s-claimer

[![Build Status](https://travis-ci.org/deis/k8s-claimer.svg?branch=master)](https://travis-ci.org/deis/k8s-claimer)
[![Go Report Card](http://goreportcard.com/badge/deis/k8s-claimer)](http://goreportcard.com/report/deis/k8s-claimer)
[![Docker Repository on Quay](https://quay.io/repository/deisci/k8s-claimer/status "Docker Repository on Quay")](https://quay.io/repository/deisci/k8s-claimer)

`k8s-claimer` is a leasing server for a pool of Kubernetes clusters. It will be used as part of our
[deis-workflow end-to-end test](https://github.com/deis/workflow-e2e) infrastructure.

Note that this repository is a work in progress. The code herein is under heavy development,
provides no guarantees and should not be expected to work in any capacity.

As such, it currently does not follow the
[Deis contributing standards](http://docs.deis.io/en/latest/contributing/standards/).

# Design

This server is responsible for holding and managing a set of [Google Container Engine](https://cloud.google.com/container-engine/)
(GKE) clusters. Each cluster can be in the `leased` or `free` state, and this server is responsible for
responding to requests to change a cluster's state, and then safely making the change.

A client who holds the lease for a cluster has a [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier)
indicating their ownership as well as the guarantee that nobody else will get the lease before
either their lease duration expires or someone releases the lease with their UUID. The client
specifies the lease duration when they acquire it.

For implementation details, see [the architecture document](doc/architecture.md)
