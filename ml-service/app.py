from fastapi import FastAPI 
from pydantic import BaseModel 

app = FastAPI(title = "ML Service")

class RecRequest(BaseModel):
    user_id : str 
    k: int = 10 

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/embed")
async def embed(req : RecRequest):
    # Call model to get Embedding
    from model_stub import get_user_embedding
    emb = get_user_embedding(req.user_id)
    return {"user_id": req.user_id, "embedding_len":len(emb)}

@app.post("/recommend")
async def recommend(req: RecRequest):
    recs = [{"track_id": f"track_{i}", "score":1.0 - i *0.01} for i in range(req.k)]
    return {"user_id": req.user_id, "recommendations": recs}
