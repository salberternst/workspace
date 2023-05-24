## create

```
workspace create name --namespace=default \
  --wait-until-ready
```

```
workspace create name --namespace=default \
  --conda-packages=python=3.9 \ 
  --pip-packages=kfp=2.3.1 \
  --wait-until-ready
```

## update

```
workspace update name --namespace=default \
  --request-gpu=1 \
  --request-gpu-type=nvidida.com/gpu
```



workspace list
workspace get <name>
workspace dev <name> \
  --sync-babslbs \
  --ssh \
  --port-forward 

