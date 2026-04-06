"""
Tokenizer 服务层
"""
from typing import List, Dict
from impl import TokenizerImpl


class TokenizerService:
    """文本分词服务"""
    
    def __init__(self):
        self._impl = TokenizerImpl()
    
    def segment(self, texts: List[str], size: int = 10) -> List[Dict]:
        """
        分词（基于 TF 词频）
        
        对每个文本分词后，统计词频并筛选重要词，适用于缩短长 query 场景。
        
        Args:
            texts: 文本列表
            size: 每个文本返回的重要词数量，默认 10
            
        Returns:
            每个文本的重要词结果
        """
        results = []
        for text in texts:
            words = self._impl.segment(text, size)
            results.append({
                "text": text,
                "words": words,
                "count": len(words)
            })
        return results