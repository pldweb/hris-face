"""Exports the official Silent-Face-Anti-Spoofing weights to ONNX.

Provenance matters here: anti-spoofing is what stops buddy punching, and a
wrong or tampered model fails silently -- it just starts passing spoofs. So the
weights come from the upstream Apache-2.0 repo and the ONNX is produced here,
rather than pulled as an opaque binary from an unvetted mirror.

    pip install torch --index-url https://download.pytorch.org/whl/cpu
    python face/scripts/export_antispoof_onnx.py

Writes face/models/{2.7_80x80_MiniFASNetV2,4_0_0_80x80_MiniFASNetV1SE}.onnx
"""

import sys
import urllib.request
from pathlib import Path

import torch

REPO = "https://raw.githubusercontent.com/minivision-ai/Silent-Face-Anti-Spoofing/master"
WEIGHTS_BASE = f"{REPO}/resources/anti_spoof_models"

# (weight file, MiniFASNet class name, crop scale used at inference)
MODELS = [
    ("2.7_80x80_MiniFASNetV2.pth", "MiniFASNetV2", 2.7),
    ("4_0_0_80x80_MiniFASNetV1SE.pth", "MiniFASNetV1SE", 4.0),
]

OUT_DIR = Path(__file__).resolve().parent.parent / "models"
WORK_DIR = Path("/tmp/spoof_export")


def fetch(url: str, dest: Path) -> Path:
    dest.parent.mkdir(parents=True, exist_ok=True)
    if not dest.exists():
        print(f"  downloading {dest.name}")
        urllib.request.urlretrieve(url, dest)
    return dest


def main() -> int:
    WORK_DIR.mkdir(parents=True, exist_ok=True)
    OUT_DIR.mkdir(parents=True, exist_ok=True)

    # Upstream architecture definition, used as-is so the weights load exactly.
    fetch(f"{REPO}/src/model_lib/MiniFASNet.py", WORK_DIR / "MiniFASNet.py")
    sys.path.insert(0, str(WORK_DIR))
    import MiniFASNet as arch  # noqa: E402

    for weight_name, class_name, scale in MODELS:
        print(f"{weight_name} (scale {scale})")
        weight_path = fetch(f"{WEIGHTS_BASE}/{weight_name}", WORK_DIR / weight_name)

        # get_kernel(80, 80) -> (5, 5); inlined to avoid another upstream file.
        model = getattr(arch, class_name)(conv6_kernel=(5, 5))

        state_dict = torch.load(weight_path, map_location="cpu")
        if next(iter(state_dict)).startswith("module."):
            state_dict = {k[7:]: v for k, v in state_dict.items()}
        model.load_state_dict(state_dict)
        model.eval()

        out_path = OUT_DIR / weight_name.replace(".pth", ".onnx")
        torch.onnx.export(
            model,
            torch.randn(1, 3, 80, 80),
            str(out_path),
            input_names=["input"],
            output_names=["logits"],
            dynamic_axes={"input": {0: "batch"}, "logits": {0: "batch"}},
            opset_version=11,
        )
        print(f"  wrote {out_path} ({out_path.stat().st_size // 1024} KB)")

    print("\nDone. The service reads these from face/models/.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
