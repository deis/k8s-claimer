# k8s-claimer architecture

This document details how the k8s claimer server performs its lease and release operations in
accordance with the guarantees it gives.

## General

Each cluster is assigned a unique name, and the GKE API/backend stores those names (along
with other metadata). However, each cluster name, its UUID and its lease expiration time is stored
as a key/value pair (one k/v pair per cluster) in [annotations](http://kubernetes.io/docs/user-guide/annotations/)
on the k8s-server [service](http://kubernetes.io/docs/user-guide/services/). Specifically, a single
key/value pair looks like the following:

```
UUID => (cluster_name, lease_expiration_time)
```

## Leasing Algorithm

1. Download the annotations from the service and list the container clusters in GKE
2. If a cluster exists in GKE that's not in the annotations, set that to `found`
3. Otherwise, look for a cluster that has passed its lease expiration time.
If there is one, set that to `found`
4. If `found` is empty, return `409` (until [#9](https://github.com/deis/k8s-claimer/issues/9) is done)
5. Otherwise, add/overwrite an annotation with a new UUID and the new lease expiration time.
6. Save the annotation. If the save failed, go back to 1 for a (statically configurable)
number of retries
7. Return the UUID set in (5) to the client

## Releasing Algorithm

1. Download the annotations from the service
2. Look for the given token (in the URL Path) in the annotation keys
3. If none found, return `401` - the given token is not a valid lease
4. Otherwise, remove the annotation
5. Save the annotations. If the save failed, go back to 1 for a (statically configurable) number
of retries
6. Return `200 OK` to the client
