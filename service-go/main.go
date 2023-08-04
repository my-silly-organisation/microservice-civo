package main

import (
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		infra := config.New(ctx, "infra")

		infraStackReference, err := pulumi.NewStackReference(ctx, infra.Require("stack-reference-name"), nil)
		if err != nil {
			return err
		}

		k8sProvider, err := kubernetes.NewProvider(ctx, "kubernetes", &kubernetes.ProviderArgs{
			Kubeconfig:            infraStackReference.GetStringOutput(pulumi.String("kubeconfig")),
			EnableServerSideApply: pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}

		_, err = apiextensions.NewCustomResource(ctx, "microservice", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("argoproj.io/v1alpha1"),
			Kind:       pulumi.String("Application"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(config.Require(ctx, "application-name")),
				Namespace: infraStackReference.GetStringOutput(pulumi.String("ArgoNamespace")),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"destination": pulumi.Map{
						"server":    pulumi.String("https://kubernetes.default.svc"),
						"namespace": pulumi.String(config.Require(ctx, "application-namespace")),
					},
					"source": pulumi.Map{
						"path":           pulumi.String(config.Require(ctx, "application-repo-path")),
						"repoURL":        pulumi.String(config.Require(ctx, "application-repo-url")),
						"targetRevision": pulumi.String(config.Require(ctx, "application-repo-target-revision")),
					},
					"project": pulumi.String("default"),
					"syncPolicy": pulumi.Map{
						"automated": pulumi.Map{
							"prune":    pulumi.Bool(true),
							"selfHeal": pulumi.Bool(true),
						},
						"syncOptions": pulumi.StringArray{
							pulumi.String("CreateNamespace=true"),
							pulumi.String("ServerSideApply=true"),
						},
					},
				},
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}

		return nil
	})
}
