## create

```
workspace create name --namespace=default \
  --wait-until-ready
```

```
workspace create name --namespace=default \
  --install-conda-packages=python=3.9 \ 
  --install-pip-packages=kfp=2.3.1,clearml,jupyter \
  --wait-until-ready
```

## update

```
workspace update name --namespace=default \
  --request-gpu=1 \
  --request-gpu-type=nvidida.com/gpu
```

## run

```
workspace dev name --namespace=default \
  --sync-folder=.:/home/workspace/data \
  --forward-port=8888:8888
```

Interactive shell?

```
exec 
```

workspace list
workspace get <name>
workspace dev <name> \
  --sync-babslbs \
  --ssh \
  --port-forward 

