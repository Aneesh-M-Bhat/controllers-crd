# controller-crd

## How to Run

either give direct path of kubeconfig in -kubeconfig flag or import it using the below command

```
cat ~/.kube/config > config
```

and then run the go controller using

```
go build && ./crds-controller -kubeconfig=config
```

# Things to check
- install crd & cr in cluster
  ```
  kubectl apply -f resources/crd.yaml
  kubectl apply -f resources/cr.yaml
  ```
- Req Output: should create a deployment & respective number of pods
- Upon editing the cr i.e. foo the pods should also change respectively
# controllers-crd
