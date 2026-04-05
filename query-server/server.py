"""
BGE-small-zh-v1.5 Embedding RPC Service
"""
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
from typing import List, Optional
import uvicorn
import os

# 模型路径
MODEL_PATH = os.path.join(os.path.dirname(__file__), "models", "bge-small-zh-v1.5")

# 加载模型
model = SentenceTransformer(MODEL_PATH)

app = FastAPI(title="Embedding Service", version="1.0.0")


class EmbeddingRequest(BaseModel):
    texts: List[str]
    normalize: bool = True


class EmbeddingResponse(BaseModel):
    embeddings: List[List[float]]
    dimension: int
    count: int


class SimilarityRequest(BaseModel):
    text1: str
    text2: str


class SimilarityResponse(BaseModel):
    similarity: float


@app.get("/health")
async def health_check():
    """健康检查"""
    return {"status": "ok", "model": "bge-small-zh-v1.5"}


@app.post("/embed", response_model=EmbeddingResponse)
async def get_embeddings(request: EmbeddingRequest):
    """
    获取文本的 embedding 向量
    """
    embeddings = model.encode(
        request.texts,
        normalize_embeddings=request.normalize,
        convert_to_numpy=True
    )
    
    return EmbeddingResponse(
        embeddings=embeddings.tolist(),
        dimension=embeddings.shape[1],
        count=len(embeddings)
    )


@app.post("/similarity", response_model=SimilarityResponse)
async def compute_similarity(request: SimilarityRequest):
    """
    计算两个文本的相似度
    """
    embeddings = model.encode(
        [request.text1, request.text2],
        normalize_embeddings=True,
        convert_to_numpy=True
    )
    
    # 计算余弦相似度
    similarity = float(embeddings[0] @ embeddings[1])
    
    return SimilarityResponse(similarity=similarity)


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)