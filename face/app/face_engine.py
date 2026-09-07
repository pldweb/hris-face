"""ArcFace detection + embedding via InsightFace (buffalo_l, ONNX Runtime CPU).

Method is locked per docs/PRD.md section 9: ArcFace 512-dim embeddings,
cosine similarity computed on the Go API side (pgvector <=> operator).
This module only extracts the embedding; it never decides match/no-match.
"""

from dataclasses import dataclass

import numpy as np
from insightface.app import FaceAnalysis

from .eyes import eye_openness, face_signature

MIN_FACE_WIDTH_PX = 112  # docs/PRD.md section 6.1 enrollment quality gate


@dataclass
class DetectedFace:
    embedding: list[float]
    quality_score: float
    face_count: int
    # (x, y, w, h) of the chosen face; the anti-spoof model needs it to crop
    # at its trained scale rather than scoring the whole frame.
    bbox: tuple[int, int, int, int] | None = None
    # Vertical eye opening / face width, recorded for the blink challenge.
    eye_openness: float = 0.0
    # Coarse 8x8 grayscale of the face crop, for detecting a replayed still.
    signature: list[int] | None = None


class FaceEngine:
    def __init__(self) -> None:
        self._app = FaceAnalysis(name="buffalo_l", providers=["CPUExecutionProvider"])
        self._app.prepare(ctx_id=0, det_size=(640, 640))

    def analyze(self, image_bgr: np.ndarray) -> DetectedFace:
        faces = self._app.get(image_bgr)
        face_count = len(faces)

        if face_count == 0:
            return DetectedFace(embedding=[], quality_score=0.0, face_count=0)

        # Largest face wins if more than one is in frame; the caller decides
        # whether multiple faces should reject the attempt outright.
        face = max(faces, key=lambda f: (f.bbox[2] - f.bbox[0]) * (f.bbox[3] - f.bbox[1]))
        x1, y1, x2, y2 = (int(v) for v in face.bbox)
        face_width = x2 - x1
        quality_score = min(1.0, face_width / (MIN_FACE_WIDTH_PX * 2))

        embedding = face.normed_embedding.astype(np.float32).tolist()
        return DetectedFace(
            embedding=embedding,
            quality_score=quality_score,
            face_count=face_count,
            bbox=(x1, y1, face_width, y2 - y1),
            eye_openness=eye_openness(getattr(face, "landmark_2d_106", None), float(face_width)),
            signature=face_signature(image_bgr, (x1, y1, face_width, y2 - y1)),
        )
