package translator

import (
    "io/ioutil"
    "log"
    "os"

    "knative.dev/pkg/apis"
    duckv1 "knative.dev/pkg/apis/duck/v1"
    "sigs.k8s.io/yaml"

    "knflows/sw/internal/types"
)

func Translate(filename string) error {
    bytes, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }

    var sw types.SW
    err = yaml.Unmarshal(bytes, &sw)

    if err != nil {
        return err
        log.Fatal(err)
    }

    Normalize(&sw)

    triggers := newTriggerTarget(sw)

    err = translateSW(sw, &triggers)
    if err != nil {
        return err
    }

    triggers.emit(os.Stdout)
    return nil
}

func translateSW(sw types.SW, triggers *TriggerTarget) error {
    for _, state := range sw.States {
        translateState(state, triggers)
    }
    return nil
}

func translateState(state types.State, triggers *TriggerTarget) error {
    switch state.Type {
    case "inject":
        return translateInject(state, triggers)
    }

    return nil
}

func translateInject(state types.State, triggers *TriggerTarget) error {
    url, _ := apis.ParseURL("?state=" + MakeK8sName(state.Name))
    triggers.AddTrigger(state.Name, duckv1.Destination{
        Ref: triggers.GetMachineRef(),
        URI: url,
    })

    return nil
}
