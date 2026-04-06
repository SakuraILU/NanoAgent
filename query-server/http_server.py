"""
Query Server - 文本处理 RPC 服务

功能：
- Embedding 向量化
- 文本分词与关键词提取
- 命名实体识别
"""
from fastapi import FastAPI
import uvicorn

from service import (
    EmbeddingService, TokenizerService, NERService,
    EmbedRequest, EmbedResponse,
    SimilarityRequest, SimilarityResponse,
    SegmentRequest, SegmentResponse,
    NERRequest,
    TextAnalysisRequest, TextAnalysisResponse
)

# 初始化服务
embedding_service = EmbeddingService()
tokenizer_service = TokenizerService()
# NER 服务延迟加载（加载较慢）
ner_service = None


def get_ner_service():
    """延迟加载 NER 服务"""
    global ner_service
    if ner_service is None:
        ner_service = NERService()
    return ner_service


app = FastAPI(title="Query Server", version="1.0.0")


# ============ Health Check ============

@app.get("/health")
async def health_check():
    """健康检查"""
    return {"status": "ok", "service": "query-server"}


# ============ Embedding APIs ============

@app.post("/embed", response_model=EmbedResponse)
async def get_embeddings(request: EmbedRequest):
    """获取文本的 embedding 向量"""
    embeddings = embedding_service.encode(
        request.texts,
        normalize=request.normalize
    )
    
    return EmbedResponse(
        embeddings=embeddings,
        dimension=embedding_service.dimension,
        count=len(embeddings)
    )


@app.post("/similarity", response_model=SimilarityResponse)
async def compute_similarity(request: SimilarityRequest):
    """计算两个文本的相似度"""
    similarity = embedding_service.similarity(request.text1, request.text2)
    return SimilarityResponse(similarity=similarity)


# ============ Tokenizer APIs ============

@app.post("/segment", response_model=SegmentResponse)
async def segment(request: SegmentRequest):
    """
    分词
    
    对每个文本分词后，统计词频并筛选重要词。
    适用于缩短长 query 场景，通过 size 参数控制返回词数。
    """
    results = tokenizer_service.segment(request.texts, size=request.size)
    return {"results": results}


# ============ NER APIs ============

@app.post("/ner")
async def named_entity_recognition(request: NERRequest):
    """
    命名实体识别
    
    识别文本中的人名、地名、机构名等实体
    """
    ner = get_ner_service()
    results = ner.recognize(request.texts)
    return {"results": results}


# ============ Text Analysis APIs ============

@app.post("/analyze", response_model=TextAnalysisResponse)
async def analyze_text(request: TextAnalysisRequest):
    """
    综合文本分析(分词、实体识别、向量)
    """
    results = []
    
    for text in request.texts:
        result = {"text": text}
        print(result)
        
        # 关键词
        if request.keywords:
            kw_result = tokenizer_service.segment([text], size=request.keyword_top_k)
            result["keywords"] = kw_result[0]["words"]
        
        # 实体识别
        if request.ner:
            ner = get_ner_service()
            ner_result = ner.recognize([text])
            result["entities"] = ner_result[0]["entities"]
        
        # 向量化
        if request.embedding:
            emb = embedding_service.encode([text], normalize=True)
            result["embedding"] = emb[0]
            result["dimension"] = embedding_service.dimension
        
        results.append(result)
    
    return TextAnalysisResponse(results=results)


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8001)