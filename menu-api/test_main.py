from fastapi.testclient import TestClient

from main import app

client = TestClient(app)


def test_health():
    resp = client.get("/health")
    assert resp.status_code == 200
    assert resp.json() == {"status": "ok"}


def test_menu_has_items():
    resp = client.get("/menu")
    assert resp.status_code == 200
    assert len(resp.json()["items"]) > 0
