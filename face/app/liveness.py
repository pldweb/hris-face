"""Passive liveness (anti-spoof), MiniFASNet ensemble.

Mirrors the upstream inference recipe exactly, because this model is picky
about how it is fed: it scores a face crop expanded by a fixed scale factor
(2.7 for V2, 4.0 for V1SE) and resized to 80x80 -- NOT the whole frame. Feeding
it a full frame produces confident nonsense, which is the worst failure mode
here: spoofs pass silently.

Two models are ensembled and their softmax outputs summed, as upstream does.
Class 1 is "real"; classes 0 and 2 are the two spoof families (print/screen).

Weights: minivision-ai/Silent-Face-Anti-Spoofing (Apache-2.0), converted to
ONNX by face/scripts/export_antispoof_onnx.py.
"""

from pathlib import Path

import cv2
import numpy as np
import onnxruntime as ort

INPUT_SIZE = 80
REAL_CLASS = 1

# (filename, crop scale) -- the scale is baked into how each model was trained.
ENSEMBLE = [
    ("2.7_80x80_MiniFASNetV2.onnx", 2.7),
    ("4_0_0_80x80_MiniFASNetV1SE.onnx", 4.0),
]


class LivenessDetector:
    def __init__(self, model_dir: Path) -> None:
        self._models = []
        for filename, scale in ENSEMBLE:
            path = model_dir / filename
            if not path.exists():
                raise FileNotFoundError(
                    f"Model anti-spoof tidak ditemukan: {path}. "
                    "Jalankan: python face/scripts/export_antispoof_onnx.py"
                )
            session = ort.InferenceSession(str(path), providers=["CPUExecutionProvider"])
            self._models.append((session, session.get_inputs()[0].name, scale))

    def score(self, image_bgr: np.ndarray, bbox: tuple[int, int, int, int]) -> float:
        """Real-face probability in [0, 1] for the face at bbox (x, y, w, h)."""
        summed = np.zeros(3, dtype=np.float64)
        for session, input_name, scale in self._models:
            patch = _crop(image_bgr, bbox, scale, INPUT_SIZE)
            # NOT divided by 255: upstream deliberately dropped that step in its
            # vendored to_tensor ("# return img.float().div(255)  modify by zkx"),
            # so these weights expect raw [0,255] input. Scaling to [0,1] makes
            # the net output a near-constant class regardless of the image --
            # it looks like it works, and silently passes every spoof.
            blob = patch.astype(np.float32).transpose(2, 0, 1)[np.newaxis, ...]
            logits = session.run(None, {input_name: blob})[0][0]
            summed += _softmax(logits)

        return float(summed[REAL_CLASS] / len(self._models))


def _crop(img: np.ndarray, bbox: tuple[int, int, int, int], scale: float, out: int) -> np.ndarray:
    """Upstream CropImage: expand the box by `scale` about its centre, clamped
    to the frame, then resize. Reimplemented rather than imported so the
    service carries no dependency on the training repo."""
    src_h, src_w = img.shape[:2]
    x, y, box_w, box_h = bbox

    scale = min((src_h - 1) / box_h, min((src_w - 1) / box_w, scale))
    new_w, new_h = box_w * scale, box_h * scale
    center_x, center_y = x + box_w / 2, y + box_h / 2

    left = center_x - new_w / 2
    top = center_y - new_h / 2
    right = center_x + new_w / 2
    bottom = center_y + new_h / 2

    # Shift rather than clip, so the crop keeps the trained aspect and size.
    if left < 0:
        right -= left
        left = 0
    if top < 0:
        bottom -= top
        top = 0
    if right > src_w - 1:
        left -= right - src_w + 1
        right = src_w - 1
    if bottom > src_h - 1:
        top -= bottom - src_h + 1
        bottom = src_h - 1

    patch = img[int(top) : int(bottom) + 1, int(left) : int(right) + 1]
    return cv2.resize(patch, (out, out))


def _softmax(x: np.ndarray) -> np.ndarray:
    e = np.exp(x - np.max(x))
    return e / e.sum()
