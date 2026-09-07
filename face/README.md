# Face service

Internal service: FastAPI + InsightFace (ArcFace `buffalo_l`) + MiniFASNet anti-spoof.
Binds to `127.0.0.1:8000` only — see `docs/PRD.md` section 9.

## Setup

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

First run downloads `buffalo_l` (~326 MB) into `~/.insightface/models/`.

## Anti-spoof models

Not vendored — generated from the upstream Apache-2.0 weights so their provenance
is verifiable rather than pulled as an opaque binary:

```bash
pip install torch --index-url https://download.pytorch.org/whl/cpu
pip install onnxscript
python scripts/export_antispoof_onnx.py
```

Writes `models/2.7_80x80_MiniFASNetV2.onnx` and
`models/4_0_0_80x80_MiniFASNetV1SE.onnx` (~260 KB each). The service refuses to
start without them. `torch` is only needed for this one-off export, not at runtime.

### Two things that will silently break this model

1. **Input is [0, 255], not [0, 1].** Upstream deliberately dropped the `/255` in
   its vendored `to_tensor`. Normalising to [0,1] makes the net emit a near-constant
   class for every input — the service still starts, still returns scores, and passes
   every spoof.
2. **It scores a scaled crop around the face, not the whole frame.** Each model has
   its own crop scale (2.7 and 4.0) baked in from training. The surrounding context
   — screen edges, moiré, paper borders — is most of the signal.

Both are handled in `app/liveness.py` and verified against upstream's own sample
images (real 1.000, fakes 0.18 / 0.002).

## Accuracy

Measured numbers, and what is still unverified, are in `docs/PRD.md` section 3.1.
Short version: face matching FAR is comfortably within target; **the liveness
threshold is not yet validated on real capture conditions and is a go-live blocker.**

## Run

```bash
uvicorn app.main:app --host 127.0.0.1 --port 8000
```
