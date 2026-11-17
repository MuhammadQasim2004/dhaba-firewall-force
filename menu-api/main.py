import json
import os

from fastapi import FastAPI
from fastapi.responses import JSONResponse

import cache
import db

GIT_COMMIT = os.environ.get("GIT_COMMIT", "unknown")
POD_NAME = os.environ.get("POD_NAME", "local")
CACHE_TTL_SECONDS = 30

app = FastAPI()


@app.get("/health")
def health():
    return {"status": "ok"}


@app.get("/info")
def info():
    return {"service": "menu-api", "version": GIT_COMMIT, "pod": POD_NAME}


@app.get("/menu")
def menu():
    cached = cache.get("menu:items")
    if cached is not None:
        return JSONResponse(content=json.loads(cached))

    items = db.fetch_menu()
    payload = {"items": items}
    cache.set("menu:items", json.dumps(payload), ex=CACHE_TTL_SECONDS)
    return payload
