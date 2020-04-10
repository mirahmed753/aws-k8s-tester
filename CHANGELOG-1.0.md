

<hr>


## [v1.0.2](https://github.com/aws/aws-k8s-tester/releases/tag/v1.0.2) (2020-04-09)

See [code changes](https://github.com/aws/aws-k8s-tester/compare/v1.0.1...v1.0.2).

### `aws-k8s-tester`

- Remove [`aws-k8s-tester eks test`](https://github.com/aws/aws-k8s-tester/commit/237075a8130f1ad29e3c3b655ca4d52fa5632426).
- Improve [`aws-k8s-tester eks check`](https://github.com/aws/aws-k8s-tester/commit/237075a8130f1ad29e3c3b655ca4d52fa5632426).
  - [`aws-k8s-tester eks check cluster` is now just `aws-k8s-tester eks check`](https://github.com/aws/aws-k8s-tester/commit/623350901946156b97ef985aa4b2344a9e654835).

### `eks`

- Use [`pkg/k8s-client.NewEKS` for `*kubernetes.Clientset`; use `pkg/k8s-client.EKS` interface](https://github.com/aws/aws-k8s-tester/commit/85db2dd0c9f64af5d37be1b304d63ff2d42cdc79).
- Move [`healthz` checks to `pkg/k8s-client.EKS` interface](https://github.com/aws/aws-k8s-tester/commit/0d7981d66303ba8384ec57b338feb084bca64bdf).
- Fix [node group instance check when `DHCP` options are set](https://github.com/aws/aws-k8s-tester/commit/2cc88ee4ab04159ec04306400f7f5d8c44b81f8d).
- Log [node `Labels` when polling node status](https://github.com/aws/aws-k8s-tester/commit/26f67f5285ffdb748914233272857bb7ff0f048e).

### `pkg/k8s-client`

- Add [`k8sclient.NewEKS` and `k8sclient.EKSConfig` for `*kubernetes.Clientset`; use `pkg/k8s-client.EKS` interface](https://github.com/aws/aws-k8s-tester/commit/e673d3388ee44889e6572dcdcee530ea06984a86).
- Move [`healthz` checks to `k8sclient.EKS` interface](https://github.com/aws/aws-k8s-tester/commit/3dac533adcf2fb0aa51f19d4f56bbc9dd2b59eb5).

### Dependency

- Upgrade [`github.com/aws/aws-sdk-go`](https://github.com/aws/aws-sdk-go/releases) from [`v1.30.7`](https://github.com/aws/aws-sdk-go/releases/tag/v1.30.7) to [`v1.30.X`](https://github.com/aws/aws-sdk-go/releases/tag/v1.30.X).

### Go

- Compile with [*Go 1.14.2*](https://golang.org/doc/devel/release.html#go1.14).


<hr>


## [v1.0.1](https://github.com/aws/aws-k8s-tester/releases/tag/v1.0.1) (2020-04-08)

See [code changes](https://github.com/aws/aws-k8s-tester/compare/v1.0.0...v1.0.1).

### `ec2config`

- Add [`DHCPOptionsDomainName`](https://github.com/aws/aws-k8s-tester/commit/1f90891e0aeaa9fcffb25acda12f5f4e4a78f706).
  - `AWS_K8S_TESTER_EC2_DHCP_OPTIONS_DOMAIN_NAME`
- Add [`DHCPOptionsDomainNameServers`](https://github.com/aws/aws-k8s-tester/commit/1f90891e0aeaa9fcffb25acda12f5f4e4a78f706).
  - `AWS_K8S_TESTER_EC2_DHCP_OPTIONS_DOMAIN_NAME_SERVERS`

### `eksconfig`

- Add [`Parameters.DHCPOptionsDomainName`](https://github.com/aws/aws-k8s-tester/commit/84dd682a673eaa01fbf6bbbf3e664ad82c1dbbf4
).
  - `AWS_K8S_TESTER_EKS_PARAMETERS_DHCP_OPTIONS_DOMAIN_NAME`
- Add [`Parameters.DHCPOptionsDomainNameServers`](https://github.com/aws/aws-k8s-tester/commit/84dd682a673eaa01fbf6bbbf3e664ad82c1dbbf4).
  - `AWS_K8S_TESTER_EKS_PARAMETERS_DHCP_OPTIONS_DOMAIN_NAME_SERVERS`
- Change [`eksconfig.Config.AddOnNodeGroups.ASGs` from `map[string]ec2config.ASG` to `map[string]eksconfig.ASG`](https://github.com/aws/aws-k8s-tester/commit/e302d15f428e014931e1f43a3a0e8cafec136e77).
  - To support `--kubelet-extra-args`.
  - Added `eksconfig.ASG` with `KubeletExtraArgs` field.
  - ref. https://github.com/awslabs/amazon-eks-ami/blob/master/files/bootstrap.sh

### `eks`

- Improve [`AddOnNodeGroups` delete operation](https://github.com/aws/aws-k8s-tester/commit/90b0b50819da58a48cfebef8f6172238426dd8b5).
- Improve [`AddOnManagedNodeGroups` delete operation](https://github.com/aws/aws-k8s-tester/commit/5a21706eaf6ff00b65ef651385b99b6f23676633).

### Dependency

- Upgrade [`github.com/aws/aws-sdk-go`](https://github.com/aws/aws-sdk-go/releases) from [`v1.30.4`](https://github.com/aws/aws-sdk-go/releases/tag/v1.30.4) to [`v1.30.7`](https://github.com/aws/aws-sdk-go/releases/tag/v1.30.7).

### Go

- Compile with [*Go 1.14.2*](https://golang.org/doc/devel/release.html#go1.14).


<hr>


## [v1.0.0](https://github.com/aws/aws-k8s-tester/releases/tag/v1.0.0) (2020-04-05)

See [code changes](https://github.com/aws/aws-k8s-tester/compare/v0.9.8...v1.0.0).

### `ec2config`

- Update [`README.md`](https://github.com/aws/aws-k8s-tester/commit/eb0d6bca8bd01da418901acfa2c7b1fd5080d9bd).
- Clean up [`RemoteAccessPrivateKeyPath` defaults](https://github.com/aws/aws-k8s-tester/commit/eb0d6bca8bd01da418901acfa2c7b1fd5080d9bd).
- Fix [`ec2config.NewDefault`, remove `DefaultConfig`](https://github.com/aws/aws-k8s-tester/commit/13eabf5034488eefa0a028449f3f23233ef74661).
  - `ec2config.NewDefault` was copying the add-on fields in reference.
- Check [`ImageID` and `ImageIDSSMParameter`](https://github.com/aws/aws-k8s-tester/commit/13151dfd539a31175a9014e2115148605c9bc001).

### `ec2`

- Improve [ASG create and delete performance](https://github.com/aws/aws-k8s-tester/commit/4a97173663a4f383b2810051fd630b93c49f6351).

### `eksconfig`

- Update [`README.md`](https://github.com/aws/aws-k8s-tester/commit/eb0d6bca8bd01da418901acfa2c7b1fd5080d9bd).
- Clean up [`RemoteAccessPrivateKeyPath` defaults](https://github.com/aws/aws-k8s-tester/commit/eb0d6bca8bd01da418901acfa2c7b1fd5080d9bd).
- Fix [`eksconfig.NewDefault`, remove `DefaultConfig`](https://github.com/aws/aws-k8s-tester/commit/13eabf5034488eefa0a028449f3f23233ef74661).
  - `eksconfig.NewDefault` was copying the add-on fields in reference.
- Check [`ImageID` and `ImageIDSSMParameter`](https://github.com/aws/aws-k8s-tester/commit/13151dfd539a31175a9014e2115148605c9bc001).

### `eks`

- Add [missing `AddOnCSRs` delete operation](https://github.com/aws/aws-k8s-tester/commit/e91e12f256a60d74a9f08dead964608f74beee5a).
- Add [missing `AddOnConfigMaps` delete operation](https://github.com/aws/aws-k8s-tester/commit/e91e12f256a60d74a9f08dead964608f74beee5a).
- Improve [inflight creation requests cancel](https://github.com/aws/aws-k8s-tester/commit/da59e6bca6c117b3737bbb36598a3830b63ec7cf).
- Upgrade [`eks/alb` `kubernetes-sigs/aws-alb-ingress-controller` version from `v1.1.5` to `v1.1.6`](https://github.com/aws/aws-k8s-tester/commit/8df3fc79196113d19ad84077aab3bdc1c3805249).
- Delete [encryption CMK at the end](https://github.com/aws/aws-k8s-tester/commit/2436dafee14582014a97a08637272211c80f1d79).
  - Otherwise, `kube-apiserver` `/healthz` check fails.

### `pkg/k8s-client`

- Increase [`DefaultNamespaceDeletionInterval` from 5-second to 15-second](https://github.com/aws/aws-k8s-tester/commit/1a41c61813e1e0872b44738773ccdda4e765be1c).
- Improve [`DeleteNamespaceAndWait` retries on `i/o timeout`](https://github.com/aws/aws-k8s-tester/commit/1a41c61813e1e0872b44738773ccdda4e765be1c).

### Dependency

- Upgrade [`github.com/go-ini/ini`](https://github.com/go-ini/ini/releases) from [`v1.46.0`](https://github.com/go-ini/ini/releases/tag/v1.46.0) to [`v1.55.0`](https://github.com/go-ini/ini/releases/tag/v1.55.0).
- Upgrade [`sigs.k8s.io/yaml`](https://github.com/kubernetes-sigs/yaml/releases) from [`v1.1.0`](https://github.com/kubernetes-sigs/yaml/releases/tag/v1.1.0) to [`v1.2.0`](https://github.com/kubernetes-sigs/yaml/releases/tag/v1.2.0).

### Go

- Compile with [*Go 1.14.1*](https://golang.org/doc/devel/release.html#go1.14).


<hr>
