"""Eye openness, used for the blink challenge (docs/PRD.md F5).

Passive anti-spoof scores one frame. A blink challenge instead asks for a
*change* over several frames, which a printed photo or a still on a screen
cannot produce. The measurement has to happen server-side: a client that
reports "yes, they blinked" is trivially bypassed.

Index ranges are the InsightFace 106-point layout, confirmed empirically against
the 5-point kps eye centres: left eye 33-42, right eye 87-96.
"""

import numpy as np

LEFT_EYE = slice(33, 43)
RIGHT_EYE = slice(87, 97)


def eye_openness(landmarks: np.ndarray, face_width: float) -> float:
    """Mean vertical eye opening as a fraction of face width.

    Normalising by face width keeps the number comparable across distances from
    the camera, so a threshold means the same thing whether someone leans in or
    sits back.
    """
    if landmarks is None or len(landmarks) < 97 or face_width <= 0:
        return 0.0

    openings = []
    for eye in (LEFT_EYE, RIGHT_EYE):
        pts = landmarks[eye]
        openings.append(float(pts[:, 1].max() - pts[:, 1].min()))

    return float(np.mean(openings) / face_width)


def face_signature(image_bgr, bbox) -> list[int]:
    """An 8x8 grayscale thumbnail of the face crop.

    The blink challenge needs to prove the frames are not the same still image
    held up to the camera. Embeddings cannot do that -- two frames of one person
    are near-identical by design -- so the comparison happens on pixels, coarse
    enough to ignore noise and compression but fine enough that a moving face
    changes it.
    """
    import cv2

    x, y, w, h = bbox
    x0, y0 = max(0, x), max(0, y)
    x1, y1 = min(image_bgr.shape[1], x + w), min(image_bgr.shape[0], y + h)
    crop = image_bgr[y0:y1, x0:x1]
    if crop.size == 0:
        return []

    gray = cv2.cvtColor(crop, cv2.COLOR_BGR2GRAY)
    small = cv2.resize(gray, (8, 8), interpolation=cv2.INTER_AREA)
    return [int(v) for v in small.flatten()]
