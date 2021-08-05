# cosign-verify-signature-fn

## Prerequisites
* _**cosign v1.0.0**_
* _**faas-cli v0.13.13**_
* _**kind v0.11.1**_
* _**docker v20.10.7**_
* _**arkade v0.7.25**_
* _**buildx v0.5.1-docker**_
* _**httpie v2.4.0**_

## Tutorial

### Run local kubernetes cluster

```shell
$ kind create cluster
```

### Install OpenFaaS

```shell
$ arkade install openfaas
```

You have to make OpenFaaS Gateway reachable from host in order to work with `faas-cli`.

```shell
$ kubectl port-forward -n openfaas svc/gateway 8080:8080 &
$ PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath="{.data.basic-auth-password}" | base64 --decode; echo)
$ echo -n $PASSWORD | faas-cli login --username admin --password-stdin
```

### Deploy Function

First, you have download the template from the store.

```shell
$ faas-cli template store pull golang-middleware
```

Let's generate our private/public key pairs.
> Don't use the default ones, please remove them and re-generate.

```shell
$ cd cosign-verify-signature-fn
$ cosign generate-key-pair
```

Then,let's build the function by using `--shrinkwrap` option, because there is no direct support in OpenFaaS functions 
for go 1.16 runtime, we have to edit our Dockerfile to make use of go version 1.16 as a base image.
> Don't forget to change prefix within the file cosign-verify-signature-fn.yml with your Docker user.

```shell
$ faas-cli build -f cosign-verify-signature-fn.yml --build-arg GO111MODULE=on --shrinkwrap
$ build/cosign-verify-signature-fn/
$ vim Dockerfile
- FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.15-alpine3.13 as build
+ FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.16-alpine3.13 as build

$ docker buildx build -t <your_docker_user>/cosign-verify-signature-fn:latest . --push
$ faas-cli deploy -f cosign-verify-signature-fn.yml
```

Please make sure that everythings work fine before move into the `Test` section.

```shell
$ kubectl get pods --namespace openfaas-fn
NAME                                          READY   STATUS    RESTARTS   AGE
cosign-verify-signature-fn-59d4bd7555-lhhcs   1/1     Running   0          16m
```

# Test

Let's sign our container image first, then verify it by using function.

```shell
$ cosign sign -key cosign.key <your_docker_user>/alpine:3.14.0
$ httpie http://127.0.0.1:8080/function/cosign-verify-signature-fn image=<your_docker_user>/alpine:3.14.0
Handling connection for 8080
HTTP/1.1 200 OK
Content-Length: 101
Content-Type: application/json
Date: Wed, 04 Aug 2021 17:08:36 GMT
X-Call-Id: 958b42a1-ac02-4ea4-be06-854d04aaa0d7
X-Duration-Seconds: 2.606183
X-Start-Time: 1628096914010520100

{
    "verification_message": "valid signatures found for an image: devopps/alpine:3.14.0",
    "verified": true
}
```

