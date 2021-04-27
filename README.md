# Argo-CD Autopilot
[![Codefresh build status]( https://g.codefresh.io/api/badges/pipeline/codefresh-inc/argocd-autopilot%2Frelease?type=cf-1&key=eyJhbGciOiJIUzI1NiJ9.NTY3MmQ4ZGViNjcyNGI2ZTM1OWFkZjYy.AN2wExsAsq7FseTbVxxWls8muNx_bBUnQWQVS8IgDTI)]( https://g.codefresh.io/pipelines/edit/new/builds?id=60881f8199c9564ef31aac61&pipeline=release&projects=argocd-autopilot&projectId=608532835cbc54366d7483cb) [![codecov](https://codecov.io/gh/argoproj-labs/argocd-autopilot/branch/main/graph/badge.svg?token=IDyZNfRUfY)](https://codecov.io/gh/argoproj-labs/argocd-autopilot)

## Introduction

The Argo-CD Autopilot is a tool which offers an opinionated way of installing Argo-CD and managing GitOps repositories.

It can:
- create a new gitops repository.
- bootstrap a new argo cd installation.
- install and manage argo-cd projects and application with ease.

## Getting Started
```
argocd-autopilot repo create --owner <owner> --name <name> --token <git_token>
argocd-autopilot repo bootstrap --repo https://github.com/owner/name --token <git_token>
```
Head over to our [Getting Started](/docs/Getting-Started.md) guide for further details.

## How it works
The autopilot bootstrap command will deploy an Argo-CD manifest to a target k8s cluster, and will commit an Argo-CD Application manifest under a specific directory in your GitOps repository. This Application will manage the Argo-CD installation itself - so after running this command, you will have an Argo-CD deployment that manages itself through GitOps.

From that point on, the use can create Projects and Applications that belong to them. Autopilot will commit the required manifests to the repository. Once committed, Argo-CD will do its magic and apply the Applications to the cluster.

An application can be added to a project from a public git repo + path, or from a directory in the local filesystem.

## Architecture
![Argo-CD Autopilot Architecture](/docs/assets/architecture.png)

Autopilot communicates with the cluster directly **only** during the bootstrap phase, when it deploys Argo-CD. After that, most commands will only require access to the GitOps repository. When adding a Project or Application to a remote k8s cluster, autopilot will require access to the Argo-CD server.

## Features
* Opinionated way to build a multi-project multi-application system, using GitOps principles.
* Create a new GitOps repository, or use an existing one.
* Supports creating the entire directory structure under any path the user requires.
* When adding applications from a public repo, allow committing as either a kustomization that references the public repo, or as a "flat" manifest file containing all the required resources.
* Use a different cluster from the one Argo-CD is running on, as a default cluster for a Project, or a target cluster for a specific Application.

## Development Status
Argo-CD autopilot is currently under active development. Some of the basic commands are not yet implemented, but we hope to complete most of them in the coming weeks.
