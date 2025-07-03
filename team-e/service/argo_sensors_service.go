package service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"

	"feather/types"

	"github.com/Masterminds/sprig/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

type ArgoSensorService interface {
	CreateArgoSensor(req *types.JobBasedJavaRequest) error
	renderSensorTemplate(req *types.JobBasedJavaRequest, workflowScript string) ([]byte, error)
}

type argoSensorServiceImpl struct {
	argoWorkflowService ArgoWorkflowService
}

func NewArgoSensorService(argoWorkflowService ArgoWorkflowService) ArgoSensorService {
	return &argoSensorServiceImpl{
		argoWorkflowService: argoWorkflowService,
	}
}

func (s *argoSensorServiceImpl) CreateArgoSensor(req *types.JobBasedJavaRequest) error {
	log.Println("=======0=======")
	config, err := GetKubeConfig()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	log.Println("=======1=======")

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}
	log.Println("=======2=======")

	workflowScript, err := s.argoWorkflowService.CreateArgoWorkflowScript(req)
	if err != nil {
		return fmt.Errorf("failed to generate Argo workflow script: %w", err)
	}
	log.Println("=======3=======")

	sensorYaml, err := s.renderSensorTemplate(req, workflowScript)
	if err != nil {
		return fmt.Errorf("failed to render sensor template: %w", err)
	}
	log.Println("=======4=======")

	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(sensorYaml, &obj); err != nil {
		return fmt.Errorf("failed to decode sensor YAML: %w", err)
	}
	log.Println("=======5=======")

	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "sensors",
	}
	log.Println("=======6=======")

	resource, err := client.Resource(gvr).Namespace(req.Namespace).Create(context.Background(), &obj, metav1.CreateOptions{})
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to create sensor resource: %w", err)
	}
	log.Println("=======7=======")

	log.Printf("Sensor created: %s", resource.GetName())
	return nil
}

func (s *argoSensorServiceImpl) renderSensorTemplate(req *types.JobBasedJavaRequest, workflowScript string) ([]byte, error) {
	params := struct {
		Namespace          string
		SensorName         string
		ServiceAccountName string
		EventSourceName    string
		EventName          string
		TriggerName        string
		WorkflowNamePrefix string
		WorkflowName       string
		WorkflowScript     string
		DockerUser         string
	}{
		Namespace:          req.Namespace,
		SensorName:         req.Name + "-sensor",
		ServiceAccountName: "gitea-sensor-sa",
		EventSourceName:    "gitea-webhook",
		EventName:          "gitea",
		TriggerName:        "argo-workflow-trigger",
		WorkflowNamePrefix: req.Name + "-",
		WorkflowName:       "build-and-push",
		WorkflowScript:     workflowScript,
		DockerUser:         "DockerUser",
	}

	tmpl, err := template.New("sensor.tmpl").Funcs(sprig.TxtFuncMap()).ParseFiles("assets/templates/argo/sensor.tmpl")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
