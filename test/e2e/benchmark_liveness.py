"""Measures anti-spoof accuracy against docs/PRD.md section 3 (>95% spoof caught).

Runs the production LivenessDetector over the CelebA-Spoof test set and reports,
per threshold, how many spoofs are caught and how many live faces are wrongly
rejected. The second number matters as much as the first: a detector that
rejects real employees blocks attendance entirely.

    cd face && .venv/bin/python ../test/e2e/benchmark_liveness.py [--limit N]
"""

import argparse
import io
import sys
from pathlib import Path

import numpy as np
import pandas as pd
from PIL import Image

sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "face"))
from app.face_engine import FaceEngine  # noqa: E402
from app.liveness import LivenessDetector  # noqa: E402

SHARD = (
    "https://huggingface.co/datasets/nguyenkhoa/celeba-spoof-for-face-antispoofing-test"
    "/resolve/main/data/test-00000-of-00010.parquet"
)
MODEL_DIR = Path(__file__).resolve().parents[2] / "face" / "models"


def to_bgr(image_bytes: bytes) -> np.ndarray:
    rgb = np.array(Image.open(io.BytesIO(image_bytes)).convert("RGB"))
    return rgb[:, :, ::-1].copy()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--limit", type=int, default=600, help="images per class")
    args = parser.parse_args()

    print("loading CelebA-Spoof test shard...")
    df = pd.read_parquet(SHARD)
    live = df[df.labels == 0].head(args.limit)
    spoof = df[df.labels == 1].head(args.limit)
    print(f"  {len(live)} live, {len(spoof)} spoof")

    engine = FaceEngine()
    detector = LivenessDetector(MODEL_DIR)

    def score_all(frame: pd.DataFrame, label: str) -> np.ndarray:
        scores, fallbacks, missing = [], 0, 0
        for n, (_, row) in enumerate(frame.iterrows(), 1):
            cell = row["cropped_image"]
            if cell is None or cell.get("bytes") is None:
                missing += 1
                continue
            img = to_bgr(cell["bytes"])
            face = engine.analyze(img)
            if face.face_count and face.bbox is not None:
                bbox = face.bbox
            else:
                # These images are already tight face crops (often ~100x150),
                # below what SCRFD reliably detects. The crop itself is the
                # face box, so fall back to it rather than dropping the sample.
                h, w = img.shape[:2]
                bbox = (0, 0, w, h)
                fallbacks += 1
            scores.append(detector.score(img, bbox))
            if n % 200 == 0:
                print(f"  {label} {n}/{len(frame)}")
        print(f"  {label}: {len(scores)} dinilai ({fallbacks} pakai bbox seluruh gambar, {missing} gambar kosong)")
        return np.array(scores)

    live_scores = score_all(live, "live")
    spoof_scores = score_all(spoof, "spoof")

    print(f"\nskor live  : mean {live_scores.mean():.3f}  median {np.median(live_scores):.3f}")
    print(f"skor spoof : mean {spoof_scores.mean():.3f}  median {np.median(spoof_scores):.3f}")

    print(f"\n{'threshold':>10} {'spoof tertangkap %':>19} {'live ditolak %':>16}   catatan")
    for threshold in [0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9]:
        caught = float((spoof_scores < threshold).mean() * 100)
        rejected = float((live_scores < threshold).mean() * 100)
        note = "target PRD >95% tertangkap" if caught >= 95 else ""
        if abs(threshold - 0.5) < 1e-9:
            note = (note + " ").strip() + " (nilai di kode saat ini)"
        print(f"{threshold:>10.2f} {caught:>19.1f} {rejected:>16.1f}   {note}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
