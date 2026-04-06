"""
命名实体识别实现类
"""
from typing import List, Dict


class NERImpl:
    """命名实体识别实现（基于 stanza）"""
    
    def __init__(self, lang: str = "zh"):
        """
        初始化 NER 服务（完全离线模式）
        
        Args:
            lang: 语言代码，默认中文 "zh"
        """
        import stanza
        self.nlp = stanza.Pipeline(
            lang, 
            processors="tokenize,ner",
            download_method=None,  # 完全离线，不检查更新
            verbose=False
        )
    
    def recognize(self, text: str) -> List[Dict]:
        """
        实体识别
        
        Args:
            text: 文本
            
        Returns:
            实体识别结果
        """
        doc = self.nlp(text)
        entities = []
        for ent in doc.ents:
            entities.append({
                "text": ent.text,
                "type": ent.type,
                "start": ent.start_char,
                "end": ent.end_char
            })
        return entities