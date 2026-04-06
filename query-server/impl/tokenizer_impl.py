"""
文本分词实现类
"""
import jieba
from typing import List, Dict
from collections import Counter
import re


class TokenizerImpl:
    """文本分词实现"""
    
    # 默认停用词（可根据需要扩展）
    DEFAULT_STOP_WORDS = {
        "的", "了", "是", "在", "我", "有", "和", "就", "不", "人", "都", "一", "一个",
        "上", "也", "很", "到", "说", "要", "去", "你", "会", "着", "没有", "看", "好",
        "自己", "这", "那", "什么", "他", "她", "它", "们", "这个", "那个", "怎么",
        "吗", "呢", "啊", "吧", "呀", "哦", "嗯", "哈", "嘿", "哎", "唉"
    }
    
    def __init__(self, stop_words: set = None):
        self.stop_words = stop_words or self.DEFAULT_STOP_WORDS
    
    def _is_valid_word(self, word: str) -> bool:
        """判断是否是有效词（非停用词、非纯数字、非纯标点）"""
        if not word:
            return False
        if word in self.stop_words:
            return False
        if word.isdigit():
            return False
        if re.match(r'^[^\w\u4e00-\u9fff]+$', word):  # 纯标点符号
            return False
        return True
    
    def segment(self, text: str, top_k: int = 10) -> List[Dict]:
        """
        分词（基于 TF 词频）
        
        对单个文本分词后，统计词频并筛选重要词，适用于缩短长 query 场景。
        
        Args:
            text: 文本
            size: 返回的重要词数量，默认 10
            
        Returns:
            重要词列表，按词频降序排列
            [{"word": "xxx", "count": 3}, ...]
        """
        word_list = [w for w in jieba.cut(text) if self._is_valid_word(w)]
        
        if not word_list:
            return []
        
        # 统计词频
        word_counts = Counter(word_list)
        
        # 按词频排序，取前 size 个
        top_words = word_counts.most_common(top_k)
        
        return [{"word": word, "count": count} for word, count in top_words]