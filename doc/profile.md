apiVersion: kubeflow.org/v1beta1
kind: Profile
metadata:
  name: profile-name
spec:
  owner:
    kind: User
    name: userid@email.com
  resourceQuotaSpec:
   hard:
     cpu: "2"
     memory: 2Gi
     requests.nvidia.com/gpu: "1"
     persistentvolumeclaims: "1"
     requests.storage: "5Gi"
---

workspace create --namespace=test --name=test \
  --request-cpu=
workspace list
workspace get <name>
workspace dev <name> \
  --sync-babslbs \
  --ssh \
  --port-forward 

