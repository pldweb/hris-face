"""Internal face service. Binds to 127.0.0.1 only -- never exposed publicly
(see docs/PRD.md section 9: "verifikasi lokal, tidak pernah lewat network publik").
Stateless: receives one JPEG frame, returns embedding + liveness score. It never
decides match/no-match -- that comparison happens in the Go API against pgvector.
"""

from pathlib import Path

import cv2
import numpy as np
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse

from .face_engine import FaceEngine
from .liveness import LivenessDetector

MODEL_DIR = Path(__file__).parent.parent / "models"

app = FastAPI(title="hris-face-service")
face_engine: FaceEngine | None = None
liveness_detector: LivenessDetector | None = None


@app.on_event("startup")
def load_models() -> None:
    global face_engine, liveness_detector
    face_engine = FaceEngine()
    liveness_detector = LivenessDetector(MODEL_DIR)


@app.get("/healthz")
def healthz() -> dict:
    return {"status": "ok"}


@app.post("/face/analyze")
async def analyze(request: Request) -> JSONResponse:
    body = await request.body()
    if not body:
        raise HTTPException(status_code=400, detail="gambar kosong")

    image = cv2.imdecode(np.frombuffer(body, dtype=np.uint8), cv2.IMREAD_COLOR)
    if image is None:
        raise HTTPException(status_code=400, detail="gambar tidak bisa dibaca")

    assert face_engine is not None and liveness_detector is not None

    result = face_engine.analyze(image)
    if result.face_count == 0:
        return JSONResponse({"embedding": [], "quality_score": 0.0, "liveness_score": 0.0, "face_count": 0})

    liveness_score = liveness_detector.score(image, result.bbox)

    return JSONResponse(
        {
            "embedding": result.embedding,
            "quality_score": result.quality_score,
            "liveness_score": liveness_score,
            "face_count": result.face_count,
            "eye_openness": result.eye_openness,
            "signature": result.signature or [],
        }
    )
