package service

import (
	"context"
	"feather/types"
	"fmt"
	"log"
	"strings"
	"text/template"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	wfclientset "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ArgoWorkflowService interface {
	CreateArgoWorkflowScript(req *types.JobBasedJavaRequest) (string, error)
	renderWorkflowScript(req *types.JobBasedJavaRequest) (string, error)
	CreateArgoWorkflowsJobBasedSpringBoot(req *types.JobBasedJavaRequest) error
}

type argoWorkflowServiceImpl struct {
}

func NewArgoWorkflowSerivce() ArgoWorkflowService {
	return &argoWorkflowServiceImpl{}
}

func (s *argoWorkflowServiceImpl) CreateArgoWorkflowScript(req *types.JobBasedJavaRequest) (string, error) {
	return s.renderWorkflowScript(req)
}

func (s *argoWorkflowServiceImpl) renderWorkflowScript(req *types.JobBasedJavaRequest) (string, error) {
	data := struct {
		JDK, BuildTool, URL, ImageRegistry, ImageName, ImageTag string
	}{
		JDK:           req.JDK,
		BuildTool:     req.BuildTool,
		URL:           req.URL,
		ImageRegistry: req.ImageRegistry,
		ImageName:     req.ImageName,
		ImageTag:      req.ImageTag,
	}

	tmpl, err := template.ParseFiles("assets/templates/argo/ci.tmpl")
	if err != nil {
		return "", fmt.Errorf("template parsing error: %w", err)
	}

	var out strings.Builder
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}
	return out.String(), nil
}

/*
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:

	generateName: $name
	namespace: $namespace

spec:

	entrypoint: build-and-push
	templates:
	- name: build-and-push
		container:
			image: docker:24.0.5-dind
			command: ["sh", "-c"]
			args: ~template~
			securityContext:
			  privileged: true
*/
func (s *argoWorkflowServiceImpl) CreateArgoWorkflowsJobBasedSpringBoot(req *types.JobBasedJavaRequest) error {
	config, err := GetKubeConfig()
	if err != nil {
		return fmt.Errorf("kubeconfig load failed: %w", err)
	}

	command, err := s.renderWorkflowScript(req)
	if err != nil {
		return fmt.Errorf("workflow command render failed: %w", err)
	}

	wfClient := wfclientset.NewForConfigOrDie(config).ArgoprojV1alpha1().Workflows(req.Namespace)
	workflow := newSpringBootWorkflow(req.Name, command)

	result, err := wfClient.Create(context.TODO(), workflow, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("workflow creation failed: %w", err)
	}

	log.Printf("Workflow submitted: %s", result.Name)
	return nil
}

func newSpringBootWorkflow(name, command string) *wfv1.Workflow {
	return &wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
		},
		Spec: wfv1.WorkflowSpec{
			Entrypoint: "build-and-push",
			Templates: []wfv1.Template{
				{
					Name: "build-and-push",
					Container: &corev1.Container{
						Image:   "docker:24.0.5-dind",
						Command: []string{"sh", "-c"},
						Args:    []string{command},
						SecurityContext: &corev1.SecurityContext{
							Privileged: boolPtr(true),
						},
					},
				},
			},
		},
	}
}

func boolPtr(b bool) *bool {
	return &b
}
