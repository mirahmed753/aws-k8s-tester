package eksconfig

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-k8s-tester/pkg/aws"
	"github.com/aws/aws-k8s-tester/pkg/logutil"
	"github.com/aws/aws-sdk-go/service/eks"
	"k8s.io/client-go/util/homedir"
)

// DefaultConfig is the default configuration.
//  - empty string creates a non-nil object for pointer-type field
//  - omitting an entire field returns nil value
//  - make sure to check both
//
// MAKE SURE TO SYNC THE DEFAULT VALUES in "eks" templates
//
var DefaultConfig = Config{
	// to be auto-generated
	ConfigPath:                "",
	KubectlCommandsOutputPath: "",
	SSHCommandsOutputPath:     "",
	KubeConfigPath:            "",
	Name:                      "",
	AWSCLIPath:                "",

	Region: "us-west-2",

	LogLevel: logutil.DefaultLogLevel,
	// default, stderr, stdout, or file name
	// log file named with cluster name will be added automatically
	LogOutputs: []string{"stderr"},

	// https://kubernetes.io/docs/tasks/tools/install-kubectl/
	// https://docs.aws.amazon.com/eks/latest/userguide/install-kubectl.html
	KubectlDownloadURL: "https://storage.googleapis.com/kubernetes-release/release/v1.14.10/bin/linux/amd64/kubectl",
	KubectlPath:        "/tmp/aws-k8s-tester/kubectl",

	OnFailureDelete:            true,
	OnFailureDeleteWaitSeconds: 60,

	Parameters: &Parameters{
		ClusterSigningName: "eks",
		Version:            "1.14",
	},

	// ref. https://docs.aws.amazon.com/eks/latest/userguide/create-managed-node-group.html
	// ref. https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-eks-nodegroup.html
	AddOnManagedNodeGroups: &AddOnManagedNodeGroups{
		Enable:      true,
		SigningName: "eks",

		RoleServicePrincipals: []string{
			"ec2.amazonaws.com",
			"eks.amazonaws.com",
		},
		RoleManagedPolicyARNs: []string{
			"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
			"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
			"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
		},

		// keep in-sync with the default value in https://pkg.go.dev/k8s.io/kubernetes/test/e2e/framework#GetSigner
		RemoteAccessPrivateKeyPath: filepath.Join(homedir.HomeDir(), ".ssh", "kube_aws_rsa"),
		RemoteAccessUserName:       "ec2-user", // assume Amazon Linux 2

		// to be auto-generated
		LogsDir: "",
	},

	AddOnNLBHelloWorld: &AddOnNLBHelloWorld{
		Enable:             true,
		DeploymentReplicas: 3,
	},
	AddOnALB2048: &AddOnALB2048{
		Enable:                 false,
		DeploymentReplicasALB:  3,
		DeploymentReplicas2048: 3,
	},
	AddOnJobPerl: &AddOnJobPerl{
		Enable:    false,
		Completes: 30,
		Parallels: 10,
	},
	AddOnJobEcho: &AddOnJobEcho{ // writes total 100 MB data to etcd
		Enable:    false,
		Completes: 1000,
		Parallels: 100,
		Size:      100 * 1024, // 100 KB
	},
	AddOnSecrets: &AddOnSecrets{
		Enable: false,

		Objects:     1000,
		Size:        10 * 1024, // 10 KB
		SecretQPS:   1,
		SecretBurst: 1,
		PodQPS:      100,
		PodBurst:    5,

		// writes total 153.6 MB for "Secret" objects,
		// plus "Pod" objects, writes total 330 MB to etcd
		//
		// with 3 nodes, takes about 1.5 hour for all
		// these "Pod"s to complete
		//
		// Objects:     15000,
		// Size:        10 * 1024, // 10 KB
		// SecretQPS:   150,
		// SecretBurst: 5,
		// PodQPS:      150,
		// PodBurst:    5,
	},
	AddOnIRSA: &AddOnIRSA{
		Enable:                false,
		RoleManagedPolicyARNs: []string{"arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"},
		DeploymentReplicas:    10,
	},

	// read-only
	Status: &Status{Up: false},
	StatusManagedNodeGroups: &StatusManagedNodeGroups{
		RoleCFNStackID:        "",
		NvidiaDriverInstalled: false,
		Nodes:                 make(map[string]StatusManagedNodeGroup),
	},
}

// NewDefault returns a copy of the default configuration.
func NewDefault() *Config {
	vv := DefaultConfig

	now := time.Now()
	vv.Name = fmt.Sprintf(
		"eks-%d%02d%02d%02d-%s",
		now.Year(),
		int(now.Month()),
		now.Day(),
		now.Hour(),
		randString(12),
	)

	// ref. https://docs.aws.amazon.com/eks/latest/userguide/create-managed-node-group.html
	// ref. https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-eks-nodegroup.html
	vv.AddOnManagedNodeGroups.MNGs = map[string]MNG{
		vv.Name + "-mng-cpu": MNG{
			Name:               vv.Name + "-mng-cpu",
			ReleaseVersion:     "", // to be auto-filled by EKS API
			AMIType:            "AL2_x86_64",
			ASGMinSize:         2,
			ASGMaxSize:         2,
			ASGDesiredCapacity: 2,
			InstanceTypes:      []string{DefaultNodeInstanceTypeCPU},
			VolumeSize:         DefaultNodeVolumeSize,
		},
	}

	return &vv
}

const (
	// DefaultNodeInstanceTypeCPU is the default EC2 instance type for CPU worker node.
	DefaultNodeInstanceTypeCPU = "c5.xlarge"
	// DefaultNodeInstanceTypeGPU is the default EC2 instance type for GPU worker node.
	DefaultNodeInstanceTypeGPU = "p3.8xlarge"

	// DefaultNodeVolumeSize is the default EC2 instance volume size for a worker node.
	DefaultNodeVolumeSize = 40

	// MNGMaxLimit is the maximum number of "Managed Node Group"s per a EKS cluster.
	MNGMaxLimit = 10
	// MNGNodesMaxLimit is the maximum number of nodes per a "Managed Node Group".
	MNGNodesMaxLimit = 100
)

func init() {
	// https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-welcome.html
	// pip3 install awscli --no-cache-dir --upgrade
	var err error
	DefaultConfig.AWSCLIPath, err = exec.LookPath("aws")
	if err != nil {
		panic(fmt.Errorf("aws CLI is not installed (%v)", err))
	}

	if runtime.GOOS == "darwin" {
		DefaultConfig.KubectlDownloadURL = strings.Replace(DefaultConfig.KubectlDownloadURL, "linux", "darwin", -1)
		DefaultConfig.AddOnManagedNodeGroups.RemoteAccessPrivateKeyPath = filepath.Join(os.TempDir(), randString(10)+".insecure.key")
	}
}

// ValidateAndSetDefaults returns an error for invalid configurations.
// And updates empty fields with default values.
// At the end, it writes populated YAML to aws-k8s-tester config path.
func (cfg *Config) ValidateAndSetDefaults() error {
	if _, ok := aws.RegionToAiport[cfg.Region]; !ok {
		return fmt.Errorf("region %q not found", cfg.Region)
	}
	if len(cfg.Name) == 0 {
		return errors.New("Name is empty")
	}
	if cfg.Name != strings.ToLower(cfg.Name) {
		return fmt.Errorf("Name %q must be in lower-case", cfg.Name)
	}
	if len(cfg.LogOutputs) == 0 {
		return errors.New("LogOutputs is not empty")
	}

	if cfg.ConfigPath == "" {
		rootDir, err := os.Getwd()
		if err != nil {
			rootDir = filepath.Join(os.TempDir(), cfg.Name)
			if err := os.MkdirAll(rootDir, 0700); err != nil {
				return err
			}
		}
		cfg.ConfigPath = filepath.Join(rootDir, cfg.Name+"-config.yaml")
		var p string
		p, err = filepath.Abs(cfg.ConfigPath)
		if err != nil {
			panic(err)
		}
		cfg.ConfigPath = p
	}
	if err := os.MkdirAll(filepath.Dir(cfg.ConfigPath), 0700); err != nil {
		return err
	}

	if cfg.KubectlCommandsOutputPath == "" {
		cfg.KubectlCommandsOutputPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + ".kubectl.sh"
	}
	if filepath.Ext(cfg.KubectlCommandsOutputPath) != ".sh" {
		cfg.KubectlCommandsOutputPath = cfg.KubectlCommandsOutputPath + ".sh"
	}
	if cfg.SSHCommandsOutputPath == "" {
		cfg.SSHCommandsOutputPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + ".ssh.sh"
	}
	if filepath.Ext(cfg.SSHCommandsOutputPath) != ".sh" {
		cfg.SSHCommandsOutputPath = cfg.SSHCommandsOutputPath + ".sh"
	}
	if cfg.KubeConfigPath == "" {
		cfg.KubeConfigPath = strings.ReplaceAll(cfg.ConfigPath, ".yaml", "") + ".kubeconfig.yaml"
	}
	if filepath.Ext(cfg.KubeConfigPath) != ".yaml" {
		cfg.KubeConfigPath = cfg.KubeConfigPath + ".yaml"
	}
	if cfg.AddOnManagedNodeGroups.LogsDir == "" {
		cfg.AddOnManagedNodeGroups.LogsDir = filepath.Join(filepath.Dir(cfg.ConfigPath), cfg.Name+"-mng-logs")
	}
	cfg.Sync()

	if !strings.Contains(cfg.KubectlDownloadURL, runtime.GOOS) {
		return fmt.Errorf("kubectl-download-url %q build OS mismatch, expected %q", cfg.KubectlDownloadURL, runtime.GOOS)
	}

	// validate role-related
	if (len(cfg.Parameters.ClusterRoleServicePrincipals) > 0 || len(cfg.Parameters.ClusterRoleManagedPolicyARNs) > 0) && cfg.Parameters.ClusterRoleARN != "" {
		return fmt.Errorf("non-empty Parameters.ClusterRoleServicePrincipals %+v or Parameters.ClusterRoleManagedPolicyARNs %+v, but got Parameters.ClusterRoleARN %q",
			cfg.Parameters.ClusterRoleServicePrincipals,
			cfg.Parameters.ClusterRoleManagedPolicyARNs,
			cfg.Parameters.ClusterRoleARN,
		)
	}
	if cfg.Status.ClusterRoleCFNStackID != "" {
		if cfg.Status.ClusterRoleName == "" {
			return fmt.Errorf("non-empty Status.ClusterRoleCFNStackID %q, but empty Status.ClusterRoleName",
				cfg.Status.ClusterRoleCFNStackID,
			)
		}
		if cfg.Status.ClusterRoleARN == "" {
			return fmt.Errorf("non-empty Status.ClusterRoleCFNStackID %q, but empty Status.ClusterRoleARN",
				cfg.Status.ClusterRoleCFNStackID,
			)
		}
	}
	if cfg.Status.ClusterRoleName != "" {
		if cfg.Status.ClusterRoleARN == "" {
			return fmt.Errorf("non-empty Status.ClusterRoleName %q, but empty Status.ClusterRoleARN",
				cfg.Status.ClusterRoleName,
			)
		}
	}
	if cfg.Status.ClusterRoleARN != "" {
		if cfg.Status.ClusterRoleName == "" {
			return fmt.Errorf("non-empty Status.ClusterRoleARN %q, but empty Status.ClusterRoleName",
				cfg.Status.ClusterRoleARN,
			)
		}
	}

	// validate VPC-related
	if cfg.Parameters.VPCCIDR != "" {
		if cfg.Parameters.PrivateSubnetCIDR1 == "" {
			return fmt.Errorf("non-empty Parameters.VPCCIDR %q, but got empty Parameters.PrivateSubnetCIDR1", cfg.Parameters.VPCCIDR)
		}
		if cfg.Parameters.PrivateSubnetCIDR2 == "" {
			return fmt.Errorf("non-empty Parameters.VPCCIDR %q, but got empty Parameters.PrivateSubnetCIDR2", cfg.Parameters.VPCCIDR)
		}
		if cfg.Parameters.PrivateSubnetCIDR3 == "" {
			return fmt.Errorf("non-empty Parameters.VPCCIDR %q, but got empty Parameters.PrivateSubnetCIDR3", cfg.Parameters.VPCCIDR)
		}
	}
	if cfg.Parameters.PrivateSubnetCIDR1 != "" {
		if cfg.Parameters.VPCCIDR == "" {
			return fmt.Errorf("non-empty Parameters.PrivateSubnetCIDR1 %q, but got empty Parameters.VPCCIDR", cfg.Parameters.PrivateSubnetCIDR1)
		}
	}
	if cfg.Parameters.PrivateSubnetCIDR2 != "" {
		if cfg.Parameters.VPCCIDR == "" {
			return fmt.Errorf("non-empty Parameters.PrivateSubnetCIDR2 %q, but got empty Parameters.VPCCIDR", cfg.Parameters.PrivateSubnetCIDR2)
		}
	}
	if cfg.Parameters.PrivateSubnetCIDR3 != "" {
		if cfg.Parameters.VPCCIDR == "" {
			return fmt.Errorf("non-empty Parameters.PrivateSubnetCIDR3 %q, but got empty Parameters.VPCCIDR", cfg.Parameters.PrivateSubnetCIDR3)
		}
	}
	if cfg.Status.VPCCFNStackID != "" {
		if cfg.Status.VPCID == "" {
			return fmt.Errorf("non-empty Status.VPCCFNStackID %q, but empty Status.VPCID",
				cfg.Status.VPCCFNStackID,
			)
		}
		if len(cfg.Status.PrivateSubnetIDs) == 0 {
			return fmt.Errorf("non-empty Status.ClusterRoleCFNStackID %q, but empty Status.PrivateSubnetIDs",
				cfg.Status.ClusterRoleCFNStackID,
			)
		}
		if cfg.Status.ControlPlaneSecurityGroupID == "" {
			return fmt.Errorf("non-empty Status.ClusterRoleCFNStackID %q, but empty Status.ControlPlaneSecurityGroupID",
				cfg.Status.ClusterRoleCFNStackID,
			)
		}
	}
	if cfg.Status.VPCID != "" {
		if cfg.Status.VPCCFNStackID == "" {
			return fmt.Errorf("non-empty Status.VPCID %q, but empty Status.VPCCFNStackID",
				cfg.Status.VPCID,
			)
		}
	}
	if len(cfg.Status.PrivateSubnetIDs) > 0 {
		if cfg.Status.ControlPlaneSecurityGroupID == "" {
			return fmt.Errorf("non-empty Status.PrivateSubnetIDs %+v, but empty Status.ControlPlaneSecurityGroupID",
				cfg.Status.PrivateSubnetIDs,
			)
		}
	}
	if cfg.Status.ControlPlaneSecurityGroupID != "" {
		if len(cfg.Status.PrivateSubnetIDs) == 0 {
			return fmt.Errorf("non-empty Status.ControlPlaneSecurityGroupID %q, but empty Status.PrivateSubnetIDs",
				cfg.Status.ControlPlaneSecurityGroupID,
			)
		}
	}

	// validate cluster-related
	if cfg.Parameters.Version == "" {
		return errors.New("empty Parameters.Version")
	}
	if len(cfg.Parameters.PrivateSubnetIDs) == 0 && cfg.Parameters.ControlPlaneSecurityGroupID != "" {
		return fmt.Errorf("empty Parameters.PrivateSubnetIDs, non-empty Parameters.ControlPlaneSecurityGroupID %q", cfg.Parameters.ControlPlaneSecurityGroupID)
	}
	if len(cfg.Parameters.PrivateSubnetIDs) > 0 && cfg.Parameters.ControlPlaneSecurityGroupID != "" {
		return fmt.Errorf("non-empty Parameters.PrivateSubnetIDs %+v, but empty Parameters.ControlPlaneSecurityGroupID", cfg.Parameters.PrivateSubnetIDs)
	}
	if cfg.Status.ClusterCFNStackID != "" {
		if cfg.Status.ClusterARN == "" {
			return fmt.Errorf("non-empty Status.ClusterCFNStackID %q, but empty Status.ClusterARN", cfg.Status.ClusterCFNStackID)
		}
		if cfg.Status.ClusterCA == "" {
			return fmt.Errorf("non-empty Status.ClusterCFNStackID %q, but empty Status.ClusterCA", cfg.Status.ClusterCFNStackID)
		}
		if cfg.Status.ClusterCADecoded == "" {
			return fmt.Errorf("non-empty Status.ClusterCFNStackID %q, but empty Status.ClusterCADecoded", cfg.Status.ClusterCFNStackID)
		}
	}
	// if created via EKS API, no need to error in the following case:
	// cfg.Status.ClusterARN != "" && cfg.Status.ClusterCA == "" || cfg.Status.ClusterCADecoded == ""

	// validate node group related
	if cfg.AddOnManagedNodeGroups.RoleName == "" {
		cfg.AddOnManagedNodeGroups.RoleName = cfg.Name + "-mng-role"
	}
	if cfg.AddOnManagedNodeGroups.SSHKeyPairName == "" {
		cfg.AddOnManagedNodeGroups.SSHKeyPairName = cfg.Name + "-ssh"
	}
	if cfg.AddOnManagedNodeGroups.Enable {
		if cfg.AddOnManagedNodeGroups.RemoteAccessPrivateKeyPath == "" {
			return errors.New("empty AddOnManagedNodeGroups.RemoteAccessPrivateKeyPath")
		}
		if cfg.AddOnManagedNodeGroups.RemoteAccessUserName == "" {
			return errors.New("empty AddOnManagedNodeGroups.RemoteAccessUserName")
		}
		n := len(cfg.AddOnManagedNodeGroups.MNGs)
		if n == 0 {
			return errors.New("AddOnManagedNodeGroups.Enable but empty AddOnManagedNodeGroups.MNGs")
		}
		if n > MNGNodesMaxLimit {
			return fmt.Errorf("AddOnManagedNodeGroups.MNGs %d exceeds maximum number of node groups per EKS which is %d", n, MNGNodesMaxLimit)
		}
		names := make(map[string]struct{})
		for k, v := range cfg.AddOnManagedNodeGroups.MNGs {
			if v.Name == "" {
				return fmt.Errorf("AddOnManagedNodeGroups.MNGs[%q].Name is empty", k)
			}
			if k != v.Name {
				return fmt.Errorf("AddOnManagedNodeGroups.MNGs[%q].Name has different Name field %q", k, v.Name)
			}
			_, ok := names[v.Name]
			if !ok {
				names[v.Name] = struct{}{}
			} else {
				return fmt.Errorf("AddOnManagedNodeGroups.MNGs[%q].Name %q is redundant", k, v.Name)
			}

			if v.VolumeSize == 0 {
				v.VolumeSize = DefaultNodeVolumeSize
			}

			switch v.AMIType {
			case eks.AMITypesAl2X8664:
				if len(v.InstanceTypes) == 0 {
					v.InstanceTypes = []string{DefaultNodeInstanceTypeCPU}
				}
			case eks.AMITypesAl2X8664Gpu:
				if len(v.InstanceTypes) == 0 {
					v.InstanceTypes = []string{DefaultNodeInstanceTypeGPU}
				}
			default:
				return fmt.Errorf("unknown AddOnManagedNodeGroups.MNGs[%q].AMIType %q", k, v.AMIType)
			}

			if cfg.AddOnNLBHelloWorld.Enable || cfg.AddOnALB2048.Enable {
				for _, itp := range v.InstanceTypes {
					// "m3.xlarge" or "c4.xlarge" will fail with "InvalidTarget: Targets {...} are not supported"
					// ref. https://github.com/aws/amazon-vpc-cni-k8s/pull/821
					// ref. https://github.com/kubernetes/kubernetes/issues/66044#issuecomment-408188524
					switch {
					case strings.HasPrefix(itp, "m3."),
						strings.HasPrefix(itp, "c4."):
						return fmt.Errorf("AddOnNLBHelloWorld.Enable[%v] || AddOnALB2048.Enable[%v], but older instance type InstanceTypes %q for %q", cfg.AddOnNLBHelloWorld.Enable, cfg.AddOnALB2048.Enable, itp, k)
					default:
					}
				}
			}

			if v.ASGMinSize > v.ASGMaxSize {
				return fmt.Errorf("AddOnManagedNodeGroups.MNGs[%q].ASGMinSize %d > ASGMaxSize %d", k, v.ASGMinSize, v.ASGMaxSize)
			}
			if v.ASGDesiredCapacity > v.ASGMaxSize {
				return fmt.Errorf("AddOnManagedNodeGroups.MNGs[%q].ASGDesiredCapacity %d > ASGMaxSize %d", k, v.ASGDesiredCapacity, v.ASGMaxSize)
			}
			if v.ASGMaxSize > MNGNodesMaxLimit {
				return fmt.Errorf("AddOnManagedNodeGroups.MNGs[%q].ASGMaxSize %d > MNGNodesMaxLimit %d", k, v.ASGMaxSize, MNGNodesMaxLimit)
			}
			if v.ASGDesiredCapacity > MNGNodesMaxLimit {
				return fmt.Errorf("AddOnManagedNodeGroups.MNGs[%q].ASGDesiredCapacity %d > MNGNodesMaxLimit %d", k, v.ASGDesiredCapacity, MNGNodesMaxLimit)
			}

			if cfg.AddOnNLBHelloWorld.Enable && cfg.AddOnNLBHelloWorld.DeploymentReplicas < int32(v.ASGDesiredCapacity) {
				cfg.AddOnNLBHelloWorld.DeploymentReplicas = int32(v.ASGDesiredCapacity)
			}
			if cfg.AddOnALB2048.Enable && cfg.AddOnALB2048.DeploymentReplicasALB < int32(v.ASGDesiredCapacity) {
				cfg.AddOnALB2048.DeploymentReplicasALB = int32(v.ASGDesiredCapacity)
			}
			if cfg.AddOnALB2048.Enable && cfg.AddOnALB2048.DeploymentReplicas2048 < int32(v.ASGDesiredCapacity) {
				cfg.AddOnALB2048.DeploymentReplicas2048 = int32(v.ASGDesiredCapacity)
			}

			cfg.AddOnManagedNodeGroups.MNGs[k] = v
		}

		if cfg.AddOnJobEcho.Size > 250000 {
			return fmt.Errorf("echo size limit is 0.25 MB, got %d", cfg.AddOnJobEcho.Size)
		}

		if cfg.AddOnNLBHelloWorld.Namespace == "" {
			cfg.AddOnNLBHelloWorld.Namespace = cfg.Name + "-nlb-hello-world"
		}
		if cfg.AddOnALB2048.Namespace == "" {
			cfg.AddOnALB2048.Namespace = cfg.Name + "-alb-2048"
		}
		if cfg.AddOnALB2048.PolicyName == "" {
			cfg.AddOnALB2048.PolicyName = cfg.Name + "-alb-ingress-controller-policy"
		}
		if cfg.AddOnJobPerl.Namespace == "" {
			cfg.AddOnJobPerl.Namespace = cfg.Name + "-job-perl"
		}
		if cfg.AddOnJobPerl.Namespace == cfg.Name {
			return fmt.Errorf("AddOnJobPerl.Namespace %q conflicts with %q", cfg.AddOnJobPerl.Namespace, cfg.Name)
		}
		if cfg.AddOnJobEcho.Namespace == "" {
			cfg.AddOnJobEcho.Namespace = cfg.Name + "-job-echo"
		}
		if cfg.AddOnJobEcho.Namespace == cfg.Name {
			return fmt.Errorf("AddOnJobEcho.Namespace %q conflicts with %q", cfg.AddOnJobEcho.Namespace, cfg.Name)
		}
		if cfg.AddOnSecrets.Namespace == "" {
			cfg.AddOnSecrets.Namespace = cfg.Name + "-secrets"
		}
		if cfg.AddOnSecrets.Namespace == cfg.Name {
			return fmt.Errorf("AddOnSecrets.Namespace %q conflicts with %q", cfg.AddOnSecrets.Namespace, cfg.Name)
		}
		if cfg.AddOnIRSA.Namespace == "" {
			cfg.AddOnIRSA.Namespace = cfg.Name + "-irsa"
		}
		if cfg.AddOnIRSA.Namespace == cfg.Name {
			return fmt.Errorf("AddOnIRSA.Namespace %q conflicts with %q", cfg.AddOnIRSA.Namespace, cfg.Name)
		}
		if cfg.AddOnIRSA.ServiceAccountName == "" {
			cfg.AddOnIRSA.ServiceAccountName = cfg.Name + "-irsa-service-account"
		}
		if cfg.AddOnIRSA.ConfigMapName == "" {
			cfg.AddOnIRSA.ConfigMapName = cfg.Name + "-irsa-configmap"
		}
		if cfg.AddOnIRSA.ConfigMapScriptFileName == "" {
			cfg.AddOnIRSA.ConfigMapScriptFileName = cfg.Name + "-irsa-configmap.sh"
		}
		if cfg.AddOnIRSA.S3BucketName == "" {
			cfg.AddOnIRSA.S3BucketName = cfg.Name + "-irsa-s3-bucket"
		}
		if cfg.AddOnIRSA.S3Key == "" {
			cfg.AddOnIRSA.S3Key = cfg.Name + "-irsa-s3-key"
		}
		if cfg.AddOnIRSA.DeploymentName == "" {
			cfg.AddOnIRSA.DeploymentName = cfg.Name + "-irsa-deployment"
		}
		if cfg.AddOnIRSA.DeploymentResultPath == "" {
			cfg.AddOnIRSA.DeploymentResultPath = filepath.Join(
				filepath.Dir(cfg.ConfigPath),
				cfg.Name+"-irsa-deployment-result.log",
			)
		}

		if cfg.AddOnSecrets.WritesResultPath == "" {
			cfg.AddOnSecrets.WritesResultPath = filepath.Join(
				filepath.Dir(cfg.ConfigPath),
				cfg.Name+"-secret-writes.csv",
			)
		}
		if filepath.Ext(cfg.AddOnSecrets.WritesResultPath) != ".csv" {
			return fmt.Errorf("expected .csv extension for WritesResultPath, got %q", cfg.AddOnSecrets.WritesResultPath)
		}
		if cfg.AddOnSecrets.ReadsResultPath == "" {
			cfg.AddOnSecrets.ReadsResultPath = filepath.Join(
				filepath.Dir(cfg.ConfigPath),
				cfg.Name+"-secret-reads.csv",
			)
		}
		if filepath.Ext(cfg.AddOnSecrets.ReadsResultPath) != ".csv" {
			return fmt.Errorf("expected .csv extension for ReadsResultPath, got %q", cfg.AddOnSecrets.ReadsResultPath)
		}

		if cfg.AddOnIRSA.RoleName == "" {
			cfg.AddOnIRSA.RoleName = cfg.Name + "-irsa-role"
		}

	} else {

		if cfg.AddOnNLBHelloWorld.Enable {
			return fmt.Errorf("AddOnManagedNodeGroups.Enable false, but got AddOnNLBHelloWorld.Enable %v", cfg.AddOnNLBHelloWorld.Enable)
		}
		if cfg.AddOnALB2048.Enable {
			return fmt.Errorf("AddOnManagedNodeGroups.Enable false, but got AddOnALB2048.Enable %v", cfg.AddOnALB2048.Enable)
		}
		if cfg.AddOnJobPerl.Enable {
			return fmt.Errorf("AddOnManagedNodeGroups.Enable false, but got AddOnJobPerl.Enable %v", cfg.AddOnJobPerl.Enable)
		}
		if cfg.AddOnJobEcho.Enable {
			return fmt.Errorf("AddOnManagedNodeGroups.Enable false, but got AddOnJobEcho.Enable %v", cfg.AddOnJobEcho.Enable)
		}
		if cfg.AddOnSecrets.Enable {
			return fmt.Errorf("AddOnManagedNodeGroups.Enable false, but got AddOnSecrets.Enable %v", cfg.AddOnSecrets.Enable)
		}
		if cfg.AddOnIRSA.Enable {
			return fmt.Errorf("AddOnManagedNodeGroups.Enable false, but got AddOnIRSA.Enable %v", cfg.AddOnIRSA.Enable)
		}
	}

	return cfg.Sync()
}