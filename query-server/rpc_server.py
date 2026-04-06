"""
Thrift RPC 服务入口
"""
import sys
import os

# 添加 gen 目录到路径
gen_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "gen")
sys.path.insert(0, gen_path)

from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol
from thrift.server import TServer

from query_server import QueryServer
from query_server.ttypes import Word, SegmentResult, Entity, NERResult, AnalysisResult

from service import EmbeddingService, TokenizerService, NERService


class QueryServerHandler:
    """QueryServer Thrift 处理器"""
    
    def __init__(self):
        self.embedding_service = EmbeddingService()
        self.tokenizer_service = TokenizerService()
        self._ner_service = None
    
    @property
    def ner_service(self):
        """延迟加载 NER 服务"""
        if self._ner_service is None:
            self._ner_service = NERService()
        return self._ner_service
    
    def health(self):
        """健康检查"""
        return "ok"
    
    def segment(self, texts, size):
        """分词（提取重要词）"""
        results = self.tokenizer_service.segment(texts, size=size)
        return [
            SegmentResult(
                text=r["text"],
                words=[Word(word=w["word"], count=w["count"]) for w in r["words"]],
                count=r["count"]
            )
            for r in results
        ]
    
    def ner(self, texts):
        """NER 实体识别"""
        results = self.ner_service.recognize(texts)
        return [
            NERResult(
                text=r["text"],
                entities=[
                    Entity(
                        text=e["text"],
                        type=e["type"],
                        start_pos=e["start"],
                        end_pos=e["end"]
                    )
                    for e in r["entities"]
                ]
            )
            for r in results
        ]
    
    def embed(self, texts, normalize):
        """向量化"""
        return self.embedding_service.encode(texts, normalize)
    
    def similarity(self, text1, text2):
        """文本相似度"""
        return self.embedding_service.similarity(text1, text2)
    
    def analyze(self, texts, keywords, ner, embedding, keyword_top_k):
        """综合分析"""
        results = []
        
        for text in texts:
            result = AnalysisResult(text=text)
            
            if keywords:
                kw_result = self.tokenizer_service.segment([text], size=keyword_top_k)
                result.keywords = [
                    Word(word=w["word"], count=w["count"]) 
                    for w in kw_result[0]["words"]
                ]
            
            if ner:
                ner_result = self.ner_service.recognize([text])
                result.entities = [
                    Entity(
                        text=e["text"],
                        type=e["type"],
                        start_pos=e["start"],
                        end_pos=e["end"]
                    )
                    for e in ner_result[0]["entities"]
                ]
            
            if embedding:
                emb = self.embedding_service.encode([text], normalize=True)
                result.embedding = emb[0]
                result.dimension = self.embedding_service.dimension
            
            results.append(result)
        
        return results


def main():
    handler = QueryServerHandler()
    processor = QueryServer.Processor(handler)
    transport = TSocket.TServerSocket(host="0.0.0.0", port=9090)
    tfactory = TTransport.TBufferedTransportFactory()
    pfactory = TBinaryProtocol.TBinaryProtocolFactory()
    
    server = TServer.TThreadedServer(processor, transport, tfactory, pfactory)
    
    print("Thrift Server started on port 9090")
    server.serve()


if __name__ == "__main__":
    main()