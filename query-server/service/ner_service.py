"""
NER 服务层
"""
from typing import List, Dict
from impl import NERImpl


class NERService:
    """命名实体识别服务"""
    
    def __init__(self, lang: str = "zh"):
        self._impl = NERImpl(lang)
    
    def recognize(self, texts: List[str]) -> List[Dict]:
        """
        实体识别
        
        Args:
            texts: 文本列表
            
        Returns:
            实体识别结果
        """
        results = []
        for text in texts:
            entities = self._impl.recognize(text)
            results.append({
                "text": text,
                "entities": entities
            })
        return results