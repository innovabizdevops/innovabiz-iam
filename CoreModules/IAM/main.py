from fastapi import FastAPI
import os

app = FastAPI(title="InnovaBiz IAM Service")

@app.get("/health")
def health():
    return {"status": "ok", "service": "IAM"}

@app.get("/")
def root():
    return {"message": "IAM Service up and running!"}

# Exemplo de endpoint protegido, ajuste conforme autenticação real
@app.get("/whoami")
def whoami():
    user = os.getenv("IAM_DB_USER", "unknown")
    return {"user": user}
