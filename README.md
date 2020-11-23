# sentinel-go-datasource-k8s-crd
k8s crd as sentinel go dynamic data source

1. create project
> kubebuilder init --domain sentinel.io

2. generate API: resource + Controller
> kubebuilder create api --group datasource --version v1 --kind FlowRules

3. make manifests generates Kubernetes object YAML, like CustomResourceDefinitions, WebhookConfigurations, and RBAC roles.
> make manifests

4. make generate generates code, like runtime.Object/DeepCopy implementations.
> make generate

5. Install the CRDs into the cluster:
> make install

6. To delete your CRDs from the cluster:
> make uninstall

7. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):
> make run