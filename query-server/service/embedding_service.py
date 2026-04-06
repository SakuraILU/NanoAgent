"""
Embedding 服务层
"""
from typing import List
from impl import EmbeddingImpl


class EmbeddingService:
    """Embedding 向量化服务"""
    
    def __init__(self):
        self._impl = EmbeddingImpl()
    
    @property
    def dimension(self) -> int:
        """向量维度"""
        return self._impl.dimension
    
    def encode(self, texts: List[str], normalize: bool = True) -> List[List[float]]:
        """获取文本的 embedding 向量"""
        return self._impl.encode(texts, normalize)
    
    def similarity(self, text1: str, text2: str) -> float:
        """计算两个文本的余弦相似度"""
        return self._impl.similarity(text1, text2)