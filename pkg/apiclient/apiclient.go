package apiclient

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"

	clusterworkflowtmplpkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/clusterworkflowtemplate"
	cronworkflowpkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/cronworkflow"
	infopkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/info"
	workflowpkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	workflowarchivepkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflowarchive"
	workflowtemplatepkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflowtemplate"
	"github.com/argoproj/argo-workflows/v3/util/instanceid"
)

type Client interface {
	NewArchivedWorkflowServiceClient() (workflowarchivepkg.ArchivedWorkflowServiceClient, error)
	NewWorkflowServiceClient() workflowpkg.WorkflowServiceClient
	NewCronWorkflowServiceClient() (cronworkflowpkg.CronWorkflowServiceClient, error)
	NewWorkflowTemplateServiceClient() (workflowtemplatepkg.WorkflowTemplateServiceClient, error)
	NewClusterWorkflowTemplateServiceClient() (clusterworkflowtmplpkg.ClusterWorkflowTemplateServiceClient, error)
	NewInfoServiceClient() (infopkg.InfoServiceClient, error)
}

type Opts struct {
	ArgoServerOpts ArgoServerOpts
	InstanceID     string
	AuthSupplier   func() string
	// DEPRECATED: use `ClientConfigSupplier`
	ClientConfig         clientcmd.ClientConfig
	ClientConfigSupplier func() clientcmd.ClientConfig
	Offline              bool
}

func (o Opts) String() string {
	return fmt.Sprintf("(argoServerOpts=%v,instanceID=%v)", o.ArgoServerOpts, o.InstanceID)
}

// DEPRECATED: use NewClientFromOpts
func NewClient(argoServer string, authSupplier func() string, clientConfig clientcmd.ClientConfig) (context.Context, Client, error) {
	return NewClientFromOpts(context.Background(), Opts{
		ArgoServerOpts: ArgoServerOpts{URL: argoServer},
		AuthSupplier:   authSupplier,
		ClientConfigSupplier: func() clientcmd.ClientConfig {
			return clientConfig
		},
	})
}

func NewClientFromOpts(ctx context.Context, opts Opts) (context.Context, Client, error) {
	log.WithField("opts", opts).Debug("Client options")
	if opts.Offline {
		return newOfflineClient()
	}
	if opts.ArgoServerOpts.URL != "" && opts.InstanceID != "" {
		return nil, nil, fmt.Errorf("cannot use instance ID with Argo Server")
	}
	if opts.ArgoServerOpts.HTTP1 {
		return newHTTP1Client(opts.ArgoServerOpts.GetURL(), opts.AuthSupplier(), opts.ArgoServerOpts.InsecureSkipVerify)
	} else if opts.ArgoServerOpts.URL != "" {
		return newArgoServerClient(opts.ArgoServerOpts, opts.AuthSupplier())
	} else {
		if opts.ClientConfigSupplier != nil {
			opts.ClientConfig = opts.ClientConfigSupplier()
		}

		ctx, client, err := newArgoKubeClient(ctx, opts.ClientConfig, instanceid.NewService(opts.InstanceID))
		return ctx, client, err
	}
}
