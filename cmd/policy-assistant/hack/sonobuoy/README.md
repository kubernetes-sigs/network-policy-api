# Sonobuoy plugin

## Create plugin

```bash
# FIXME use a real image repository
sonobuoy gen plugin \
  --name=policy-assistant \
  --image=docker.io/sonobuoy-policy-assistant:latest \
  --cmd ./run-sonobuoy-plugin.sh \ > policy-assistant-plugin.yaml
```

## Run plugin

```bash
sonobuoy run --plugin policy-assistant-plugin.yaml --wait
```

## Look at results

```bash
outfile=$(sonobuoy retrieve) && \
  mkdir results && tar -xf $outfile -C results
```

Then crack open the `results` dir and have a look!
