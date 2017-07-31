package handlers

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/rancher/kubernetes-agent/kubernetesclient"
	"github.com/rancher/kubernetes-model/model"
)

type GenericHandler struct {
	client      *kubernetesclient.Client
	kindHandled string
}

type ServiceAlias struct {
	Metadata *model.ObjectMeta
	Spec     *AliasSpec
}

type AliasSpec struct {
	Links []string
}

func NewHandler(kubernetesClient *kubernetesclient.Client, kindHandled string) *GenericHandler {
	return &GenericHandler{
		client:      kubernetesClient,
		kindHandled: kindHandled,
	}
}

func (h *GenericHandler) Handle(event model.WatchEvent) error {

	if i, ok := event.Object.(map[string]interface{}); ok {
		if h.kindHandled == "servicealiases" {
			var alias ServiceAlias
			err := mapstructure.Decode(i, &alias)
			if err != nil {
				return fmt.Errorf("Failed to decode service alias due to %v", err)
			}
			name := alias.Metadata.Name
			switch event.Type {
			case "MODIFIED":
				logrus.Info("Service alias [%s] is modified", name)
			case "ADDED":
				logrus.Info("Service alias [%s] is added", name)
				_, err := h.createServiceForAlias(alias)
				if err != nil {
					return fmt.Errorf("Failed to create service [%s] due to %v", name, err)
				}
				logrus.Info("Created service for service alias")
			case "DELETED":
				if err := h.deleteServiceForAlias(alias); err != nil {
					return fmt.Errorf("Failed to delete service [%s] due to %v", name, err)
				}
				logrus.Info("Service alias is deleted %v", i)
			default:
				return nil
			}
		} else {
			return fmt.Errorf("Unrecognized handled kind [%s]", h.kindHandled)
		}
		return nil

	}
	return fmt.Errorf("Couldn't decode event [%#v]", event)
}

func (h *GenericHandler) GetKindHandled() string {
	return h.kindHandled
}

func (h *GenericHandler) deleteServiceForAlias(alias ServiceAlias) error {
	_, err := h.client.Service.DeleteService("default", alias.Metadata.Name)
	return err
}

func (h *GenericHandler) createServiceForAlias(alias ServiceAlias) (*model.Service, error) {
	meta := &model.ObjectMeta{
		Name: alias.Metadata.Name,
	}

	var epAddresses []model.EndpointAddress
	var epPorts []model.EndpointPort
	for _, linkedServiceName := range alias.Spec.Links {
		linkedService, err := h.client.Service.ByName("default", linkedServiceName)
		if err != nil {
			return nil, fmt.Errorf("Failed to get service by name [%s]", linkedServiceName)
		}

		for _, port := range linkedService.Spec.Ports {
			epPort := model.EndpointPort{
				Port: port.Port,
				Name: port.Name,
			}
			epPorts = append(epPorts, epPort)
		}
		epAddress := model.EndpointAddress{
			Ip: linkedService.Spec.ClusterIP,
		}
		epAddresses = append(epAddresses, epAddress)
	}
	ports := make([]model.ServicePort, 0)
	port := model.ServicePort{
		Protocol:   "TCP",
		Port:       80,
		TargetPort: 80,
		Name:       "http",
	}
	ports = append(ports, port)
	spec := &model.ServiceSpec{
		SessionAffinity: "None",
		Type:            "ClusterIP",
		Ports:           ports,
	}
	svc := &model.Service{
		Metadata: meta,
		Spec:     spec,
	}

	eps := &model.Endpoints{
		Metadata: &model.ObjectMeta{},
	}
	subset := model.EndpointSubset{
		Addresses: epAddresses,
		Ports:     epPorts,
	}
	subsets := []model.EndpointSubset{subset}
	eps.Subsets = subsets
	eps.Metadata.Name = alias.Metadata.Name
	// create endpoints
	_, err := h.client.Service.CreateEndpoint("default", eps)
	if err != nil {
		return nil, err
	}
	// create service
	return h.client.Service.CreateService("default", svc)
}

func (h *GenericHandler) createEndpoints(eps *model.Endpoints) error {
	_, err := h.client.Service.CreateEndpoint("default", eps)
	return err
}
