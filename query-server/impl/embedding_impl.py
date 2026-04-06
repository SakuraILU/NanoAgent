"""
向量化实现类
"""
from sentence_transformers import SentenceTransformer
from typing import List
import os


class EmbeddingImpl:
    """Embedding 向量化实现"""
    
    def __init__(self, model_path: str = None):
        if model_path is None:
            model_path = os.path.join(
                os.path.dirname(os.path.dirname(__file__)), 
                "models", 
                "bge-small-zh-v1.5"
            )
        self.model = SentenceTransformer(model_path)
        self.dimension = 512  # bge-small-zh-v1.5 的维度
    
    def encode(self, texts: List[str], normalize: bool = True) -> List[List[float]]:
        """获取文本的 embedding 向量"""
        embeddings = self.model.encode(
            texts,
            normalize_embeddings=normalize,
            convert_to_numpy=True
        )
        return embeddings.tolist()
    
    def similarity(self, text1: str, text2: str) -> float:
        """计算两个文本的余弦相似度"""
        embeddings = self.model.encode(
            [text1, text2],
            normalize_embeddings=True,
            convert_to_numpy=True
        )
        return float(embeddings[0] @ embeddings[1])