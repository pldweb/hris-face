"""Measures FAR/FRR of the real ArcFace pipeline against docs/PRD.md section 3.

Runs the production FaceEngine (face/app/face_engine.py) -- not a copy -- over
the official LFW verification pairs, then sweeps the cosine threshold to find
where the PRD's targets (FAR < 0.1%, FRR < 3%) actually land.

    cd face && .venv/bin/python ../test/e2e/benchmark_accuracy.py [--limit N]
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

PAIRS_URL = "https://huggingface.co/datasets/logasja/lfw/resolve/main/pairs/test-00000-of-00001.parquet"


def to_bgr(image_bytes: bytes) -> np.ndarray:
    rgb = np.array(Image.open(io.BytesIO(image_bytes)).convert("RGB"))
    return rgb[:, :, ::-1].copy()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--limit", type=int, default=0, help="only the first N pairs")
    args = parser.parse_args()

    print("loading LFW pairs...")
    df = pd.read_parquet(PAIRS_URL)
    if args.limit:
        # Interleave so a truncated run keeps both classes.
        df = pd.concat([df[df.pair == 1].head(args.limit // 2), df[df.pair == 0].head(args.limit // 2)])

    print("loading ArcFace (buffalo_l)...")
    engine = FaceEngine()

    genuine, impostor = [], []
    undetected = 0

    for n, (_, row) in enumerate(df.iterrows(), 1):
        a = engine.analyze(to_bgr(row["img_0"]["bytes"]))
        b = engine.analyze(to_bgr(row["img_1"]["bytes"]))

        if a.face_count == 0 or b.face_count == 0:
            undetected += 1
            continue

        # Same normalised embeddings and same metric the Go side uses:
        # pgvector's 1 - (a <=> b) is exactly this dot product.
        score = float(np.dot(a.embedding, b.embedding))
        (genuine if row["pair"] == 1 else impostor).append(score)

        if n % 200 == 0:
            print(f"  {n}/{len(df)} pairs")

    genuine_scores = np.array(genuine)
    impostor_scores = np.array(impostor)

    print(f"\npairs scored: {len(genuine_scores)} sama-orang, {len(impostor_scores)} beda-orang")
    print(f"wajah tak terdeteksi (pasangan dilewati): {undetected}")
    print(f"skor sama-orang : mean {genuine_scores.mean():.3f}  min {genuine_scores.min():.3f}")
    print(f"skor beda-orang : mean {impostor_scores.mean():.3f}  max {impostor_scores.max():.3f}")

    print(f"\n{'threshold':>10} {'FAR %':>9} {'FRR %':>9}   catatan")
    best = None
    for threshold in np.arange(0.10, 0.86, 0.05):
        far = float((impostor_scores >= threshold).mean() * 100)
        frr = float((genuine_scores < threshold).mean() * 100)
        note = ""
        if far < 0.1 and frr < 3.0 and best is None:
            best = (threshold, far, frr)
            note = "<- memenuhi target PRD"
        if abs(threshold - 0.45) < 1e-9:
            note = (note + "  ").strip() + " (nilai di kode saat ini)"
        print(f"{threshold:>10.2f} {far:>9.3f} {frr:>9.3f}   {note}")

    print()
    if best:
        print(f"Threshold terendah yang memenuhi FAR<0.1% dan FRR<3%: {best[0]:.2f} "
              f"(FAR {best[1]:.3f}%, FRR {best[2]:.3f}%)")
    else:
        print("Tidak ada threshold yang memenuhi kedua target PRD pada dataset ini.")

    far_at_current = float((impostor_scores >= 0.45).mean() * 100)
    frr_at_current = float((genuine_scores < 0.45).mean() * 100)
    print(f"Pada threshold 0.45 yang dipakai kode: FAR {far_at_current:.3f}%, FRR {frr_at_current:.3f}%")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
