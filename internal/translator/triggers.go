package translator

import (
    "encoding/json"
    "io"

    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    kne "knative.dev/eventing/pkg/apis/eventing/v1"
    duckv1 "knative.dev/pkg/apis/duck/v1"
    kns "knative.dev/serving/pkg/apis/serving/v1"
    "sigs.k8s.io/yaml"

    "knflows/sw/internal/types"
)

type TriggerTarget struct {
    machine  kns.Service
    triggers []kne.Trigger
}

func newTriggerTarget(sw types.SW) TriggerTarget {
    asjson, _ := json.Marshal(sw)

    machine := kns.Service{
        TypeMeta: metav1.TypeMeta{
            Kind:       "Service",
            APIVersion: "serving.knative.dev/v1",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name: MakeK8sName(sw.ID),
        },
        Spec: kns.ServiceSpec{
            ConfigurationSpec: kns.ConfigurationSpec{Template: kns.RevisionTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Annotations: map[string]string{
                        "autoscaling.knative.dev/min-scale": "1",
                    },
                },
                Spec: kns.RevisionSpec{
                    PodSpec: corev1.PodSpec{
                        Containers: []corev1.Container{
                            {
                                Image:           "docker.io/villardl/sw",
                                ImagePullPolicy: corev1.PullAlways,
                                Env: []corev1.EnvVar{{
                                    Name:  "STATES",
                                    Value: string(asjson),
                                }, {
                                    Name:  "BROKER",
                                    Value: "http://broker-ingress.knative-eventing.svc.cluster.local/sw/default",
                                }},
                            },
                        },
                    },
                },
            }},
        },
    }

    return TriggerTarget{
        machine: machine,
    }
}

func (t *TriggerTarget) GetMachineRef() *duckv1.KReference {
    return &duckv1.KReference{
        Kind:       "Service",
        Name:       t.machine.Name,
        APIVersion: "serving.knative.dev/v1",
    }

}

func (t *TriggerTarget) AddTrigger(name string, subscriber duckv1.Destination) {
    trigger := kne.Trigger{
        TypeMeta: metav1.TypeMeta{
            Kind:       "Trigger",
            APIVersion: "eventing.knative.dev/v1",
        },
        ObjectMeta: metav1.ObjectMeta{
            Name: MakeK8sName(name),
        },
        Spec: kne.TriggerSpec{
            Broker:     "default", // TODO: configurable
            Subscriber: subscriber,
            Delivery:   nil,
        },
    }
    t.triggers = append(t.triggers, trigger)
}

func (t *TriggerTarget) emit(writer io.Writer) {
    b, _ := yaml.Marshal(t.machine)
    writer.Write(b)
    for _, trigger := range t.triggers {
        writer.Write([]byte("---\n"))
        b, _ = yaml.Marshal(trigger)
        writer.Write(b)
    }
}
