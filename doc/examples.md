## create

```
workspace create name --namespace=default \
  --wait-until-ready
```

```
workspace create name --namespace=default \
  --init-conda-packages=python=3.10
  --install-conda-packages=python=3.9 \ 
  --install-conda-packages=cmake \
  --install-pip-package=kfp===2. \
  --wait-until-ready
```

## update

```
workspace update name --namespace=default \
  --request-gpu=1 \
  --request-gpu-type=nvidida.com/gpu
```

## dev

```
workspace dev name --namespace=default \
  --sync-folder=.:/home/workspace/data
```

Interactive shell?


