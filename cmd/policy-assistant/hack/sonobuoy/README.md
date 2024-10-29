# Sonobuoy plugin

## Create plugin

```bash
# FIXME use a real image repository
sonobuoy gen plugin \
  --name=pola \
  --image=docker.io/sonobuoy-pola:latest \
  --cmd ./run-sonobuoy-plugin.sh \ > pola-plugin.yaml
```

## Run plugin

```bash
sonobuoy run --plugin pola-plugin.yaml --wait
```

## Look at results

```bash
outfile=$(sonobuoy retrieve) && \
  mkdir results && tar -xf $outfile -C results
```

Then crack open the `results` dir and have a look!
