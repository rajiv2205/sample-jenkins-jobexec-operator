# sample-jenkins-jobexec-operator
Kubernetes Operator to execute Jenkins job from the controller whenever a CR of kind: JenkinsJobExec will be created in the cluster.

## Description
On the creation of a custom resource "JenkinsJobExec" in kubernetes cluster a job which is specifed in the CR will be executed
by the operator. 

**Note:** This will only work for new create requests.

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Run/Test On your Local Machine

Start the Minikube and execute the following commands to generate the CRD and apply it in your minikube cluster.

```sh
make manifests
make install
```

Get the username and corresponding apiToken from Jenkins which can execute jobs remotely. Generate the base64 of the username and apiToken and create a secret yaml and apply.   

```sh

[devops@the-wise-mortal samples (⎈|minikube:default)]$ echo -n '11884462973f6f7cfe34b88927612c4f58' | base64
MTE4ODQ0NjI5NzNmNmY3Y2ZlMzRiODg5Mjc2MTJjNGY1OA==

apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
data:
  username: cmFqaXY=  # base64 encoded value of 'username'
  token: MTE4ODQ0NjI5NzNmNmY3Y2ZlMzRiODg5Mjc2MTJjNGY1OA==  # base64 encoded apiToken of Jenkins


kubectl apply -f ./config/samples/my-secret.yaml
```

Create a configmap contains the jenkinsURL of your Jenkins. 

```sh

apiVersion: v1
kind: ConfigMap
metadata:
  name: my-configmap
  namespace: default
data:
  jenkinsURL: http://localhost:8080

kubectl apply -f ./config/samples/my-configmap.yaml
```

Run the operator on your local machine

```sh
make run
```

sample output:
```sh
/home/devops/git/operators-tutorials/examples/advanced/kubebuilder/projects/sample-jenkins-jobexec-operator/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/devops/git/operators-tutorials/examples/advanced/kubebuilder/projects/sample-jenkins-jobexec-operator/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
pkg/jenkinsutil/jenkins_utility.go
go vet ./...
go run ./cmd/main.go
2024-08-13T22:01:58+05:30	INFO	setup	starting manager
2024-08-13T22:01:58+05:30	INFO	starting server	{"name": "health probe", "addr": "[::]:8081"}
2024-08-13T22:01:58+05:30	INFO	Starting EventSource	{"controller": "jenkinsjobexec", "controllerGroup": "jenkins.operatortest.io", "controllerKind": "JenkinsJobExec", "source": "kind source: *v1.JenkinsJobExec"}
2024-08-13T22:01:58+05:30	INFO	Starting Controller	{"controller": "jenkinsjobexec", "controllerGroup": "jenkins.operatortest.io", "controllerKind": "JenkinsJobExec"}
2024-08-13T22:01:58+05:30	INFO	Starting workers	{"controller": "jenkinsjobexec", "controllerGroup": "jenkins.operatortest.io", "controllerKind": "JenkinsJobExec", "worker count": 1}
2024-08-13T22:01:58+05:30	INFO	Skipping already processed resource	{"name": "jenkinsjobexec-sample"}
2024-08-13T22:01:58+05:30	INFO	Skipping already processed resource	{"name": "jenkinsjobexec-sample-2"}
```

Now create the jenkinsjobexec.yaml

```sh

apiVersion: jenkins.operatortest.io/v1
kind: JenkinsJobExec
metadata:
  labels:
    app.kubernetes.io/name: sample-jenkins-jobexec-operator
    app.kubernetes.io/managed-by: kustomize
  name: jenkinsjobexec-with-param
spec:
  jobname: test_job_with_param        # jenkins job name
  parameters:
    test: successrun_from_operator    # parameter of jenkins job
  secretRef:
    name: my-secret     # secret that we created above
  configMapRef:
    name: my-configmap  # configmap that we created above


kubectl apply -f ./config/samples/jenkins_v1_jenkinsjobexec-3.yaml

```

you will see the following lines in the logs of the controller:

```sh
2024-08-13T22:05:13+05:30	INFO	Jenkins Job to be executed: test_job_with_param

2024-08-13T22:05:13+05:30	INFO	Going to execute Parameterized job
2024-08-13T22:05:13+05:30	INFO	Executed the job, status code: 201
2024-08-13T22:05:13+05:30	INFO	queue build Id is: 1
2024-08-13T22:05:13+05:30	INFO	Job is already in progress, cancel the job or wait
2024-08-13T22:05:23+05:30	INFO	Build Job URL: http://127.0.0.1:8080/job/test_job_with_param/65/
2024-08-13T22:05:23+05:30	INFO	Build still is in progress mode, waiting for its completion
2024-08-13T22:05:33+05:30	INFO	Build http://127.0.0.1:8080/job/test_job_with_param/65/ status: FAILURE

2024-08-13T22:05:33+05:30	INFO	Updated the build status url: http://127.0.0.1:8080/job/test_job_with_param/65/ status: FAILURE


```

The above job got failed and the same info you will find when you execute the following: 

```sh

kubectl get jenkinsjobexec

NAME                        JOBNAME               JOBSTATUS   BUILDURL
jenkinsjobexec-sample       test_job              SUCCESS     http://127.0.0.1:8080/job/test_job/56/
jenkinsjobexec-sample-2     test_job              CANCELLED   
jenkinsjobexec-with-param   test_job_with_param   FAILURE     http://127.0.0.1:8080/job/test_job_with_param/65/

```

