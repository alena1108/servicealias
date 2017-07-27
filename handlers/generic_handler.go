package handlers

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/rancher/kubernetes-agent/kubernetesclient"
	"github.com/rancher/kubernetes-model/model"
)

type GenericHandler struct {
	client      *kubernetesclient.Client
	kindHandled string
}

func NewHandler(kubernetesClient *kubernetesclient.Client, kindHandled string) *GenericHandler {
	return &GenericHandler{
		client:      kubernetesClient,
		kindHandled: kindHandled,
	}
}

func (h *GenericHandler) Handle(event model.WatchEvent) error {

	if i, ok := event.Object.(map[string]interface{}); ok {
		if h.kindHandled == "servicealias" {
			switch event.Type {
			case "MODIFIED":
				logrus.Info("Service alias is modified %v", i)
			case "ADDED":
				logrus.Info("Service alias is added %v", i)
			case "DELETED":
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
