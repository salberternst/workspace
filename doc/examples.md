## create

```
workspace create name --namespace=default \
  --wait-until-ready
```

```
workspace create name --namespace=default \
  --install-conda-packages=python=3.9 \ 
  --install-conda-packages=cmake \
  --init-command='pip install kfp --pre' \
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


