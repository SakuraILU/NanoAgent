"""
Request/Response 数据模型定义
"""
from pydantic import BaseModel
from typing import List, Optional


# ============ Embedding ============

class EmbedRequest(BaseModel):
    texts: List[str]
    normalize: bool = True


class EmbedResponse(BaseModel):
    embeddings: List[List[float]]
    dimension: int
    count: int


class SimilarityRequest(BaseModel):
    text1: str
    text2: str


class SimilarityResponse(BaseModel):
    similarity: float


# ============ Tokenizer ============

class SegmentRequest(BaseModel):
    texts: List[str]
    size: int = 10  # 每个文本返回的重要词数量


class SegmentResponse(BaseModel):
    results: List[dict]


# ============ NER ============

class NERRequest(BaseModel):
    texts: List[str]


# ============ Text Analysis ============

class TextAnalysisRequest(BaseModel):
    texts: List[str]
    keywords: bool = True
    ner: bool = False
    embedding: bool = False
    keyword_top_k: int = 5


class TextAnalysisResponse(BaseModel):
    results: List[dict]