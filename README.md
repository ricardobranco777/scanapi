![Build Status](https://github.com/ricardobranco777/scanapi/actions/workflows/ci.yml/badge.svg)

scanapi is an API service scanner that uses Goroutines for faster service discovery.

Services supported:
- Bugzilla
- Docker Distribution
- Gitea
- Gitlab
- Jira
- Pagure
- Redmine

## Usage

```
Usage: ./scanapi [OPTIONS] URL
  -H, --header strings   HTTP header (may be specified multiple times
  -t, --timeout int      timeout (default 60)
      --version          print version and exit
```

## Examples

```
$ scanapi https://bugzilla.opensuse.org
Bugzilla: {"version":"5.0.6"}

$ scanapi https://registry.opensuse.org
Docker Registry

$ scanapi https://src.opensuse.org
Gitea: {"version":"1.22.4"}
Docker Registry

$ scanapi -H "PRIVATE-TOKEN: $GITLAB_TOKEN" https://gitlab.suse.de
GitLabv4: {"version":"77.7.0","revision":"88888888","kas":{"enabled":false,"externalUrl":null,"externalK8sProxyUrl":null,"version":null},"enterprise":false}

$ scanapi https://issues.redhat.com
Jira: {"baseUrl":"https://issues.redhat.com","version":"9.12.15","versionNumbers":[9,12,15],"deploymentType":"Server","buildNumber":9120015,"buildDate":"2024-11-06T00:00:00.000+0000","databaseBuildNumber":9120015,"scmInfo":"4d9c4bd7a744d02f6fd1c40b1e7bc5cff95adbdc","serverTitle":"Red Hat Issue Tracker"}

$ scanapi https://code.opensuse.org
Pagure: {"version":"0.31"}

$ scanapi https://progress.opensuse.org
RedMine
```
